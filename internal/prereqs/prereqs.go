// Package prereqs provides prerequisite checking and installation for IDP
// platform tools. It wraps internal/platform for command resolution (including
// Nix-managed paths) and adds tool-specific health checks.
package prereqs

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/platform"
)

// CheckResult represents the outcome of a prerequisite check.
type CheckResult struct {
	Name    string
	Status  Status
	Version string
	Detail  string
}

// Status represents the state of a prerequisite check.
type Status int

const (
	StatusPass Status = iota
	StatusFail
	StatusWarn
)

// String returns the display string for a status.
func (s Status) String() string {
	switch s {
	case StatusPass:
		return "pass"
	case StatusFail:
		return "fail"
	case StatusWarn:
		return "warn"
	default:
		return "unknown"
	}
}

// Platform returns the current OS (darwin, linux).
func Platform() string {
	return runtime.GOOS
}

// Arch returns the current architecture (amd64, arm64).
func Arch() string {
	return runtime.GOARCH
}

// IsDarwin returns true if running on macOS.
func IsDarwin() bool {
	return platform.IsDarwin()
}

// CommandExists checks if a command is available on PATH or Nix-managed paths.
func CommandExists(name string) bool {
	return platform.CommandExists(name)
}

// CommandPath returns the full path to a command, or empty string if not found.
func CommandPath(name string) string {
	path, err := exec.LookPath(name)
	if err != nil {
		return ""
	}
	return path
}

// CommandOutput runs a command and returns its trimmed stdout.
func CommandOutput(name string, args ...string) (string, error) {
	return platform.CommandOutput(name, args...)
}

// CommandOutputCombined runs a command and returns trimmed stdout+stderr.
// Use this for tools like colima that write status info to stderr.
func CommandOutputCombined(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running %s: %w", name, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// RunCommand executes a command, returning any error.
func RunCommand(name string, args ...string) error {
	return platform.RunSilent(name, args...)
}

// RunCommandVisible executes a command with stdout/stderr connected to the terminal.
func RunCommandVisible(name string, args ...string) error {
	return platform.RunVisible(name, args...)
}

// HasNix checks if Nix is available.
func HasNix() bool {
	return CommandExists("nix")
}

// HasBrew checks if Homebrew is available.
func HasBrew() bool {
	return CommandExists("brew")
}

// NixInstall installs a package via nix profile install.
func NixInstall(pkg string) error {
	fmt.Printf("  Installing via Nix: nixpkgs#%s...\n", pkg)
	return RunCommandVisible("nix", "profile", "install", "nixpkgs#"+pkg)
}

// BrewInstall installs packages via Homebrew.
func BrewInstall(pkgs ...string) error {
	fmt.Printf("  Installing via Homebrew: %s...\n", strings.Join(pkgs, ", "))
	args := append([]string{"install"}, pkgs...)
	return RunCommandVisible("brew", args...)
}

// CheckWorkspace verifies the org directory and idpbuilder repo.
func CheckWorkspace(orgDir string) []CheckResult {
	var results []CheckResult

	// Check org directory exists
	if info, err := os.Stat(orgDir); err != nil || !info.IsDir() {
		results = append(results, CheckResult{
			Name:   "org directory",
			Status: StatusFail,
			Detail: orgDir + " — not found",
		})
		return results
	}
	results = append(results, CheckResult{
		Name:    "org directory",
		Status:  StatusPass,
		Version: orgDir,
	})

	return results
}
