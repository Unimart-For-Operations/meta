package cmd

import (
	"fmt"
	"strings"

	"github.com/idpbuilder/meta/internal/prereqs"
	"github.com/spf13/cobra"
)

var bootstrapYes bool

var freezerBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Install missing prerequisites",
	Long: `Checks for missing prerequisites and installs them interactively.

Installs via Nix when available, with Homebrew as a fallback:
  - Go         (nix profile install nixpkgs#go)
  - Docker CLI (nix profile install nixpkgs#docker-client)
  - Kind       (nix profile install nixpkgs#kind)
  - Colima     (brew install colima — macOS only, manages a VM)`,
	RunE: runFreezerBootstrap,
}

func init() {
	freezerBootstrapCmd.Flags().BoolVarP(&bootstrapYes, "yes", "y", false, "Skip confirmation prompts")
	freezerCmd.AddCommand(freezerBootstrapCmd)
}

type installStep struct {
	name    string
	check   func() prereqs.CheckResult
	install func() error
}

func runFreezerBootstrap(cmd *cobra.Command, args []string) error {
	steps := []installStep{
		{
			name:    "Go",
			check:   prereqs.CheckGo,
			install: prereqs.InstallGo,
		},
		{
			name:    "Docker CLI",
			check:   prereqs.CheckDocker,
			install: prereqs.InstallDockerCLI,
		},
		{
			name:    "Podman (optional)",
			check:   prereqs.CheckPodman,
			install: prereqs.InstallPodman,
		},
	}

	// Colima is only needed on macOS (Linux runs Docker daemon natively)
	if prereqs.IsDarwin() {
		steps = append(steps, installStep{
			name:    "Colima",
			check:   prereqs.CheckColima,
			install: prereqs.InstallColima,
		})
	}

	steps = append(steps, installStep{
		name:    "Kind",
		check:   prereqs.CheckKind,
		install: prereqs.InstallKind,
	})

	// Report which package managers are available
	managers := []string{}
	if prereqs.HasNix() {
		managers = append(managers, "nix (primary)")
	}
	if prereqs.HasBrew() {
		managers = append(managers, "brew (fallback)")
	}
	if len(managers) == 0 {
		return fmt.Errorf("no package manager found — install Nix (https://nixos.org) or Homebrew (https://brew.sh)")
	}
	fmt.Printf("Package managers: %s\n\n", bold(strings.Join(managers, ", ")))

	installed := 0
	skipped := 0

	for _, step := range steps {
		result := step.check()
		if result.Status == prereqs.StatusPass || result.Status == prereqs.StatusWarn {
			fmt.Printf("  %s %s — already installed\n", pass("[ok]"), step.name)
			skipped++
			continue
		}

		fmt.Printf("  %s %s — missing\n", fail("[missing]"), step.name)

		if !bootstrapYes {
			if !promptYesNo(fmt.Sprintf("  Install %s? [Y/n] ", step.name), true) {
				fmt.Printf("  Skipping %s\n", step.name)
				continue
			}
		}

		if err := step.install(); err != nil {
			return fmt.Errorf("installing %s: %w", step.name, err)
		}

		// Verify installation
		result = step.check()
		if result.Status == prereqs.StatusFail {
			return fmt.Errorf("%s installed but not found on PATH — you may need to restart your shell", step.name)
		}

		fmt.Printf("  %s %s installed\n", pass("[ok]"), step.name)
		installed++
	}

	fmt.Println()
	if installed > 0 {
		fmt.Printf("%s Installed %d prerequisite(s)\n", pass("done"), installed)
	} else {
		fmt.Printf("%s All prerequisites already installed\n", pass("done"))
	}

	return nil
}
