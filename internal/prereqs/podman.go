package prereqs

import (
	"fmt"
	"strings"
)

// CheckPodman verifies that the podman CLI is installed and the machine/daemon is reachable.
func CheckPodman() CheckResult {
	if !CommandExists("podman") {
		return CheckResult{
			Name:   "podman",
			Status: StatusFail,
			Detail: "not installed — install via Nix: nix profile add nixpkgs#podman",
		}
	}

	// Check version
	version, err := CommandOutput("podman", "version", "--format", "json")
	if err != nil {
		version = "unknown"
	} else {
		// shorten the JSON-ish output to first token for reporting
		version = strings.SplitN(version, "\n", 2)[0]
	}

	// Check if machine is running (podman machine status may not exist on all platforms)
	if err := RunCommand("podman", "info"); err != nil {
		return CheckResult{
			Name:    "podman",
			Status:  StatusWarn,
			Version: version,
			Detail:  "podman CLI present but machine not running — run: podman machine init && podman machine start",
		}
	}

	return CheckResult{
		Name:    "podman",
		Status:  StatusPass,
		Version: version,
	}
}

// InstallPodman installs podman via Nix or Homebrew fallback.
func InstallPodman() error {
	if CommandExists("podman") {
		return nil
	}
	if HasNix() {
		if err := NixInstall("podman"); err != nil {
			return fmt.Errorf("installing podman via Nix: %w", err)
		}
		return nil
	}
	if HasBrew() {
		if err := BrewInstall("podman"); err != nil {
			return fmt.Errorf("installing podman via Homebrew: %w", err)
		}
		return nil
	}
	return fmt.Errorf("no package manager available to install podman (need nix or brew)")
}
