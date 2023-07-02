package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/logger"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/transit"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/tui"
)

// args var
var kvbMount string

func init() {
	// bind to root command
	rootCmd.AddCommand(uiCmd)
	// add flags to sub command
	uiCmd.Flags().StringVarP(&kvbMount, "kv2-mount", "", "secret", "Mount path of kv2 backend")

}

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "ui for kv2 backend",
	// Long:  "uis private PKCS8 PEM private key into transit backend",
	Run: uiRun,

	Example: `
   hc-vault-util ui  

Mandatory Environment Variables:
- VAULT_ADDR: Address of the vault server 
- VAULT_TOKEN: Vault authentication token. 

Optional Environment Variables:
- VAULT_CACERT: Path to a PEM encoded CA file to verify TLS on the VAULT_ADDR.
- VAULT_CAPATH: Path to a directory of PEM encoded CA files to verify TLS on the VAULT_ADDR.
- VAULT_SKIP_VERIFY: To disable TLS verification completely.

`,
}

// uiRun cobra server handler
func uiRun(cmd *cobra.Command, args []string) {

	logger := logger.GenLogger(Debug, noColor)

	client, err := transit.NewVaultClient(logger)
	if err != nil {
		logger.Error("Error creating vault client", "error", err)
		os.Exit(1)
	}

	err = tui.StartUI(client, kvbMount)
	if err != nil {
		logger.Error("Error starting TUI", "error", err)
		os.Exit(1)
	}
}
