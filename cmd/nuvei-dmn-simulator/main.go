package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := newRootCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "nuvei-dmn-simulator",
		Short: "Generate and send signed Nuvei DMN payloads for development and QA",
		Long: `nuvei-dmn-simulator generates signed Nuvei Direct Merchant Notification payloads.

The simulator is intended for local, staging, and explicitly trusted test targets.
It must not be used to send unsigned DMNs or to bypass merchant webhook security.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newPreviewCommand())
	rootCmd.AddCommand(newSendCommand())
	rootCmd.AddCommand(newServerCommand())

	return rootCmd
}
