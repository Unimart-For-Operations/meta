package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/colima"
	"github.com/Unimart-For-Operations/meta/internal/prereqs"
	"github.com/spf13/cobra"
)

var freezerDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check all prerequisites and workspace health",
	Long:  `Verifies that all required tools are installed and the workspace is properly configured.`,
	RunE:  runFreezerDoctor,
}

func init() {
	freezerCmd.AddCommand(freezerDoctorCmd)
}

func runFreezerDoctor(cmd *cobra.Command, args []string) error {
	failures := 0
	warnings := 0

	// On macOS, ensure DOCKER_HOST points to Colima before running checks
	if prereqs.IsDarwin() && prereqs.CommandExists("colima") && colima.IsRunning() {
		_ = colima.EnsureDockerHost()
	}

	printResult := func(r prereqs.CheckResult) {
		var tag string
		switch r.Status {
		case prereqs.StatusPass:
			tag = pass("[pass]")
		case prereqs.StatusFail:
			tag = fail("[fail]")
			failures++
		case prereqs.StatusWarn:
			tag = warn("[warn]")
			warnings++
		}

		line := fmt.Sprintf("  %s %s", tag, r.Name)
		if r.Version != "" {
			line += fmt.Sprintf(" (%s)", r.Version)
		}
		if r.Detail != "" {
			line += fmt.Sprintf(" — %s", r.Detail)
		}
		fmt.Println(line)
	}

	fmt.Println()
	fmt.Println(bold("Prerequisites:"))
	printResult(prereqs.CheckGo())
	printResult(prereqs.CheckDocker())
	if strings.Contains(os.Getenv("DOCKER_HOST"), "/podman/") {
		fmt.Printf("  %s docker runtime — DOCKER_HOST points to Podman; use Docker Engine on Linux or Colima's Docker daemon on macOS\n", warn("[warn]"))
		warnings++
	}
	if prereqs.IsDarwin() {
		printResult(prereqs.CheckColima())
	}
	printResult(prereqs.CheckKind())
	printResult(prereqs.CheckKubectl())

	fmt.Println()
	fmt.Println(bold("Package Managers:"))
	if prereqs.IsDarwin() {
		if prereqs.HasBrew() {
			fmt.Printf("  %s homebrew\n", pass("[pass]"))
		} else {
			fmt.Printf("  %s homebrew — not found\n", warn("[warn]"))
			warnings++
		}
	}
	if prereqs.HasNix() {
		fmt.Printf("  %s nix\n", pass("[pass]"))
	} else {
		fmt.Printf("  %s nix — not found\n", warn("[warn]"))
		warnings++
	}

	orgDir, err := resolveOrgDir()
	if err != nil {
		fmt.Printf("  %s org directory — %v\n", fail("[fail]"), err)
		failures++
	} else {
		fmt.Println()
		fmt.Println(bold("Workspace:"))
		for _, r := range prereqs.CheckWorkspace(orgDir) {
			printResult(r)
		}
		for _, r := range prereqs.CheckWorkspaceIdpbuilder(orgDir) {
			printResult(r)
		}
	}

	fmt.Println()
	if failures == 0 && warnings == 0 {
		fmt.Println(pass("All checks passed"))
	} else if failures == 0 {
		fmt.Printf("%s %d warning(s)\n", warn("!"), warnings)
	} else {
		fmt.Printf("%s %d issue(s) found", fail("!"), failures)
		if warnings > 0 {
			fmt.Printf(", %d warning(s)", warnings)
		}
		fmt.Println()
		os.Exit(1)
	}

	return nil
}
