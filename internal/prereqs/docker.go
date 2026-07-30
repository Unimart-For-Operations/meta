package prereqs

import (
	"fmt"
	"strings"
)

// CheckDocker verifies that the Docker CLI is installed and the daemon is reachable.
func CheckDocker() CheckResult {
	if !CommandExists("docker") {
		return CheckResult{
			Name:   "docker",
			Status: StatusFail,
			Detail: "not installed — run: unimart freezer bootstrap",
		}
	}

	// Check version
	version, err := CommandOutput("docker", "version", "--format", "{{.Client.Version}}")
	if err != nil {
		version = "unknown"
	}

	// Check if daemon is reachable
	if err := RunCommand("docker", "info"); err != nil {
		detail := "CLI installed but daemon not reachable"
		if IsDarwin() {
			detail += " — is Colima running?"
		} else {
			detail += " — is the Docker daemon started?"
		}
		return CheckResult{
			Name:    "docker",
			Status:  StatusWarn,
			Version: version,
			Detail:  detail,
		}
	}

	return CheckResult{
		Name:    "docker",
		Status:  StatusPass,
		Version: version,
	}
}

// CheckColima verifies that Colima is installed and reports its status.
func CheckColima() CheckResult {
	if !CommandExists("colima") {
		return CheckResult{
			Name:   "colima",
			Status: StatusFail,
			Detail: "not installed — run: unimart freezer bootstrap",
		}
	}

	version, err := CommandOutput("colima", "version")
	if err != nil {
		version = "unknown"
	} else {
		// Parse "colima version 0.8.1\n..." -> "0.8.1"
		for _, line := range strings.Split(version, "\n") {
			if strings.HasPrefix(line, "colima version") {
				parts := strings.Fields(line)
				if len(parts) >= 3 {
					version = parts[2]
				}
				break
			}
		}
	}

	// Check if running — colima writes status to stderr, so use combined output
	status, _ := CommandOutputCombined("colima", "status")
	if strings.Contains(strings.ToLower(status), "running") {
		return CheckResult{
			Name:    "colima",
			Status:  StatusPass,
			Version: version,
			Detail:  "running",
		}
	}

	return CheckResult{
		Name:    "colima",
		Status:  StatusWarn,
		Version: version,
		Detail:  "installed but not running — run: unimart freezer up",
	}
}

// InstallDocker installs the Docker CLI via Nix and Colima via Homebrew (macOS only).
func InstallDocker() error {
	if err := InstallDockerCLI(); err != nil {
		return err
	}
	if IsDarwin() {
		return InstallColima()
	}
	return nil
}

// InstallDockerCLI installs only the Docker CLI via Nix (or Homebrew fallback).
func InstallDockerCLI() error {
	if CommandExists("docker") {
		return nil
	}
	if HasNix() {
		if err := NixInstall("docker-client"); err != nil {
			return fmt.Errorf("installing Docker CLI via Nix: %w", err)
		}
	} else if HasBrew() {
		fmt.Println("  Nix not available, falling back to Homebrew for Docker CLI")
		if err := BrewInstall("docker"); err != nil {
			return fmt.Errorf("installing Docker CLI via Homebrew: %w", err)
		}
	} else {
		return fmt.Errorf("no package manager available for Docker CLI (need nix or brew)")
	}
	return nil
}

// InstallColima installs Colima via Homebrew (macOS only).
func InstallColima() error {
	if CommandExists("colima") {
		return nil
	}
	if !IsDarwin() {
		fmt.Println("  Linux detected — skipping Colima (Docker daemon runs natively)")
		return nil
	}
	if !HasBrew() {
		return fmt.Errorf("Homebrew is required to install Colima on macOS (it manages a Lima VM)")
	}
	fmt.Println("  Colima requires Homebrew on macOS (VM lifecycle management)")
	if err := BrewInstall("colima"); err != nil {
		return fmt.Errorf("installing Colima: %w", err)
	}
	return nil
}
