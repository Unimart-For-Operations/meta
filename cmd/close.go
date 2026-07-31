package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Unimart-For-Operations/meta/internal/builder"
	"github.com/Unimart-For-Operations/meta/internal/colima"
	"github.com/Unimart-For-Operations/meta/internal/prereqs"
	"github.com/spf13/cobra"
)

var (
	closeStopColima bool
	closeYes        bool
)

var closeCmd = &cobra.Command{
	Use:   "close",
	Short: "Close up shop — tear down the IDP platform",
	Long: `Symmetric inverse of "unimart open".

Deletes the idpbuilder Kind cluster and optionally stops the
Colima VM on macOS.

This is equivalent to "unimart freezer down" but lives at the
top level to pair with "unimart open".`,
	RunE: runClose,
}

func init() {
	if prereqs.IsDarwin() {
		closeCmd.Flags().BoolVar(&closeStopColima, "stop-colima", false, "Also stop the Colima VM (macOS only)")
	}
	closeCmd.Flags().BoolVarP(&closeYes, "yes", "y", false, "Skip confirmation prompt")
	rootCmd.AddCommand(closeCmd)
}

func runClose(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}
	idpDir := filepath.Join(orgDir, "idpbuilder")

	if !closeYes {
		if !promptYesNo("This will delete the IDP cluster. Continue? [Y/n] ", true) {
			fmt.Println("Cancelled")
			return nil
		}
	}

	if prereqs.IsDarwin() {
		// Ensure DOCKER_HOST points to Colima before any Docker/Kind operations
		if prereqs.CommandExists("colima") && colima.IsRunning() {
			if err := colima.EnsureDockerHost(); err != nil {
				return fmt.Errorf("could not set DOCKER_HOST: %w", err)
			}
		}

		fmt.Printf("%s Deleting IDP platform\n\n", bold("[1/2]"))
		if err := builder.Delete(idpDir); err != nil {
			return err
		}

		if closeStopColima {
			fmt.Printf("\n%s Stopping Colima\n\n", bold("[2/2]"))
			if err := colima.Stop(); err != nil {
				return err
			}
		} else {
			fmt.Printf("\n%s Colima still running (use --stop-colima to also stop)\n", bold("[2/2]"))
		}
	} else {
		fmt.Printf("%s Deleting IDP platform\n\n", bold("[1/1]"))
		if err := builder.Delete(idpDir); err != nil {
			return err
		}
	}

	fmt.Printf("\n%s unimart is closed for the day\n", pass("done"))
	return nil
}
