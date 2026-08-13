package cmd

import (
	"fmt"

	"github.com/Unimart-For-Operations/meta/internal/colima"
	"github.com/Unimart-For-Operations/meta/internal/prereqs"
	"github.com/spf13/cobra"
)

var (
	downStopColima bool
	downYes        bool
	downName       string
)

var freezerDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Tear down the IDP platform",
	Long: `Deletes the idpbuilder Kind cluster.
On macOS, optionally stops Colima with --stop-colima.`,
	RunE: runFreezerDown,
}

func init() {
	if prereqs.IsDarwin() {
		freezerDownCmd.Flags().BoolVar(&downStopColima, "stop-colima", false, "Also stop the Colima VM (macOS only)")
	}
	freezerDownCmd.Flags().BoolVarP(&downYes, "yes", "y", false, "Skip confirmation prompts")
	freezerDownCmd.Flags().StringVar(&downName, "name", "localdev", "Name of the Kind cluster to delete")
	freezerCmd.AddCommand(freezerDownCmd)
}

func runFreezerDown(cmd *cobra.Command, args []string) error {
	if !downYes {
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

		// macOS: two-step process (delete cluster, optionally stop Colima)
		fmt.Printf("%s Deleting IDP platform\n\n", bold("[1/2]"))
		if err := deleteIDP(cmd.Context(), downName); err != nil {
			return err
		}

		if downStopColima {
			fmt.Printf("\n%s Stopping Colima\n\n", bold("[2/2]"))
			if err := colima.Stop(); err != nil {
				return err
			}
		} else {
			fmt.Printf("\n%s Colima still running (use --stop-colima to also stop)\n", bold("[2/2]"))
		}
	} else {
		// Linux: single step (delete cluster)
		fmt.Printf("%s Deleting IDP platform\n\n", bold("[1/1]"))
		if err := deleteIDP(cmd.Context(), downName); err != nil {
			return err
		}
	}

	fmt.Printf("\n%s IDP platform torn down\n", pass("done"))
	return nil
}
