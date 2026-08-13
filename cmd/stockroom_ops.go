package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/submodule"
	"github.com/spf13/cobra"
)

var stockroomStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show submodule state",
	RunE:  runStockroomStatus,
}

var stockroomDriftCmd = &cobra.Command{
	Use:   "drift",
	Short: "Check whether submodules are ahead/behind origin/main",
	RunE:  runStockroomDrift,
}

var stockroomUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update submodules to the latest origin/main",
	RunE:  runStockroomUpdate,
}

func init() {
	stockroomCmd.AddCommand(stockroomStatusCmd)
	stockroomCmd.AddCommand(stockroomDriftCmd)
	stockroomCmd.AddCommand(stockroomUpdateCmd)
}

func runStockroomStatus(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	subs, err := submodule.ParseGitmodules(dir)
	if err != nil {
		return err
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%s\n", bold("Submodule Status"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for _, mod := range subs {
		modDir := filepath.Join(dir, mod.Path)
		if _, err := os.Stat(filepath.Join(modDir, ".git")); err != nil {
			fmt.Printf("  %s %-14s not initialized\n", fail("[fail]"), mod.DisplayName())
			continue
		}

		branch, _ := gitOutput(modDir, "rev-parse", "--abbrev-ref", "HEAD")
		short, _ := gitOutput(modDir, "rev-parse", "--short", "HEAD")
		tag, _ := gitOutput(modDir, "describe", "--tags", "--exact-match", "HEAD")
		dirty, _ := gitOutput(modDir, "status", "--porcelain")

		ref := branch + "@" + short
		if tag != "" {
			ref = tag + " (" + branch + "@" + short + ")"
		}
		if dirty != "" {
			ref += " " + warn("(dirty)")
		}

		fmt.Printf("  %s %-14s %s\n", cyan("[ok]"), mod.DisplayName(), ref)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

func runStockroomDrift(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	subs, err := submodule.ParseGitmodules(dir)
	if err != nil {
		return err
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%s\n", bold("Submodule Drift Check"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	changes := 0
	for _, mod := range subs {
		modDir := filepath.Join(dir, mod.Path)
		if _, err := os.Stat(filepath.Join(modDir, ".git")); err != nil {
			fmt.Printf("  %s %-14s not initialized\n", fail("[fail]"), mod.DisplayName())
			changes++
			continue
		}

		if _, err := gitOutput(modDir, "fetch", "origin", "main", "--quiet"); err != nil {
			fmt.Printf("  %s %-14s fetch failed\n", warn("[warn]"), mod.DisplayName())
			changes++
			continue
		}

		local, _ := gitOutput(modDir, "rev-parse", "HEAD")
		remote, _ := gitOutput(modDir, "rev-parse", "origin/main")
		if local == remote {
			fmt.Printf("  %s %-14s up to date\n", pass("[pass]"), mod.DisplayName())
			continue
		}

		behind, _ := gitOutput(modDir, "rev-list", "--count", "HEAD..origin/main")
		ahead, _ := gitOutput(modDir, "rev-list", "--count", "origin/main..HEAD")
		status := ""
		if behind != "0" && behind != "" {
			status = behind + " behind"
		}
		if ahead != "0" && ahead != "" {
			if status != "" {
				status += ", "
			}
			status += ahead + " ahead"
		}
		if status == "" {
			status = "drifted"
		}
		fmt.Printf("  %s %-14s %s\n", warn("[warn]"), mod.DisplayName(), status)
		changes++
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if changes == 0 {
		fmt.Printf("%s\n", pass("All submodules in sync"))
	} else {
		fmt.Printf("%s\n", warn(fmt.Sprintf("%d submodule(s) drifted", changes)))
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

func runStockroomUpdate(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	subs, err := submodule.ParseGitmodules(dir)
	if err != nil {
		return err
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%s\n", bold("Updating submodules to origin/main"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	for _, mod := range subs {
		modDir := filepath.Join(dir, mod.Path)
		if _, err := os.Stat(filepath.Join(modDir, ".git")); err != nil {
			fmt.Printf("  %s %-14s not initialized\n", fail("[fail]"), mod.DisplayName())
			continue
		}

		before, _ := gitOutput(modDir, "rev-parse", "--short", "HEAD")
		if _, err := gitOutput(modDir, "fetch", "origin", "main", "--quiet"); err != nil {
			fmt.Printf("  %s %-14s fetch failed\n", warn("[warn]"), mod.DisplayName())
			continue
		}
		if _, err := gitOutput(modDir, "checkout", "main", "--quiet"); err != nil {
			fmt.Printf("  %s %-14s checkout main failed\n", warn("[warn]"), mod.DisplayName())
			continue
		}
		if _, err := gitOutput(modDir, "merge", "--ff-only", "origin/main", "--quiet"); err != nil {
			fmt.Printf("  %s %-14s merge failed\n", warn("[warn]"), mod.DisplayName())
			continue
		}
		after, _ := gitOutput(modDir, "rev-parse", "--short", "HEAD")
		if before == after {
			fmt.Printf("  %s %-14s already up to date (%s)\n", pass("[pass]"), mod.DisplayName(), before)
		} else {
			fmt.Printf("  %s %-14s %s -> %s\n", pass("[pass]"), mod.DisplayName(), before, after)
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%s\n", pass("Submodule update complete"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	return nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
