// Package submodule provides git submodule discovery and operations.
//
// Instead of hardcoding submodule paths, this package reads .gitmodules
// at runtime so that adding or removing submodules requires no code changes.
package submodule

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Submodule represents a single git submodule entry from .gitmodules.
type Submodule struct {
	Name string // section header: [submodule "name"]
	Path string // path = value (relative to org root)
	URL  string // url = value
}

// DisplayName returns a short name suitable for terminal output.
// For nested paths like "unimart-employee-handbooks/cdc", returns "cdc".
// For flat paths like "cmdr", returns "cmdr".
func (s Submodule) DisplayName() string {
	return filepath.Base(s.Path)
}

// IsInitialized checks if the submodule directory has a .git entry
// (either a directory for regular clones or a file for submodules).
func (s Submodule) IsInitialized(orgDir string) bool {
	gitEntry := filepath.Join(orgDir, s.Path, ".git")
	_, err := os.Stat(gitEntry)
	return err == nil
}

// IsSourceModule checks whether this submodule produces documentation
// by testing for a docs/ directory at its path.
func (s Submodule) IsSourceModule(orgDir string) bool {
	docsDir := filepath.Join(orgDir, s.Path, "docs")
	info, err := os.Stat(docsDir)
	return err == nil && info.IsDir()
}

// HasDocChanges checks if any files under docs/ changed between two refs.
// Returns false if the diff command fails or produces no output.
func (s Submodule) HasDocChanges(orgDir, oldRef, newRef string) bool {
	out, err := s.GitSilent(orgDir, "diff", "--name-only", oldRef+".."+newRef, "--", "docs/")
	return err == nil && strings.TrimSpace(out) != ""
}

// Git runs a git command inside the submodule and returns trimmed stdout.
// Stderr is connected to the terminal.
func (s Submodule) Git(orgDir string, args ...string) (string, error) {
	modDir := filepath.Join(orgDir, s.Path)
	fullArgs := append([]string{"-C", modDir}, args...)
	cmd := exec.Command("git", fullArgs...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), s.DisplayName(), err)
	}
	return strings.TrimSpace(out.String()), nil
}

// GitSilent runs a git command inside the submodule and returns trimmed stdout.
// Both stdout and stderr are captured; stderr is discarded. Use this for
// commands that may fail normally (e.g., git describe when no tag exists).
func (s Submodule) GitSilent(orgDir string, args ...string) (string, error) {
	modDir := filepath.Join(orgDir, s.Path)
	fullArgs := append([]string{"-C", modDir}, args...)
	cmd := exec.Command("git", fullArgs...)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s in %s: %w", strings.Join(args, " "), s.DisplayName(), err)
	}
	return strings.TrimSpace(out.String()), nil
}

// GitRun runs a git command inside the submodule without capturing output.
// Stdout and stderr are discarded. Returns only the error status.
func (s Submodule) GitRun(orgDir string, args ...string) error {
	modDir := filepath.Join(orgDir, s.Path)
	fullArgs := append([]string{"-C", modDir}, args...)
	cmd := exec.Command("git", fullArgs...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// MaxDisplayWidth returns the longest DisplayName length across a slice
// of submodules. Useful for aligning terminal output columns.
func MaxDisplayWidth(subs []Submodule) int {
	max := 0
	for _, s := range subs {
		if n := len(s.DisplayName()); n > max {
			max = n
		}
	}
	return max
}

// ParseGitmodules reads the .gitmodules file in orgDir and returns all
// submodule entries. Returns an error if the file cannot be read.
//
// Parsing is line-based and handles the standard git config format:
//
//	[submodule "name"]
//	    path = value
//	    url = value
//	    ignore = dirty
func ParseGitmodules(orgDir string) ([]Submodule, error) {
	path := filepath.Join(orgDir, ".gitmodules")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open .gitmodules: %w", err)
	}
	defer f.Close()

	var subs []Submodule
	var current *Submodule

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Section header: [submodule "name"]
		if strings.HasPrefix(line, "[submodule ") && strings.HasSuffix(line, "]") {
			// Flush previous entry
			if current != nil && current.Path != "" {
				subs = append(subs, *current)
			}
			name := line[len("[submodule \"") : len(line)-len("\"]")]
			current = &Submodule{Name: name}
			continue
		}

		if current == nil {
			continue
		}

		// Key-value pairs: key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "path":
			current.Path = val
		case "url":
			current.URL = val
		}
	}

	// Flush last entry
	if current != nil && current.Path != "" {
		subs = append(subs, *current)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .gitmodules: %w", err)
	}

	return subs, nil
}

// Paths returns just the path strings from a slice of submodules.
// Useful for passing to git add.
func Paths(subs []Submodule) []string {
	paths := make([]string, len(subs))
	for i, s := range subs {
		paths[i] = s.Path
	}
	return paths
}
