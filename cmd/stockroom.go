package cmd

import (
	"github.com/spf13/cobra"
)

var stockroomCmd = &cobra.Command{
	Use:   "stockroom",
	Short: "Cross-repo coordination (submodules, drift, updates)",
	Long: `The stockroom — back-of-house inventory management.

Manages cross-repo operations: check submodule drift, pull updates,
and coordinate releases across the organization.`,
}

func init() {
	rootCmd.AddCommand(stockroomCmd)
}
