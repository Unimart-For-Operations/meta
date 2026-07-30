package cmd

import (
	"github.com/spf13/cobra"
)

var deliCmd = &cobra.Command{
	Use:   "deli",
	Short: "Workstation configuration (Nix/Home Manager)",
	Long: `The deli counter — custom-sliced workstation configuration.

Manages Nix/Home Manager host configs: switch profiles, check prerequisites,
bootstrap new machines, and list available host configurations.`,
}

func init() {
	rootCmd.AddCommand(deliCmd)
}
