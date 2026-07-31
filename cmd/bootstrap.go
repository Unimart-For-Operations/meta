package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Unimart-For-Operations/meta/internal/host"
	"github.com/Unimart-For-Operations/meta/internal/platform"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Full setup: submodules, prerequisites, host detection, apply",
	Long: `Run the complete onboarding flow for a fresh clone:

  [1/5] Initialize git submodules
  [2/5] Install prerequisites (Nix, Homebrew on macOS)
  [3/5] Detect or register host configuration
  [4/5] Apply configuration (switch)
  [5/5] Verify environment (doctor)

Safe to re-run — each step is idempotent.`,
	RunE: runBootstrap,
}

func init() {
	deliCmd.AddCommand(bootstrapCmd)
}

func runBootstrap(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	cmdrDir := filepath.Join(dir, "cmdr")

	// [1/5] Submodules
	fmt.Printf("\n%s [1/5] Initializing submodules\n", bold("▸"))
	flakeNix := filepath.Join(cmdrDir, "flake.nix")
	if _, err := os.Stat(flakeNix); err == nil {
		fmt.Printf("  %s Submodules already initialized\n", pass("[ok]"))
	} else {
		fmt.Println("  Running git submodule update --init --recursive...")
		if err := platform.RunVisibleDir(dir, "git", "submodule", "update", "--init", "--recursive"); err != nil {
			return fmt.Errorf("submodule init: %w", err)
		}
		fmt.Printf("  %s Submodules initialized\n", pass("[ok]"))
	}

	// [2/5] Prerequisites
	fmt.Printf("\n%s [2/5] Checking prerequisites\n", bold("▸"))
	if platform.CommandExists("nix") {
		nixVer, _ := platform.CommandOutput("nix", "--version")
		fmt.Printf("  %s Nix installed (%s)\n", pass("[ok]"), nixVer)

		// On macOS, also check Homebrew
		if platform.IsDarwin() && !platform.CommandExists("brew") {
			fmt.Println("  Running cmdr bootstrap for Homebrew...")
			bootstrapScript := filepath.Join(cmdrDir, "scripts", "bootstrap.sh")
			if err := platform.RunVisible("bash", bootstrapScript); err != nil {
				return fmt.Errorf("cmdr bootstrap: %w", err)
			}
		}
	} else {
		fmt.Printf("  %s Nix not found — running cmdr bootstrap...\n", warn("[warn]"))
		bootstrapScript := filepath.Join(cmdrDir, "scripts", "bootstrap.sh")
		if err := platform.RunVisible("bash", bootstrapScript); err != nil {
			return fmt.Errorf("cmdr bootstrap: %w", err)
		}

		// Re-check after bootstrap
		if !platform.CommandExists("nix") {
			fmt.Printf("\n  %s Nix not found in PATH after bootstrap\n", fail("[fail]"))
			fmt.Println("  Restart your shell and re-run: unimart deli bootstrap")
			return nil
		}
	}

	// [3/5] Host detection
	fmt.Printf("\n%s [3/5] Detecting host configuration\n", bold("▸"))
	info, err := host.Detect(dir)
	if err != nil {
		fmt.Printf("  %s %s\n", warn("[warn]"), err)
		fmt.Println("  Registering this machine via cmdr...")
		if err := platform.RunVisibleDir(cmdrDir, "make", "register"); err != nil {
			return fmt.Errorf("host registration: %w", err)
		}
		// Re-detect after registration
		info, err = host.Detect(dir)
		if err != nil {
			return fmt.Errorf("host detection after registration: %w", err)
		}
	}
	fmt.Printf("  %s Host: %s (platform: %s, user: %s)\n", pass("[ok]"), info.Name, info.Platform, info.Username)

	// [4/5] Apply configuration
	fmt.Printf("\n%s [4/5] Applying configuration\n", bold("▸"))
	if err := runSwitch(cmd, []string{info.Name}); err != nil {
		return err
	}

	// [5/5] Verify
	fmt.Printf("\n%s [5/5] Verifying environment\n", bold("▸"))
	if err := platform.RunVisibleDir(cmdrDir, "make", "doctor"); err != nil {
		fmt.Printf("  %s Doctor check reported issues (non-fatal)\n", warn("[warn]"))
	} else {
		fmt.Printf("  %s Environment verified\n", pass("[ok]"))
	}

	fmt.Printf("\n%s Bootstrap complete. Reload your shell: exec zsh\n", pass("[pass]"))
	return nil
}
