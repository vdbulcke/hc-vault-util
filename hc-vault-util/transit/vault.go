package transit

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	vault "github.com/hashicorp/vault/api"
)

// NewVaultClient returns a vault api client configured based on the
// Vault standard env variables VAULT_ADDR, VAULT_TOKEN, etc
func NewVaultClient(logger hclog.Logger) (*vault.Client, error) {

	VAULT_ADDR := os.Getenv("VAULT_ADDR")
	if VAULT_ADDR == "" {
		return nil, fmt.Errorf("VAULT_ADDR no set")
	}

	VAULT_TOKEN := os.Getenv("VAULT_TOKEN")
	if VAULT_TOKEN == "" {
		return nil, fmt.Errorf("VAULT_TOKEN no set")
	}

	config := vault.DefaultConfig()

	config.Address = VAULT_ADDR

	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}

	// Authenticate
	client.SetToken(VAULT_TOKEN)

	// tls
	tlsConfig := &vault.TLSConfig{}

	VAULT_CACERT := os.Getenv("VAULT_CACERT")
	if VAULT_CACERT != "" {
		tlsConfig.CACert = VAULT_CACERT
	}

	VAULT_CAPATH := os.Getenv("VAULT_CAPATH")
	if VAULT_CAPATH != "" {
		tlsConfig.CAPath = VAULT_CAPATH
	}

	VAULT_SKIP_VERIFY := os.Getenv("VAULT_SKIP_VERIFY")
	if VAULT_SKIP_VERIFY != "" {
		logger.Warn("VAULT_SKIP_VERIFY is enabled ")
		tlsConfig.Insecure = true
	}

	err = config.ConfigureTLS(tlsConfig)
	if err != nil {
		return nil, err
	}

	return vault.NewClient(config)
}
