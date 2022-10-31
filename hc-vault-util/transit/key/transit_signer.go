package key

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap"
)

// TransitSigner  implement crypto.signer interface
// https://pkg.go.dev/crypto#Signer
type TransitSigner struct {
	Key *VaultTransitKey
	// signature_algorithm one of "pss" or "pkcs1v15"
	SigAlg string
}

// NewTransitSigner with transit key and signature_algorithm one of "pss" or "pkcs1v15"
func NewTransitSigner(k *VaultTransitKey, SigAlg string) *TransitSigner {

	return &TransitSigner{
		Key:    k,
		SigAlg: SigAlg,
	}
}

// Public returns the public key corresponding to the opaque,
// private key.
func (s *TransitSigner) Public() crypto.PublicKey {

	version := s.Key.Version

	var pub crypto.PublicKey
	for _, k := range s.Key.PublicKeys {

		if k.Version == version {
			pub = k.PublicKey
			break
		}
	}

	if pub == nil {
		// if cannot find version default to last public key
		pub = s.Key.PublicKeys[len(s.Key.PublicKeys)-1]
	}

	return pub
}

// Sign signs digest with the private key, possibly using entropy from
// rand. For an RSA key, the resulting signature should be either a
// PKCS #1 v1.5 or PSS signature (as indicated by opts). For an (EC)DSA
// key, it should be a DER-serialised, ASN.1 signature structure.
//
// Hash implements the SignerOpts interface and, in most cases, one can
// simply pass in the hash function used as opts. Sign may also attempt
// to type assert opts to other types in order to obtain algorithm
// specific values. See the documentation in each package for details.
//
// Note that when a signature of a hash of a larger message is needed,
// the caller is responsible for hashing the larger message and passing
// the hash (as digest) and the hash function (as opts) to Sign.
func (s *TransitSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {

	// convert hash
	hash := opts.HashFunc()
	hashAlg, ok := cryptoHashToVaultHash[hash]
	if !ok {
		return nil, fmt.Errorf("unsupported hash %s", hash.String())

	}

	s.Key.logger.Debug("Hash alg", zap.String("opts.HashFunc()", hash.String()), zap.String("ohashAlg", hashAlg))
	// sign with vault transit key
	sigVault, err := s.Key.Sign(digest, s.SigAlg, hashAlg, "asn1", true)
	if err != nil {
		return nil, err
	}
	// check format
	if !strings.HasPrefix(sigVault, "vault:v") {
		return nil, fmt.Errorf("invalid signature expecting prefix 'vault:v' but got %s", sigVault)
	}

	// Vault transit signature are prefixed with
	// 'vault:vX:' indicating the version of key used for this signature
	// split on ':' and return last part
	sigParts := strings.Split(sigVault, ":")

	// base64 decode signature
	sigBytes, err := base64.StdEncoding.DecodeString(sigParts[2])
	if err != nil {
		return nil, err
	}

	return sigBytes, nil

}
