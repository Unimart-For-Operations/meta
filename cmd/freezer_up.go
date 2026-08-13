package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Unimart-For-Operations/meta/internal/colima"
	"github.com/Unimart-For-Operations/meta/internal/theme"
	"github.com/spf13/cobra"
)

var (
	upCPU                   int
	upMemory                int
	upDisk                  int
	upPublishLocalRepos     bool
	upPublishGiteaURL       string
	upPublishOwner          string
	upPublishToken          string
	upPublishDryRun         bool
	upPublishNonInteractive bool
	upPublishUseSSH         bool
	upPublishSSHKeyPath     string
	upCreateOpts            createOptions
)

var freezerUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Start the full IDP platform",
	Long: `Performs the full startup sequence:
  1. Verify prerequisites (run doctor)
  2. Ensure Docker daemon is reachable (start Colima on macOS)
  3. Generate developer configs (tmux, k9s themes)
  4. Run idpbuilder create (in-process)

Extra arguments after -- are passed to idpbuilder create.`,
	RunE: runFreezerUp,
}

func init() {
	defaults := colima.DefaultConfig()
	freezerUpCmd.Flags().IntVar(&upCPU, "cpu", defaults.CPU, "Colima VM CPU count (macOS only)")
	freezerUpCmd.Flags().IntVar(&upMemory, "memory", defaults.Memory, "Colima VM memory in GB (macOS only)")
	freezerUpCmd.Flags().IntVar(&upDisk, "disk", defaults.Disk, "Colima VM disk in GB (macOS only)")
	freezerUpCmd.Flags().BoolVar(&upPublishLocalRepos, "publish-local-repos", false, "Publish local org repos into in-cluster Gitea (opt-in)")
	freezerUpCmd.Flags().StringVar(&upPublishGiteaURL, "publish-gitea-url", "", "Gitea base URL (required when publishing)")
	freezerUpCmd.Flags().StringVar(&upPublishOwner, "publish-owner", "Unimart-For-Operations", "Gitea owner/org to publish into")
	freezerUpCmd.Flags().StringVar(&upPublishToken, "publish-token", "", "Gitea admin token (optional)")
	freezerUpCmd.Flags().BoolVar(&upPublishDryRun, "publish-dry-run", true, "Dry-run the publish step (default true)")
	freezerUpCmd.Flags().BoolVar(&upPublishNonInteractive, "publish-non-interactive", false, "Run without prompts (requires token and defaults)")
	freezerUpCmd.Flags().BoolVar(&upPublishUseSSH, "publish-use-ssh", false, "Use SSH remotes during publish step (non-interactive)")
	freezerUpCmd.Flags().StringVar(&upPublishSSHKeyPath, "publish-ssh-key-path", "", "Path to public SSH key to upload during publish (non-interactive)")
	addCreateFlags(freezerUpCmd, &upCreateOpts)
	freezerCmd.AddCommand(freezerUpCmd)
}

func runFreezerUp(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	// Step 1: Check prerequisites
	fmt.Printf("%s Checking prerequisites\n\n", bold("[1/4]"))
	if err := checkPlatformPrereqs(); err != nil {
		return err
	}

	// Step 2: Ensure Docker daemon is reachable
	cfg := &colima.Config{CPU: upCPU, Memory: upMemory, Disk: upDisk}
	if err := ensureDocker("[2/4]", cfg); err != nil {
		return err
	}

	// Generate developer configs (tmux, k9s) into the org workspace
	genOut := filepath.Join(orgDir, ".workspace", "generated")
	if err := os.MkdirAll(genOut, 0755); err != nil {
		fmt.Printf("warning: failed to create generated output dir: %v\n", err)
	} else if err := writeWorkspaceThemes(orgDir, genOut); err != nil {
		fmt.Printf("warning: failed to generate dev theme files: %v\n", err)
	}

	// Step 3: Create IDP
	fmt.Printf("\n%s Creating IDP platform\n\n", bold("[3/4]"))

	packagesDir := filepath.Join(orgDir, "packages")
	createArgs := idpCreateArgs(packagesDir, upCreateOpts, args)

	if err := createIDP(cmd.Context(), createArgs); err != nil {
		return err
	}

	fmt.Printf("\n%s IDP platform is running\n", pass("done"))
	fmt.Println()
	fmt.Println("  Access the platform:")
	fmt.Println("    ArgoCD:  https://argocd.cnoe.localtest.me:8443")
	fmt.Println("    Gitea:   https://gitea.cnoe.localtest.me:8443")
	fmt.Println()
	fmt.Println("  Get credentials:  unimart freezer status --secrets")
	fmt.Println("  Stop platform:    unimart freezer down")

	// Optionally publish local repos into in-cluster Gitea
	if upPublishLocalRepos {
		fmt.Println()
		fmt.Println("Publishing local repositories into in-cluster Gitea (opt-in)...")

		if upPublishGiteaURL == "" {
			fmt.Println("  publish skipped: --publish-gitea-url is required")
			return nil
		}
		publishGiteaURL = upPublishGiteaURL
		publishOwner = upPublishOwner
		publishToken = upPublishToken
		publishDryRun = upPublishDryRun
		publishNonInteractive = upPublishNonInteractive
		publishUseSSH = upPublishUseSSH
		publishSSHKeyPath = upPublishSSHKeyPath

		if err := runFreezerReposPublish(nil, nil); err != nil {
			fmt.Printf("warning: publish step failed: %v\n", err)
		}
	}

	return nil
}

// writeWorkspaceThemes writes the tmux/k9s theme files used by the IDP
// workspace. On DMS hosts it copies the matugen-generated themes (which reflect
// the active wallpaper theme, regenerated on every theme change), falling back
// to the static Catppuccin export otherwise.
func writeWorkspaceThemes(orgDir, genOut string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("determining home dir: %w", err)
	}

	// DMS hosts: k9s/tmux colors are owned by matugen (~/.config/k9s/skins/dank.yaml,
	// ~/.config/tmux/dank-theme.conf). Prefer those when present.
	dmsK9s, dmsTmux := theme.DMSThemePaths(home)
	wroteAny := false
	for _, pair := range []struct{ src, dst string }{
		{src: dmsTmux, dst: "tmux-theme.conf"},
		{src: dmsK9s, dst: "k9s-skin.yaml"},
	} {
		if b, err := os.ReadFile(pair.src); err == nil {
			dst := filepath.Join(genOut, pair.dst)
			if err := os.WriteFile(dst, b, 0644); err != nil {
				fmt.Printf("warning: failed to write %s: %v\n", dst, err)
			} else {
				fmt.Printf("wrote %s (DMS matugen)\n", dst)
				wroteAny = true
			}
		}
	}
	if wroteAny {
		return nil
	}

	// Non-DMS fallback: export the static Catppuccin theme.
	t, err := theme.LoadFromOrg(orgDir, "catppuccin-frappe")
	if err != nil {
		return fmt.Errorf("loading theme: %w", err)
	}

	tmuxPath := filepath.Join(genOut, "tmux-theme.conf")
	if err := os.WriteFile(tmuxPath, []byte(theme.GenerateTmuxStatus(t)+"\n"), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", tmuxPath, err)
	}
	fmt.Printf("wrote %s\n", tmuxPath)

	skinPath := filepath.Join(genOut, "k9s-skin.yaml")
	if err := os.WriteFile(skinPath, []byte(theme.GenerateK9sSkin(t)+"\n"), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", skinPath, err)
	}
	fmt.Printf("wrote %s\n", skinPath)

	return nil
}
