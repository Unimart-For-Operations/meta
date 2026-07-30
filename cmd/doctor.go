package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/idpbuilder/meta/internal/host"
	"github.com/idpbuilder/meta/internal/platform"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system prerequisites and environment health",
	RunE:  runDoctor,
}

func init() {
	deliCmd.AddCommand(doctorCmd)
}

type check struct {
	name  string
	check func() (string, bool)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	fmt.Printf("%s System Health Check\n\n", bold("▸"))

	// ── Environment ──────────────────────────────────────────────────────────
	fmt.Printf("  %s\n", dim("Environment"))

	envChecks := []check{
		{"host", checkHost(dir)},
		{"platform", checkPlatform()},
		{"unimart", checkUnimartVersion()},
		{"cmdr flake", checkCmdrFlake(dir)},
	}

	allOk := true
	for _, c := range envChecks {
		detail, ok := c.check()
		printCheck(c.name, detail, ok, false) // env items are informational — don't fail allOk
	}

	// ── Tools ─────────────────────────────────────────────────────────────────
	fmt.Printf("\n  %s\n", dim("Tools"))

	toolChecks := []check{
		{"git", checkCommand("git", "--version")},
		{"nix", checkCommand("nix", "--version")},
	}

	hostPlatform := detectedHostPlatform(dir)
	toolChecks = append(toolChecks, platformToolChecks(hostPlatform)...)

	toolChecks = append(toolChecks,
		check{"docker", checkCommand("docker", "--version")},
		check{"kubectl", checkCommand("kubectl", "version", "--client")},
		check{"kind", checkCommand("kind", "--version")},
		check{"go", checkCommand("go", "version")},
	)

	for _, c := range toolChecks {
		detail, ok := c.check()
		printCheck(c.name, detail, ok, true)
		if !ok {
			allOk = false
		}
	}

	// ── Workspace ─────────────────────────────────────────────────────────────
	fmt.Printf("\n  %s\n", dim("Workspace"))
	fmt.Printf("  %s %-18s %s\n", pass("[ok]"), "org-dir", dir)

	if !allOk {
		fmt.Printf("\n%s Some tools are missing. Run: unimart deli bootstrap\n", warn("[warn]"))
	} else {
		fmt.Printf("\n%s All checks passed\n", pass("[pass]"))
	}

	return nil
}

// platformToolChecks returns the platform-specific apply-tool checks:
// macOS uses darwin-rebuild (+brew), NixOS uses nixos-rebuild, and other
// Linux distros use standalone home-manager.
func platformToolChecks(hostPlatform string) []check {
	switch hostPlatform {
	case "macos":
		return []check{
			{"darwin-rebuild", checkExists("darwin-rebuild")},
			{"brew", checkCommand("brew", "--version")},
		}
	case "nixos":
		return []check{
			{"nixos-rebuild", checkExists("nixos-rebuild")},
		}
	default:
		return []check{
			{"home-manager", checkCommand("home-manager", "--version")},
		}
	}
}

func detectedHostPlatform(orgDir string) string {
	if h, err := host.Detect(orgDir); err == nil && h.Platform != "" {
		return h.Platform
	}

	if platform.IsDarwin() {
		return "macos"
	}

	return "linux"
}

func printCheck(name, detail string, ok, fatal bool) {
	if ok {
		fmt.Printf("  %s %-18s %s\n", pass("[ok]"), name, detail)
	} else {
		marker := warn("[--]")
		if fatal {
			marker = warn("[--]")
		}
		msg := "not found"
		if detail != "" {
			msg = detail
		}
		fmt.Printf("  %s %-18s %s\n", marker, name, msg)
	}
}

// checkHost detects the cmdr host config for the current user.
func checkHost(orgDir string) func() (string, bool) {
	return func() (string, bool) {
		h, err := host.Detect(orgDir)
		if err != nil {
			return "not detected", false
		}
		return fmt.Sprintf("%s (%s)", h.Name, h.Platform), true
	}
}

// checkPlatform returns OS version + architecture.
func checkPlatform() func() (string, bool) {
	return func() (string, bool) {
		return fmt.Sprintf("%s %s", platform.OSVersion(), platform.Arch()), true
	}
}

// checkUnimartVersion returns the current unimart build version.
func checkUnimartVersion() func() (string, bool) {
	return func() (string, bool) {
		return Version, true
	}
}

// checkCmdrFlake reads the cmdr flake metadata and returns a short summary.
func checkCmdrFlake(orgDir string) func() (string, bool) {
	return func() (string, bool) {
		cmdrDir := filepath.Join(orgDir, "cmdr")
		out, err := platform.CommandOutputSilent("nix", "flake", "metadata", cmdrDir, "--json")
		if err != nil {
			return "unavailable", false
		}

		var meta struct {
			Locked struct {
				Rev          string `json:"rev"`
				LastModified int64  `json:"lastModified"`
			} `json:"locked"`
		}
		if err := json.Unmarshal([]byte(out), &meta); err != nil {
			return "unavailable", false
		}

		rev := meta.Locked.Rev
		if len(rev) > 7 {
			rev = rev[:7]
		}
		age := humanAge(time.Unix(meta.Locked.LastModified, 0))
		return fmt.Sprintf("rev %s, %s", rev, age), true
	}
}

// humanAge returns a human-readable relative time string (e.g. "3 days ago").
func humanAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < 2*time.Minute:
		return "just now"
	case d < 2*time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	case d < 14*24*time.Hour:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	case d < 60*24*time.Hour:
		return fmt.Sprintf("%d weeks ago", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%d months ago", int(d.Hours()/(24*30)))
	}
}

func checkCommand(name string, args ...string) func() (string, bool) {
	return func() (string, bool) {
		if !platform.CommandExists(name) {
			return "", false
		}
		out, err := platform.CommandOutput(name, args...)
		if err != nil {
			return "", false
		}
		if idx := strings.IndexByte(out, '\n'); idx >= 0 {
			out = out[:idx]
		}
		return out, true
	}
}

// checkExists verifies a command exists without running it with args.
func checkExists(name string) func() (string, bool) {
	return func() (string, bool) {
		if platform.CommandExists(name) {
			return "found", true
		}
		return "", false
	}
}
