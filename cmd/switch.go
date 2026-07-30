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
var switchHomeOnly bool

var switchCmd = &cobra.Command{
	Use:   "switch [host]",
	Short: "Apply Nix configuration for the current host",
	Long: `Apply the Nix/Home Manager configuration for this machine.

By default, auto-detects the current host by matching the system username
against cmdr/home/02-hosts/*/meta.nix files. Override with --host or by
passing a host name as an argument.

On macOS, runs darwin-rebuild switch. On Linux, runs home-manager switch.
On NixOS, runs nixos-rebuild switch.

Use --home-only to apply only Home Manager state (no system rebuild).
On macOS, --home-only is currently unsupported.
First-time macOS runs will bootstrap nix-darwin automatically.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func init() {
	switchCmd.Flags().StringVar(&switchHost, "host", "", "host name to apply (overrides auto-detection)")
	switchCmd.Flags().BoolVar(&switchHomeOnly, "home-only", false, "apply Home Manager only (no system rebuild)")
	deliCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}
	cmdrDir := filepath.Join(dir, "cmdr")

	// Resolve host: positional arg > --host flag > auto-detect
	info, err := resolveSwitchHost(dir, args, switchHost)
	if err != nil {
		return err
	}
	hostName := info.Name
	hostPlatform := info.Platform

	fmt.Printf("%s Applying configuration for: %s\n\n", bold("▸"), cyan(hostName))

	flakeRef := fmt.Sprintf(".#%s", hostName)
	mode, err := selectApplyMode(hostPlatform, switchHomeOnly)
	if err != nil {
		return err
	}

	switch mode {
	case applyModeDarwin:
		return switchDarwin(cmdrDir, flakeRef)
	case applyModeNixOS:
		return switchNixOS(cmdrDir, flakeRef)
	case applyModeLinux:
		return switchLinux(cmdrDir, flakeRef)
	case applyModeHomeOnly:
		return switchHomeOnlyApply(cmdrDir, hostName)
	default:
		return fmt.Errorf("unsupported apply mode: %s", mode)
	}
}

const (
	applyModeDarwin   = "darwin"
	applyModeNixOS    = "nixos"
	applyModeLinux    = "linux"
	applyModeHomeOnly = "home-only"
)

// Host lookup seams — swapped in tests.
var (
	getHostFn    = host.GetHost
	detectHostFn = host.Detect
)

// Exec seams — swapped in tests to record argv instead of running commands.
var (
	runVisibleDirFn = platform.RunVisibleDir
	runVisibleFn    = platform.RunVisible
	commandExistsFn = platform.CommandExists
)

// resolveSwitchHost resolves the target host with precedence:
// positional arg > --host flag > auto-detect.
func resolveSwitchHost(dir string, args []string, flagHost string) (*host.Info, error) {
	name := flagHost
	if len(args) > 0 {
		name = args[0]
	}

	if name != "" {
		info, err := getHostFn(dir, name)
		if err != nil {
			return nil, fmt.Errorf("host %q not found: %w\n\nRun: unimart deli hosts   to list available hosts", name, err)
		}
		return info, nil
	}

	info, err := detectHostFn(dir)
	if err != nil {
		return nil, fmt.Errorf("host auto-detection failed: %w\n\nUse: unimart deli switch <host-name>\nOr:  unimart deli switch --host <host-name>\nRun: unimart deli hosts   to list available hosts", err)
	}
	return info, nil
}

func selectApplyMode(hostPlatform string, homeOnly bool) (string, error) {
	if homeOnly {
		if hostPlatform == "macos" {
			return "", fmt.Errorf("--home-only is not supported on macOS hosts yet")
		}
		return applyModeHomeOnly, nil
	}

	switch hostPlatform {
	case "macos":
		return applyModeDarwin, nil
	case "nixos":
		return applyModeNixOS, nil
	default:
		return applyModeLinux, nil
	}
}

func switchDarwin(cmdrDir, flakeRef string) error {
	if commandExistsFn("darwin-rebuild") {
		fmt.Printf("  Running %s...\n", bold("darwin-rebuild switch"))
		if err := runVisibleDirFn(cmdrDir, "sudo", "darwin-rebuild", "switch", "--flake", flakeRef); err != nil {
			return fmt.Errorf("darwin-rebuild switch failed: %w", err)
		}
	} else {
		fmt.Printf("  %s darwin-rebuild not found — running first-time bootstrap...\n", warn("[warn]"))

		// Move /etc/bashrc if it exists (nix-darwin requirement)
		if _, err := os.Stat("/etc/bashrc"); err == nil {
			fmt.Println("  Moving /etc/bashrc to /etc/bashrc.before-nix-darwin...")
			if err := runVisibleFn("sudo", "mv", "/etc/bashrc", "/etc/bashrc.before-nix-darwin"); err != nil {
				return fmt.Errorf("move /etc/bashrc: %w", err)
			}
		}

		fmt.Printf("  Running %s (first-time bootstrap)...\n", bold("nix-darwin"))
		if err := runVisibleDirFn(cmdrDir, "sudo", "nix", "run", "nix-darwin/master#darwin-rebuild", "--", "switch", "--flake", flakeRef); err != nil {
			return fmt.Errorf("nix-darwin bootstrap failed: %w", err)
		}
	}

	fmt.Printf("\n%s Configuration applied\n", pass("[pass]"))
	return nil
}

func switchLinux(cmdrDir, flakeRef string) error {
	fmt.Printf("  Running %s...\n", bold("home-manager switch"))
	if err := runVisibleDirFn(cmdrDir, "home-manager", "switch", "--flake", flakeRef); err != nil {
		return fmt.Errorf("home-manager switch failed: %w", err)
	}

	fmt.Printf("\n%s Configuration applied\n", pass("[pass]"))
	return nil
}

func switchNixOS(cmdrDir, flakeRef string) error {
	fmt.Printf("  Running %s...\n", bold("nixos-rebuild switch"))
	if err := runVisibleDirFn(cmdrDir, "sudo", "nixos-rebuild", "switch", "--flake", flakeRef); err != nil {
		return fmt.Errorf("nixos-rebuild switch failed: %w", err)
	}

	fmt.Printf("\n%s Configuration applied\n", pass("[pass]"))
	return nil
}

func switchHomeOnlyApply(cmdrDir, hostName string) error {
	activationRef := fmt.Sprintf(".#homeConfigurations.%s.activationPackage", hostName)
	fmt.Printf("  Running %s...\n", bold("nix run home activationPackage"))
	if err := runVisibleDirFn(cmdrDir, "nix", "run", activationRef); err != nil {
		return fmt.Errorf("home-only apply failed for %s: %w", activationRef, err)
	}

	fmt.Printf("\n%s Home Manager configuration applied\n", pass("[pass]"))
	return nil
}
