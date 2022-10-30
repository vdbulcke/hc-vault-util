package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	// bind to root command
	rootCmd.AddCommand(transitCmd)

}

var transitCmd = &cobra.Command{
	Use:   "transit",
	Short: "Commands for transit Vault backend",
	// Long: "",
	Run: func(cmd *cobra.Command, args []string) {

		// command does nothing
		err := cmd.Help()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(1)
	},
}
