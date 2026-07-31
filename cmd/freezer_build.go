package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Unimart-For-Operations/meta/internal/builder"
	"github.com/spf13/cobra"
)

var freezerBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build idpbuilder from source",
	Long:  `Runs 'make build' in the idpbuilder repository directory.`,
	RunE:  runFreezerBuild,
}

func init() {
	freezerCmd.AddCommand(freezerBuildCmd)
}

func runFreezerBuild(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}
	idpDir := filepath.Join(orgDir, "idpbuilder")

	fmt.Printf("%s Building idpbuilder\n\n", bold(">>"))

	if err := builder.Build(idpDir, verbose); err != nil {
		return err
	}

	fmt.Printf("\n%s idpbuilder built successfully\n", pass("done"))
	return nil
}
