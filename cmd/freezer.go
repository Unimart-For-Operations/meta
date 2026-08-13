package cmd

import "github.com/spf13/cobra"

var freezerCmd = &cobra.Command{
	Use:   "freezer",
	Short: "IDP platform lifecycle (clusters, repos)",
	Long: `The freezer section — spin up and cool down IDP clusters.

Manages the full IDP platform lifecycle: start/stop clusters, check status,
manage repos, and configure the environment.`,
}

func init() {
	rootCmd.AddCommand(freezerCmd)
}
