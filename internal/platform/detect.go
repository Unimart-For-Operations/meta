// Package platform provides platform detection and command execution utilities.
package platform

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// IsDarwin returns true if we're running on macOS.
func IsDarwin() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if we're running on Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// CommandExists checks if a command is available on PATH or well-known Nix paths.
func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	if err == nil {
		return true
	}
	// Check well-known Nix-managed paths that may not be in PATH
	for _, dir := range []string{
		"/run/current-system/sw/bin",
		"/etc/profiles/per-user/" + os.Getenv("USER") + "/bin",
	} {
		if _, err := os.Stat(dir + "/" + name); err == nil {
			return true
		}
	}
	return false
}

// CommandOutput runs a command and returns its trimmed stdout.
func CommandOutput(name string, args ...string) (string, error) {
	path := resolveCommand(name)
	cmd := exec.Command(path, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run %s: %w", name, err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// CommandOutputSilent runs a command and returns its trimmed stdout, suppressing stderr.
// Use this for commands that are expected to fail (e.g., git describe when no tag exists).
func CommandOutputSilent(name string, args ...string) (string, error) {
	path := resolveCommand(name)
	cmd := exec.Command(path, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("run %s: %w", name, err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// resolveCommand finds the full path of a command, checking PATH and Nix paths.
func resolveCommand(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	for _, dir := range []string{
		"/run/current-system/sw/bin",
		"/etc/profiles/per-user/" + os.Getenv("USER") + "/bin",
	} {
		full := dir + "/" + name
		if _, err := os.Stat(full); err == nil {
			return full
		}
	}
	return name
}

// RunVisible runs a command with stdout/stderr connected to the terminal.
func RunVisible(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunVisibleDir runs a command in a specific directory with stdout/stderr connected.
func RunVisibleDir(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// RunSilent runs a command and returns its error without printing output.
func RunSilent(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// OSVersion returns a human-readable OS version string.
// On Darwin: "macOS 26.3.1". On Linux: PRETTY_NAME from /etc/os-release, or "Linux".
func OSVersion() string {
	if IsDarwin() {
		v, err := CommandOutput("sw_vers", "-productVersion")
		if err != nil {
			return "macOS"
		}
		return "macOS " + v
	}
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			val := strings.TrimPrefix(line, "PRETTY_NAME=")
			val = strings.Trim(val, `"`)
			return val
		}
	}
	return "Linux"
}

// Arch returns the machine hardware architecture (e.g. "arm64", "x86_64").
func Arch() string {
	return runtime.GOARCH
}
