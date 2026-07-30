package prereqs

import "fmt"

// CheckKind verifies that Kind is installed.
func CheckKind() CheckResult {
	if !CommandExists("kind") {
		return CheckResult{
			Name:   "kind",
			Status: StatusFail,
			Detail: "not installed — run: unimart freezer bootstrap",
		}
	}

	version, err := CommandOutput("kind", "version")
	if err != nil {
		version = "unknown"
	}

	return CheckResult{
		Name:    "kind",
		Status:  StatusPass,
		Version: version,
	}
}

// InstallKind installs Kind via Nix (preferred) or Homebrew (fallback).
func InstallKind() error {
	if HasNix() {
		return NixInstall("kind")
	}
	if HasBrew() {
		fmt.Println("  Nix not available, falling back to Homebrew")
		return BrewInstall("kind")
	}
	return fmt.Errorf("no package manager available (need nix or brew)")
}
