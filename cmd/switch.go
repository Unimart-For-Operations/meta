package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/idpbuilder/meta/internal/host"
	"github.com/idpbuilder/meta/internal/platform"
	"github.com/spf13/cobra"
)

var switchHost string

var switchCmd = &cobra.Command{
	Use:   "switch [host]",
	Short: "Apply Nix configuration for the current host",
	Long: `Apply the Nix/Home Manager configuration for this machine.

By default, auto-detects the current host by matching the system username
against cmdr/home/02-hosts/*/meta.nix files. Override with --host or by
passing a host name as an argument.

On macOS, runs darwin-rebuild switch. On Linux, runs home-manager switch.
First-time macOS runs will bootstrap nix-darwin automatically.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func init() {
	switchCmd.Flags().StringVar(&switchHost, "host", "", "host name to apply (overrides auto-detection)")
	deliCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}
	cmdrDir := filepath.Join(dir, "cmdr")

	// Resolve host name: positional arg > --host flag > auto-detect
	hostName := ""
	if len(args) > 0 {
		hostName = args[0]
	} else if switchHost != "" {
		hostName = switchHost
	} else {
		info, err := host.Detect(dir)
		if err != nil {
			return fmt.Errorf("host auto-detection failed: %w\n\nUse: unimart deli switch <host-name>\nOr:  unimart deli switch --host <host-name>\nRun: unimart deli hosts   to list available hosts", err)
		}
		hostName = info.Name
	}

	fmt.Printf("%s Applying configuration for: %s\n\n", bold("▸"), cyan(hostName))

	flakeRef := fmt.Sprintf(".#%s", hostName)

	if platform.IsDarwin() {
		return switchDarwin(cmdrDir, flakeRef)
	}
	return switchLinux(cmdrDir, flakeRef)
}

func switchDarwin(cmdrDir, flakeRef string) error {
	if platform.CommandExists("darwin-rebuild") {
		fmt.Printf("  Running %s...\n", bold("darwin-rebuild switch"))
		if err := platform.RunVisibleDir(cmdrDir, "sudo", "darwin-rebuild", "switch", "--flake", flakeRef); err != nil {
			return fmt.Errorf("darwin-rebuild switch failed: %w", err)
		}
	} else {
		fmt.Printf("  %s darwin-rebuild not found — running first-time bootstrap...\n", warn("[warn]"))

		// Move /etc/bashrc if it exists (nix-darwin requirement)
		if _, err := os.Stat("/etc/bashrc"); err == nil {
			fmt.Println("  Moving /etc/bashrc to /etc/bashrc.before-nix-darwin...")
			if err := platform.RunVisible("sudo", "mv", "/etc/bashrc", "/etc/bashrc.before-nix-darwin"); err != nil {
				return fmt.Errorf("move /etc/bashrc: %w", err)
			}
		}

		fmt.Printf("  Running %s (first-time bootstrap)...\n", bold("nix-darwin"))
		if err := platform.RunVisibleDir(cmdrDir, "sudo", "nix", "run", "nix-darwin/master#darwin-rebuild", "--", "switch", "--flake", flakeRef); err != nil {
			return fmt.Errorf("nix-darwin bootstrap failed: %w", err)
		}
	}

	fmt.Printf("\n%s Configuration applied\n", pass("[pass]"))
	return nil
}

func switchLinux(cmdrDir, flakeRef string) error {
	fmt.Printf("  Running %s...\n", bold("home-manager switch"))
	if err := platform.RunVisibleDir(cmdrDir, "home-manager", "switch", "--flake", flakeRef); err != nil {
		return fmt.Errorf("home-manager switch failed: %w", err)
	}

	fmt.Printf("\n%s Configuration applied\n", pass("[pass]"))
	return nil
}
