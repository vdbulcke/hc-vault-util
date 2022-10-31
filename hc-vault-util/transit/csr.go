package transit

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/log"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/logger"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/transit/key"
)

func (t *TransitClient) GenCSR(cfsslCSRFile string, keyVersion int) error {
	// read private key file
	data, err := os.ReadFile(cfsslCSRFile)
	if err != nil {
		t.logger.Error("Error reading csr config file", "error", err)
		return err
	}

	// parse Cfssl CSR JSON format
	req := &csr.CertificateRequest{}
	err = json.Unmarshal(data, req)
	if err != nil {
		return err
	}

	// validation

	// get zap logger from hclog properties
	hasDebug := t.logger.IsDebug()
	hasNoColor := true
	zapLog := logger.GetZapLogger(hasDebug, hasNoColor)
	// quite cfssl logger
	log.Level = log.LevelCritical

	// create a new transit key
	k, err := key.NewVaultTransitKey(t.ctx, zapLog, t.client, t.transitMount, t.keyName)
	if err != nil {
		return err
	}

	// get latest key info
	// version public keys
	err = k.SyncKeyInfo()
	if err != nil {
		return err
	}

	// validate key version
	if keyVersion > 0 {

		if keyVersion < k.MinVersion || keyVersion > k.Version {
			t.logger.Error("invalid key version, must be within", "min_version", k.MinVersion, "max_version", k.Version)
			return fmt.Errorf("invalid key version %d", keyVersion)
		}

		k.Version = keyVersion
	}

	// set default signing alg pkcs1v15 for RSA
	vaultSigAlg := "pss"
	if strings.HasPrefix(k.Type, "rsa-") {
		vaultSigAlg = "pkcs1v15"
	}

	// create new Transit SIgner
	signer := key.NewTransitSigner(k, vaultSigAlg)

	// use cfssl to generate the CSR with the transit signer
	csrPem, err := csr.Generate(signer, req)
	if err != nil {
		return err
	}

	t.logger.Info("PEM encoded CSR")

	// print CSR
	fmt.Println(string(csrPem[:]))
	return nil

}
