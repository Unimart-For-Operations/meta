package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/builder"
	"github.com/Unimart-For-Operations/meta/internal/cluster"
	"github.com/Unimart-For-Operations/meta/internal/gitea"
	"github.com/Unimart-For-Operations/meta/internal/platform"
	"github.com/Unimart-For-Operations/meta/internal/repos"
	"github.com/spf13/cobra"
)

const (
	defaultGiteaURL = "https://gitea.cnoe.localtest.me:8443"
	defaultOwner    = "Unimart-For-Operations"
	argoCDURL       = "https://argocd.cnoe.localtest.me:8443"
)

var (
	openSkipBuild     bool
	openNoBrowser     bool
	openRecreate      bool
	openRebuildImages bool
)

var openCmd = &cobra.Command{
	Use:   "open [-- extra idpbuilder create flags]",
	Short: "Open for business — bring the full IDP platform online",
	Long: `One command to bring the entire org's IDP online locally.

Performs a 7-step startup sequence:
  1. Check prerequisites (Go, Docker, Kind, kubectl)
  2. Start container runtime (Colima on macOS)
  3. Build idpbuilder from source (unless --skip-build)
  4. Build custom images (backstage-platform, unless --skip-build)
  5. Create IDP platform (ArgoCD + Gitea + nginx on Kind)
  6. Load custom images into Kind
  7. Publish all org repos to in-cluster Gitea + open browser

Opinionated defaults: dev password enabled, exit-after-sync mode,
all org repos published to Gitea via HTTPS.

Extra arguments after -- are passed through to idpbuilder create.`,
	RunE: runOpen,
}

func init() {
	openCmd.Flags().BoolVar(&openSkipBuild, "skip-build", false, "Skip the idpbuilder and custom image build steps")
	openCmd.Flags().BoolVar(&openNoBrowser, "no-browser", false, "Don't auto-open the ArgoCD dashboard")
	openCmd.Flags().BoolVar(&openRecreate, "recreate", false, "Tear down existing cluster first, then recreate")
	openCmd.Flags().BoolVar(&openRebuildImages, "rebuild-images", false, "Force rebuild of custom images even if they exist")
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}
	idpDir := filepath.Join(orgDir, "idpbuilder")

	// --recreate: tear down first
	if openRecreate {
		fmt.Printf("%s Tearing down existing cluster\n\n", bold("[0/7]"))
		if err := builder.Delete(idpDir); err != nil {
			// Non-fatal: cluster may not exist yet
			fmt.Printf("  %s teardown: %v (continuing)\n", warn("[warn]"), err)
		} else {
			fmt.Printf("  %s cluster removed\n", pass("[ok]"))
		}
		fmt.Println()
	}

	// Step 1: Check prerequisites
	fmt.Printf("%s Checking prerequisites\n\n", bold("[1/7]"))
	if err := checkPlatformPrereqs(); err != nil {
		return err
	}

	// Step 2: Ensure Docker daemon is reachable
	if err := ensureDocker("[2/7]", nil); err != nil {
		return err
	}

	// Step 3: Build idpbuilder
	if err := buildIdpbuilder(idpDir, "[3/7]", openSkipBuild); err != nil {
		return err
	}

	// Step 4: Build custom images
	skipImageBuild := openSkipBuild && !openRebuildImages
	if err := buildCustomImages("[4/7]", orgDir, skipImageBuild); err != nil {
		return err
	}

	// Step 5: Create IDP platform with opinionated defaults
	fmt.Printf("\n%s Creating IDP platform\n\n", bold("[5/7]"))

	// Resolve packages dir (may not exist yet — that's fine, idpbuilder handles it)
	packagesDir := filepath.Join(orgDir, "packages")
	createArgs := idpCreateArgs(packagesDir, args)

	if err := createIDP(idpDir, createArgs); err != nil {
		return err
	}
	fmt.Printf("  %s IDP platform running\n", pass("[ok]"))

	// Step 6: Load custom images into Kind
	if err := loadCustomImages("[6/7]"); err != nil {
		return err
	}

	// Step 7: Publish all org repos to in-cluster Gitea
	fmt.Printf("\n%s Publishing org repos to in-cluster Gitea\n\n", bold("[7/7]"))

	token, err := cluster.GetGiteaAdminToken(defaultGiteaURL)
	if err != nil {
		fmt.Printf("  %s could not discover Gitea token: %v\n", warn("[warn]"), err)
		fmt.Println("  Skipping publish — run manually: unimart freezer repos publish-to-gitea")
	} else {
		// Ensure the target org exists in Gitea
		if err := gitea.EnsureOrg(defaultGiteaURL, defaultOwner, token, true); err != nil {
			fmt.Printf("  %s could not create org %s: %v\n", warn("[warn]"), defaultOwner, err)
			fmt.Println("  Repos will be created under giteaAdmin instead")
		} else {
			fmt.Printf("  %s org %s ready\n", pass("[ok]"), defaultOwner)
		}

		if err := publishAllRepos(orgDir, token); err != nil {
			fmt.Printf("  %s publish failed: %v\n", warn("[warn]"), err)
			fmt.Println("  Run manually: unimart freezer repos publish-to-gitea")
		} else {
			fmt.Printf("  %s all repos published\n", pass("[ok]"))
		}
	}

	// Open browser
	fmt.Println()
	if openNoBrowser {
		fmt.Println("  Skipping browser (--no-browser)")
	} else {
		if err := platform.OpenBrowser(argoCDURL); err != nil {
			fmt.Printf("  %s could not open browser: %v\n", warn("[warn]"), err)
		} else {
			fmt.Printf("  %s opened %s\n", pass("[ok]"), argoCDURL)
		}
	}

	// Summary
	fmt.Println()
	fmt.Printf("%s unimart is open for business\n", pass("done"))
	fmt.Println()
	fmt.Println("  Access the platform:")
	fmt.Printf("    ArgoCD:  %s\n", argoCDURL)
	fmt.Printf("    Gitea:   %s\n", defaultGiteaURL)
	fmt.Println()
	fmt.Println("  Credentials:")
	fmt.Println("    ArgoCD:  admin / developer")
	fmt.Println("    Gitea:   giteaAdmin / developer")
	fmt.Println()
	fmt.Println("  Close up shop:  unimart close")
	fmt.Println("  Platform status: unimart freezer status")

	return nil
}

