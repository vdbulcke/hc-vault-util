package transit

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/google/tink/go/kwp/subtle"
	"github.com/hashicorp/go-hclog"
	vault "github.com/hashicorp/vault/api"
)

type TransitClient struct {
	logger hclog.Logger

	client *vault.Client

	ctx context.Context
}

func NewTransitClient(l hclog.Logger) (*TransitClient, error) {

	ctx := context.Background()

	client, err := NewVaultClient(l)
	if err != nil {
		return nil, err
	}

	return &TransitClient{
		logger: l,
		client: client,
		ctx:    ctx,
	}, nil

}

func (t *TransitClient) ImportPrivateKey(transitMount, keyName, keyFile string) error {
	// read private key file
	data, err := os.ReadFile(keyFile)
	if err != nil {
		t.logger.Error("Error reading private key", "error", err)
		return err
	}

	key, _ := pem.Decode(data)
	if key == nil {
		return fmt.Errorf("Error Decoding PEM file, invalid format")
	}
	privKey, err := x509.ParsePKCS8PrivateKey(key.Bytes)
	if err != nil {
		t.logger.Error("Error parsing PEM PKCS8 private key", "error", err)
		t.logger.Warn("You can use openssl to convert your key into PKCS8 PEM format:\n\n  openssl pkcs8 -topk8 -outform DER -in key.pem -out key_pk8.pem -nocrypt \n")
		return err
	}

	// case key to derive hc vault key type
	var keyType string
	switch priv := privKey.(type) {
	case *rsa.PrivateKey:

		// *8 to convert bytes to bits
		bitsSize := priv.Size() * 8
		keyType = fmt.Sprintf("rsa-%d", bitsSize)

	case *ecdsa.PrivateKey:

		keyType = fmt.Sprintf("ecdsa-p%d", priv.Params().BitSize)
	case *ed25519.PrivateKey:

		keyType = "ed25519"
	default:
		supportedType := []string{"rsa-2048", "rsa-3072", "rsa-4096", "ecdsa-p256", "ecdsa-p384", "ecdsa-p521", "ed25519"}
		t.logger.Error("Unsupported key type", "format", supportedType)
		return fmt.Errorf("Unsupported key type")

	}

	t.logger.Info("Importing key type", "type", keyType)

	// implementing steps from doc
	// https://developer.hashicorp.com/vault/docs/secrets/transit/key-wrapping-guide#software-example-go
	t.logger.Debug("generating wrapping key from transit...")
	wrappingKey, err := t.getWrappingKey(transitMount)
	if err != nil {
		t.logger.Error("error generating wrapping key", "error", err)
		return err
	}

	t.logger.Debug("generating AES ephemeral key...")
	ephemeralAESKey, err := t.genAESKey()
	if err != nil {
		t.logger.Error("error generating AES ephemeral key", "error", err)
		return err
	}

	t.logger.Debug("generating key wrap from AES key")
	wrapKWP, err := subtle.NewKWP(ephemeralAESKey)
	if err != nil {
		t.logger.Error("error generating key Wrap", "error", err)
		return err
	}

	t.logger.Debug("Wrapping private key...")
	wrappedTargetKey, err := wrapKWP.Wrap(key.Bytes)
	if err != nil {
		t.logger.Error("error wrapping private key", "error", err)
		return err
	}

	//
	t.logger.Debug("encrypting AES ephemeral key with RSA wrapping key")
	wrappedAESKey, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		wrappingKey,
		ephemeralAESKey,
		[]byte{},
	)
	if err != nil {
		t.logger.Error("encrypting AES ephemeral key", "error", err)
		return err
	}

	// combined the payload and base64 encode
	combinedCiphertext := append(wrappedAESKey, wrappedTargetKey...)
	base64Ciphertext := base64.StdEncoding.EncodeToString(combinedCiphertext)

	// import into transit backend
	return t.TransitImportKey(transitMount, keyName, "SHA256", base64Ciphertext, keyType)

}

func (t *TransitClient) TransitImportKey(transitMount, keyName, hashFunc, base64Ciphertext, keyType string) error {

	args := map[string]interface{}{
		// transit required input to base64 encoded
		"ciphertext":     base64Ciphertext,
		"hash_function ": hashFunc,
		"type":           keyType,
		"exportable":     false,
	}

	apiPath := fmt.Sprintf("%s/keys/%s/import", transitMount, keyName)
	_, err := t.client.Logical().WriteWithContext(t.ctx, apiPath, args)
	if err != nil {
		return err
	}

	return nil

}

func (t *TransitClient) getWrappingKey(transitMount string) (*rsa.PublicKey, error) {

	// transit key api path
	apiPath := fmt.Sprintf("%s/wrapping_key", transitMount)
	// read transit key
	keyInfo, err := t.client.Logical().ReadWithContext(t.ctx, apiPath)
	if err != nil {
		return nil, err
	}

	// parse key type
	publicKeyPem, ok := keyInfo.Data["public_key"].(string)
	if !ok {
		return nil, fmt.Errorf("error parsing wrapping key %s", apiPath)
	}

	t.logger.Debug("go wrapping key", "public_key", publicKeyPem)

	block, _ := pem.Decode([]byte(publicKeyPem))

	if block == nil {
		return nil, fmt.Errorf("error Pem Decoding pub key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// cast is as a RSA
	var wrappingKey *rsa.PublicKey
	switch pub := pub.(type) {
	case *rsa.PublicKey:

		wrappingKey = pub

	default:
		return nil, fmt.Errorf("unknown type of wrapping key for %s", apiPath)
	}

	return wrappingKey, nil

}

func (t *TransitClient) genAESKey() ([]byte, error) {
	ephemeralAESKey := make([]byte, 32)
	_, err := rand.Read(ephemeralAESKey)
	if err != nil {
		return nil, err
	}

	return ephemeralAESKey, nil
}
