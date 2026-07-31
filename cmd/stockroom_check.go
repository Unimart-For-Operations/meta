package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/submodule"
	"github.com/spf13/cobra"
)

var stockroomCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Validate cross-repo contracts (CI)",
	Long: `Run contract validation checks across the organization.

Verifies submodule initialization, remote URLs, AGENTS.md presence,
docs directory structure, Makefile convention, and theme export contract.`,
	RunE: runStockroomCheck,
}

func init() {
	stockroomCmd.AddCommand(stockroomCheckCmd)
}

// requiredDocDirs are the expected subdirectories under each source repo's docs/.
var requiredDocDirs = []string{"Contributing", "Getting-Started", "Reference"}

// requiredMakeTargets are the Makefile targets every submodule should have.
var requiredMakeTargets = []string{"help", "hooks"}

// allowedSubmoduleGitHubOrgs are approved GitHub orgs for absolute submodule remotes.
var allowedSubmoduleGitHubOrgs = []string{"Unimart-For-Operations"}

func runStockroomCheck(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	subs, err := submodule.ParseGitmodules(dir)
	if err != nil {
		return err
	}

	width := submodule.MaxDisplayWidth(subs)
	fmtStr := fmt.Sprintf("  %%s %%-%ds %%s\n", width)

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%s\n", bold("Unimart-For-Operations — Contract Validation"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	errors := 0

	// ── 1. Submodule initialization ──────────────────────────────────────
	fmt.Printf("%s\n", bold("[1/6] Submodule initialization"))
	for _, mod := range subs {
		name := mod.DisplayName()
		if mod.IsInitialized(dir) {
			fmt.Printf(fmtStr, pass("[pass]"), name, "")
		} else {
			fmt.Printf(fmtStr, fail("[fail]"), name, "not initialized")
			errors++
		}
	}
	fmt.Println()

	// ── 2. Remote URLs (security check) ─────────────────────────────────
	fmt.Printf("%s\n", bold("[2/6] Remote URLs (security check)"))
	for _, mod := range subs {
		name := mod.DisplayName()
		if remoteLabel, ok := classifySubmoduleRemote(mod.URL); ok {
			fmt.Printf(fmtStr, pass("[pass]"), name, remoteLabel)
		} else {
			fmt.Printf(fmtStr, fail("[fail]"), name, fmt.Sprintf("→ %s (UNEXPECTED REMOTE)", mod.URL))
			errors++
		}
	}
	fmt.Println()

	// ── 3. AGENTS.md presence ───────────────────────────────────────────
	fmt.Printf("%s\n", bold("[3/6] AGENTS.md presence"))
	for _, mod := range subs {
		name := mod.DisplayName()
		agentsFile := filepath.Join(dir, mod.Path, "AGENTS.md")
		if _, err := os.Stat(agentsFile); err == nil {
			fmt.Printf(fmtStr, pass("[pass]"), name, "AGENTS.md")
		} else {
			fmt.Printf(fmtStr, warn("[warn]"), name, "AGENTS.md missing")
		}
	}
	// Also check org root
	rootAgents := filepath.Join(dir, "AGENTS.md")
	if _, err := os.Stat(rootAgents); err == nil {
		fmt.Printf(fmtStr, pass("[pass]"), "(org root)", "AGENTS.md")
	} else {
		fmt.Printf(fmtStr, fail("[fail]"), "(org root)", "AGENTS.md missing")
		errors++
	}
	fmt.Println()

	// ── 4. Docs directory structure ─────────────────────────────────────
	fmt.Printf("%s\n", bold("[4/6] Docs directory structure"))
	for _, mod := range subs {
		if !mod.IsSourceModule(dir) {
			continue
		}
		name := mod.DisplayName()
		docsDir := filepath.Join(dir, mod.Path, "docs")
		if info, err := os.Stat(docsDir); err != nil || !info.IsDir() {
			fmt.Printf(fmtStr, fail("[fail]"), name, "docs/ not found")
			errors++
			continue
		}

		var missing []string
		for _, sub := range requiredDocDirs {
			subDir := filepath.Join(docsDir, sub)
			if info, err := os.Stat(subDir); err != nil || !info.IsDir() {
				missing = append(missing, sub)
			}
		}

		if len(missing) == 0 {
			fmt.Printf(fmtStr, pass("[pass]"), name, "docs/ ("+strings.Join(requiredDocDirs, ", ")+")")
		} else {
			fmt.Printf(fmtStr, warn("[warn]"), name, "docs/ missing: "+strings.Join(missing, " "))
		}
	}
	fmt.Println()

	// ── 5. Makefile convention ──────────────────────────────────────────
	fmt.Printf("%s\n", bold("[5/6] Makefile convention"))
	for _, mod := range subs {
		name := mod.DisplayName()
		makefilePath := filepath.Join(dir, mod.Path, "Makefile")
		if _, err := os.Stat(makefilePath); err != nil {
			// Gracefully skip submodules without a Makefile (e.g., vaults)
			fmt.Printf(fmtStr, cyan("•"), name, "no Makefile (skipped)")
			continue
		}

		targets, err := parseMakefileTargets(makefilePath)
		if err != nil {
			fmt.Printf(fmtStr, fail("[fail]"), name, fmt.Sprintf("Makefile unreadable: %v", err))
			errors++
			continue
		}

		var missingTargets []string
		for _, tgt := range requiredMakeTargets {
			if !targets[tgt] {
				missingTargets = append(missingTargets, tgt)
			}
		}

		if len(missingTargets) == 0 {
			fmt.Printf(fmtStr, pass("[pass]"), name, strings.Join(requiredMakeTargets, ", "))
		} else {
			fmt.Printf(fmtStr, warn("[warn]"), name, "missing targets: "+strings.Join(missingTargets, " "))
		}
	}
	fmt.Println()

	// ── 6. Theme export contract ────────────────────────────────────────
	fmt.Printf("%s\n", bold("[6/6] Theme export contract"))
	themeExport := filepath.Join(dir, "cmdr", "scripts", "theme-export.sh")
	if info, err := os.Stat(themeExport); err == nil && info.Mode()&0111 != 0 {
		fmt.Printf("  %s cmdr/scripts/theme-export.sh (producer)\n", pass("[pass]"))
	} else {
		fmt.Printf("  %s cmdr/scripts/theme-export.sh not found\n", fail("[fail]"))
		errors++
	}

	themeConsumer := filepath.Join(dir, "internal", "theme", "theme.go")
	if _, err := os.Stat(themeConsumer); err == nil {
		fmt.Printf("  %s internal/theme/theme.go (consumer — unimart)\n", pass("[pass]"))
	} else {
		fmt.Printf("  %s internal/theme/theme.go not found\n", warn("[warn]"))
	}
	fmt.Println()

	// ── Result ──────────────────────────────────────────────────────────
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if errors == 0 {
		fmt.Printf("%s\n", pass("All contract checks passed"))
	} else {
		fmt.Printf("%s\n", fail(fmt.Sprintf("%d check(s) failed", errors)))
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return fmt.Errorf("%d contract check(s) failed", errors)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

func classifySubmoduleRemote(remoteURL string) (string, bool) {
	// Relative URLs (../foo.git) resolve against the parent remote,
	// so they inherit the org.
	if strings.HasPrefix(remoteURL, "../") {
		return fmt.Sprintf("→ %s (relative)", remoteURL), true
	}

	if !strings.Contains(remoteURL, "github.com") {
		return "", false
	}

	for _, org := range allowedSubmoduleGitHubOrgs {
		if strings.Contains(remoteURL, "/"+org+"/") || strings.Contains(remoteURL, ":"+org+"/") {
			return fmt.Sprintf("→ %s org", org), true
		}
	}

	return "", false
}

// parseMakefileTargets scans a Makefile and returns a set of target names.
func parseMakefileTargets(path string) (map[string]bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	targets := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Match lines like "target:" or "target: deps"
		if idx := strings.Index(line, ":"); idx > 0 {
			// Skip variable assignments like "CYAN := \033[0;36m"
			if idx+1 < len(line) && line[idx+1] == '=' {
				continue
			}
			name := strings.TrimSpace(line[:idx])
			// Skip variables (contain =), conditionals, etc.
			if strings.ContainsAny(name, " \t=$.#") {
				continue
			}
			targets[name] = true
		}
	}
	return targets, scanner.Err()
}
