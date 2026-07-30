package prereqs

import (
	"fmt"
	"strings"
)

// CheckGo verifies that Go is installed and returns version info.
func CheckGo() CheckResult {
	if !CommandExists("go") {
		return CheckResult{
			Name:   "go",
			Status: StatusFail,
			Detail: "not installed — run: unimart freezer bootstrap",
		}
	}

	version, err := CommandOutput("go", "version")
	if err != nil {
		return CheckResult{
			Name:   "go",
			Status: StatusWarn,
			Detail: "installed but could not determine version",
		}
	}

	// Parse "go version go1.22.0 darwin/arm64" -> "go1.22.0"
	parts := strings.Fields(version)
	ver := ""
	if len(parts) >= 3 {
		ver = parts[2]
	}

	return CheckResult{
		Name:    "go",
		Status:  StatusPass,
		Version: ver,
	}
}

// InstallGo installs Go using Nix (preferred) or Homebrew (fallback).
func InstallGo() error {
	if HasNix() {
		return NixInstall("go")
	}
	if HasBrew() {
		fmt.Println("  Nix not available, falling back to Homebrew")
		return BrewInstall("go")
	}
	return fmt.Errorf("no package manager available (need nix or brew)")
}
