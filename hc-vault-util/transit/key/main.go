package key

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strconv"

	vault "github.com/hashicorp/vault/api"
	"github.com/savaki/jq"
	"go.uber.org/zap"
)

var (
	cryptoHashToVaultHash = map[crypto.Hash]string{
		crypto.SHA1:     "sha1",
		crypto.SHA224:   "sha2-224",
		crypto.SHA256:   "sha2-256",
		crypto.SHA384:   "sha2-384",
		crypto.SHA512:   "sha2-512",
		crypto.SHA3_224: "sha3-224",
		crypto.SHA3_256: "sha3-256",
		crypto.SHA3_384: "sha3-384",
		crypto.SHA3_512: "sha3-512",
	}
)

type TransitPublicKey struct {
	// pub key for JWKS
	PublicKey crypto.PublicKey
	// Version
	Version int

	// Name
	Name string
}

func NewTransitPublicKey(pub crypto.PublicKey, v int, name string) *TransitPublicKey {
	return &TransitPublicKey{
		PublicKey: pub,
		Version:   v,
		Name:      name,
	}
}

type VaultTransitKey struct {
	// transit backend mount
	MountPath string
	// transit Key Name
	Name string

	// 'key' type
	Type string

	// Version
	Version int

	// Min Version
	MinVersion int

	// Set sig version
	SigVersion int

	// List of public keys
	PublicKeys []*TransitPublicKey

	// the vault api client
	client *vault.Client
	// context for vault client
	ctx context.Context

	// logger
	logger *zap.Logger
}

func NewVaultTransitKey(ctx context.Context, l *zap.Logger, client *vault.Client, mount string, name string) (*VaultTransitKey, error) {

	// instantiate Transit vault key
	k := &VaultTransitKey{
		Name:      name,
		ctx:       ctx,
		client:    client,
		MountPath: mount,
		logger:    l,
	}

	return k, nil

}

// SyncKeyInfo read transit key info
func (k *VaultTransitKey) SyncKeyInfo() error {

	// transit key api path
	keyPath := fmt.Sprintf("%s/keys/%s", k.MountPath, k.Name)
	// read transit key
	keyInfo, err := k.client.Logical().ReadWithContext(k.ctx, keyPath)
	if err != nil {
		return err
	}

	k.logger.Debug("transit read key response", zap.Any("resp", keyInfo))

	if keyInfo == nil {
		k.logger.Error("No response for transit key read", zap.String("key", keyPath))
		return fmt.Errorf("error reading key %s", keyPath)
	}

	// parse key type
	keyType, ok := keyInfo.Data["type"].(string)
	if !ok {
		k.logger.Debug("Key type not found in transit read response", zap.Any("resp", keyInfo))
		return fmt.Errorf("key type not found for %s", keyPath)
	}

	keyVersionJson, ok := keyInfo.Data["latest_version"].(json.Number)
	if !ok {
		k.logger.Debug("Key latest_version not found in transit read response", zap.Any("resp", keyInfo))
		return fmt.Errorf("key latest_version not found for %s", keyPath)
	}

	keyVersion, err := keyVersionJson.Int64()
	if err != nil {
		return err
	}

	minVersionJson, ok := keyInfo.Data["min_decryption_version"].(json.Number)
	if !ok {
		k.logger.Debug("Key min_decryption_version not found in transit read response", zap.Any("resp", keyInfo))
		return fmt.Errorf("key min_decryption_version not found for %s", keyPath)
	}

	minVersion, err := minVersionJson.Int64()
	if err != nil {
		return err
	}

	pubKeys := []*TransitPublicKey{}

	// for each pub keys within range min version to latest_version
	for i := int(minVersion); i <= int(keyVersion); i++ {

		pub, err := k.GetPublicKeyFromTransitResponse(keyInfo, i)
		if err != nil {
			k.logger.Error("error parsing pub key from response", zap.Error(err))
			return err
		}

		pubKeys = append(pubKeys, NewTransitPublicKey(pub, i, k.Name))

	}

	k.Type = keyType
	k.Version = int(keyVersion)
	k.SigVersion = int(keyVersion)
	k.PublicKeys = pubKeys
	k.MinVersion = int(minVersion)

	return nil

}

// Sign byte payload, and returns "signature" output of transit sign api
func (k *VaultTransitKey) Sign(inputBytes []byte, apiSigAlg string, apiHashAlg string, marshallingAlg string, prehashed bool) (string, error) {

	args := map[string]interface{}{
		// transit required input to base64 encoded
		"input":                base64.StdEncoding.EncodeToString(inputBytes),
		"signature_algorithm":  apiSigAlg,
		"marshaling_algorithm": marshallingAlg,
		"prehashed":            prehashed,
		"key_version":          k.Version,
	}

	// sign with transit API
	signingPath := fmt.Sprintf("%s/sign/%s/%s", k.MountPath, k.Name, apiHashAlg)
	transitResp, err := k.client.Logical().WriteWithContext(k.ctx, signingPath, args)
	if err != nil {
		return "", err
	}

	if transitResp == nil {
		k.logger.Error("No response for transit signing ", zap.String("key", signingPath))
		return "", fmt.Errorf("error signing key %s", signingPath)
	}

	sig, ok := transitResp.Data["signature"].(string)
	if !ok {
		return "", fmt.Errorf("unable to get 'signature' from transit response")
	}

	return sig, nil
}

// verify byte payload,  and signature (without the "vault:v1")
//
//	returns true if signature is valid for byte payload
func (k *VaultTransitKey) Verify(inputBytes []byte, signature string, apiSigAlg string, apiHashAlg string, marshallingAlg string, prehashed bool) (bool, error) {

	args := map[string]interface{}{
		// transit required input to base64 encoded
		"input":                base64.StdEncoding.EncodeToString(inputBytes),
		"signature":            fmt.Sprintf("vault:v%d:%s", k.SigVersion, signature),
		"signature_algorithm":  apiSigAlg,
		"marshaling_algorithm": marshallingAlg,
		"prehashed":            prehashed,
	}

	// sign with transit API
	signingPath := fmt.Sprintf("%s/verify/%s/%s", k.MountPath, k.Name, apiHashAlg)
	transitResp, err := k.client.Logical().WriteWithContext(k.ctx, signingPath, args)
	if err != nil {
		return false, err
	}

	if transitResp == nil {
		k.logger.Error("No response for transit verifying ", zap.String("key", signingPath))
		return false, fmt.Errorf("error verifying key %s", signingPath)
	}

	sigValid, ok := transitResp.Data["valid"].(bool)
	if !ok {
		return false, fmt.Errorf("unable to get 'valid' from transit response")
	}

	return sigValid, nil
}

func (k *VaultTransitKey) SetSigKeyVersion(v int) {
	k.SigVersion = v
}

// GetPublicKeyFromTransitResponse return parsed public key from the keyInfo transit read API response
func (k *VaultTransitKey) GetPublicKeyFromTransitResponse(keyInfo *vault.Secret, version int) (crypto.PublicKey, error) {

	// Build jq query
	jqQuery := fmt.Sprintf(".keys.%d.public_key", version)

	// extract and parse public key PEM
	op, err := jq.Parse(jqQuery)
	if err != nil {
		k.logger.Debug("jq query", zap.String("query", jqQuery))
		return nil, err
	}
	data, err := json.Marshal(keyInfo.Data)
	if err != nil {
		return nil, err
	}
	value, err := op.Apply(data)
	if err != nil {
		return nil, err
	}
	key, err := strconv.Unquote(string(value))
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(key))

	if block == nil {
		return nil, fmt.Errorf("error Pem Decoding pub key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub, nil
}