// publishAllRepos publishes every local org repo (including meta itself)
// into in-cluster Gitea. Non-interactive, HTTPS, no dry-run.
func publishAllRepos(orgDir, token string) error {
	local, sourceDir, err := repos.ListPublishable(orgDir)
	if err != nil {
		return err
	}
	if len(local) == 0 {
		fmt.Printf("  %s no git repos found in %s\n", warn("[warn]"), sourceDir)
		fmt.Printf("  %s add repos under %s to auto-publish during 'unimart open'\n", dim("[hint]"), repos.RepositoriesDir(orgDir))
		return nil
	}

	allRepos := local
	// Legacy behavior: if repositories/ is absent and we are scanning org root,
	// include the meta repo itself.
	if filepath.Clean(sourceDir) == filepath.Clean(orgDir) {
		metaRepo := repos.LocalRepo{Name: "meta", Path: orgDir}
		allRepos = append([]repos.LocalRepo{metaRepo}, local...)
	}

	for _, r := range allRepos {
		repoName := r.Name
		fmt.Printf("  %s/%s ", defaultOwner, repoName)

		exists, err := gitea.RepoExists(defaultGiteaURL, defaultOwner, repoName, token, true)
		if err != nil {
			fmt.Printf("— %s check failed: %v\n", fail("[fail]"), err)
			continue
		}
		if !exists {
			if err := gitea.CreateRepo(defaultGiteaURL, defaultOwner, repoName, token, true, true); err != nil {
				fmt.Printf("— %s create failed: %v\n", fail("[fail]"), err)
				continue
			}
		}

		repoPath := r.Path
		if repoPath == "" {
			repoPath = filepath.Join(orgDir, repoName)
		}
		if err := repos.SetRemoteAndPush(repoPath, defaultGiteaURL, defaultOwner, repoName, token, false); err != nil {
			fmt.Printf("— %s push failed: %v\n", fail("[fail]"), err)
			continue
		}
		fmt.Printf("— %s\n", pass("[ok]"))
	}

	return nil
}

// hasPackages returns true if the packages directory exists and contains
// at least one file that isn't .gitkeep.
func hasPackages(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(e.Name()) != ".gitkeep" {
			return true
		}
	}
	return false
}
