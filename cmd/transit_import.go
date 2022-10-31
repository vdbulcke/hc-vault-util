package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/logger"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/transit"
)

// args var
var privKey string
var transitMount string
var transitKey string

func init() {
	// bind to root command
	transitCmd.AddCommand(importCmd)
	// add flags to sub command
	importCmd.Flags().StringVarP(&privKey, "pkcs8-pem-key", "k", "", "The private key to import")
	importCmd.Flags().StringVarP(&transitKey, "transit-key", "t", "", "The name of the transit key to import")
	importCmd.Flags().StringVarP(&transitMount, "mount", "", "transit", "Mount path of transit backend")

	// required flags
	//nolint
	importCmd.MarkFlagRequired("pkcs8-pem-key")
	//nolint
	importCmd.MarkFlagRequired("transit-key")

}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports private key into transit backend",
	Long:  "Imports private PKCS8 PEM private key into transit backend",
	Run:   importRun,

	Example: `
   hc-vault-util transit import --pkcs8-pem-key ./key.pem --transit-key rsa-key 

Mandatory Environment Variables:
- VAULT_ADDR: Address of the vault server 
- VAULT_TOKEN: Vault authentication token. With permission to read transit/wrapping_key and write transit/keys/[KEY-NAME]/import.

Optional Environment Variables:
- VAULT_CACERT: Path to a PEM encoded CA file to verify TLS on the VAULT_ADDR.
- VAULT_CAPATH: Path to a directory of PEM encoded CA files to verify TLS on the VAULT_ADDR.
- VAULT_SKIP_VERIFY: To disable TLS verification completely.

Docs: 
- https://developer.hashicorp.com/vault/docs/secrets/transit/key-wrapping-guide#software-example-go
`,
}

// importRun cobra server handler
func importRun(cmd *cobra.Command, args []string) {

	logger := logger.GenLogger(Debug, noColor)

	transitClient, err := transit.NewTransitClient(logger)
	if err != nil {
		logger.Error("Error creating transit client", "error", err)
		os.Exit(1)
	}

	// set key properties
	transitClient.SetKeyProperties(transitMount, transitKey)

	err = transitClient.ImportPrivateKey(privKey)
	if err != nil {
		logger.Error("Error importing key", "error", err)
		os.Exit(1)
	}
	apiPath := fmt.Sprintf("%s/keys/%s", transitMount, transitKey)
	logger.Info("Import successful", "path", apiPath)
}
