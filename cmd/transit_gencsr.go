package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/logger"
	"github.com/vdbulcke/hc-vault-util/hc-vault-util/transit"
)

var cfsslCSRFile string
var keyVersion int

func init() {
	// bind to root command
	transitCmd.AddCommand(genCSRCmd)
	// add flags to sub command
	genCSRCmd.Flags().StringVarP(&cfsslCSRFile, "csr-json", "c", "", "The path to a cfssl csr file")
	genCSRCmd.Flags().StringVarP(&transitKey, "transit-key", "t", "", "The name of the transit key to import")
	genCSRCmd.Flags().StringVarP(&transitMount, "mount", "", "transit", "Mount path of transit backend")
	genCSRCmd.Flags().IntVarP(&keyVersion, "version", "", 0, "Version of the transit key, or 0 for latest (default 0)")

	// required flags
	//nolint
	genCSRCmd.MarkFlagRequired("transit-key")
	//nolint
	genCSRCmd.MarkFlagRequired("csr-json")

}

var genCSRCmd = &cobra.Command{
	Use:   "gencsr",
	Short: "Generate a CSR from private key in transit backend",
	Long:  "Generate a CSR from private key in transit backend",
	Run:   genCSRRun,

	Example: `
   hc-vault-util transit gencsr --csr-json example/csr.json --transit-key "rsa" 

Mandatory Environment Variables:
- VAULT_ADDR: Address of the vault server 
- VAULT_TOKEN: Vault authentication token. With permission to read 'transit/keys/[KEY-NAME]' and write 'transit/sign/[KEY-NAME]'.

Optional Environment Variables:
- VAULT_CACERT: Path to a PEM encoded CA file to verify TLS on the VAULT_ADDR.
- VAULT_CAPATH: Path to a directory of PEM encoded CA files to verify TLS on the VAULT_ADDR.
- VAULT_SKIP_VERIFY: To disable TLS verification completely.


CSR JSON format: 
{
	"CN": "Foo",
    "hosts": [
        "cloudflare.com",
        "www.cloudflare.com"
    ],
    "names": [
        {
            "C": "US",
            "L": "San Francisco",
            "O": "CloudFlare",
            "OU": "Systems Engineering",
            "ST": "California"
        }
    ]
}
 

`,
}

// importRun cobra server handler
func genCSRRun(cmd *cobra.Command, args []string) {

	logger := logger.GenLogger(Debug, noColor)

	transitClient, err := transit.NewTransitClient(logger)
	if err != nil {
		logger.Error("Error creating transit client", "error", err)
		os.Exit(1)
	}

	// set key properties
	transitClient.SetKeyProperties(transitMount, transitKey)

	err = transitClient.GenCSR(cfsslCSRFile, keyVersion)
	if err != nil {
		logger.Error("Error generating CSR", "error", err)
		os.Exit(1)
	}

}
