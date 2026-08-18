package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/builder"
	"github.com/Unimart-For-Operations/meta/internal/cluster"
	"github.com/Unimart-For-Operations/meta/internal/colima"
	"github.com/Unimart-For-Operations/meta/internal/idp"
	"github.com/Unimart-For-Operations/meta/internal/prereqs"
	"github.com/spf13/cobra"
)

// checkPlatformPrereqs runs prerequisite checks and returns an error if
// any non-Docker/Colima checks fail. Docker and Colima failures are
// deferred to ensureDocker() which can start Colima on macOS.
func checkPlatformPrereqs() error {
	checks := []prereqs.CheckResult{
		prereqs.CheckGo(),
		prereqs.CheckDocker(),
		prereqs.CheckKind(),
		prereqs.CheckKubectl(),
	}
	if prereqs.IsDarwin() {
		checks = append(checks, prereqs.CheckColima())
	}

	hasFailures := false
	for _, c := range checks {
		if c.Status == prereqs.StatusFail {
			// Docker and Colima failures are handled by ensureDocker
			if c.Name != "docker" && c.Name != "colima" {
				hasFailures = true
				fmt.Printf("  %s %s — %s\n", fail("[fail]"), c.Name, c.Detail)
			}
		}
	}

	if hasFailures {
		return fmt.Errorf("missing prerequisites — run: unimart freezer bootstrap")
	}
	fmt.Printf("  %s prerequisites OK\n", pass("[ok]"))
	return nil
}

// ensureDocker ensures the Docker daemon is reachable, starting Colima
// on macOS if needed. cfg controls the Colima VM resources; pass nil for
// defaults. The stepLabel is printed before the action (e.g. "[2/6]").
func ensureDocker(stepLabel string, cfg *colima.Config) error {
	if prereqs.IsDarwin() {
		fmt.Printf("\n%s Starting Colima\n\n", bold(stepLabel))

		if !prereqs.CommandExists("colima") {
			return fmt.Errorf("Colima not installed — run: unimart freezer bootstrap")
		}

		c := colima.DefaultConfig()
		if cfg != nil {
			c = *cfg
		}
		if err := colima.Start(c); err != nil {
			return err
		}
	} else {
		fmt.Printf("\n%s Verifying Docker daemon\n\n", bold(stepLabel))
	}

	// Verify Docker is now reachable
	dockerCheck := prereqs.CheckDocker()
	if dockerCheck.Status == prereqs.StatusFail {
		if prereqs.IsDarwin() {
			return fmt.Errorf("Colima started but Docker is not reachable")
		}
		return fmt.Errorf("Docker daemon is not reachable — is the Docker service running?")
	}
	if dockerCheck.Status == prereqs.StatusWarn && !prereqs.IsDarwin() {
		return fmt.Errorf("Docker CLI found but daemon not reachable — start the Docker service (e.g. sudo systemctl start docker)")
	}
	if dockerEndpointIsPodman() {
		return fmt.Errorf("DOCKER_HOST points to Podman's Docker-compatible socket; unimart freezer/open require Docker Engine (Linux) or Colima's Docker daemon (macOS)")
	}
	fmt.Printf("  %s Docker daemon reachable\n", pass("[ok]"))
	return nil
}

// initIDP sets up the in-process idpbuilder logger before create/delete.
func initIDP() error {
	level := "info"
	if verbose {
		level = "debug"
	}
	return idp.SetLogger(level, true)
}

// createIDP runs idpbuilder's create engine in-process with the given args.
func createIDP(ctx context.Context, args []string) error {
	if err := initIDP(); err != nil {
		return err
	}
	return idp.Run(ctx, args)
}

// deleteIDP tears down the IDP cluster in-process.
func deleteIDP(ctx context.Context, name string) error {
	if err := initIDP(); err != nil {
		return err
	}
	return idp.Delete(ctx, name)
}

// devPasswordFlag returns "--dev-password" when the existing cluster was
// created with it, or when the cluster state can't be determined (fresh
// cluster — unimart's default). Returns "" for an existing cluster created
// with a generated password, so idpbuilder's create can reconcile it.
func devPasswordFlag(determined, enabled bool) string {
	if !determined || enabled {
		return "--dev-password"
	}
	return ""
}

// createOptions carries the curated idpbuilder create settings exposed by
// open, reload, and freezer up. Values map 1:1 onto idpbuilder create flags.
type createOptions struct {
	name           string
	kubeVersion    string
	extraPorts     string
	registryConfig []string
	usePathRouting bool
	packageCustom  []string
	kindConfig     string
	ingressHost    string
	recreate       bool
}

// addCreateFlags registers the curated idpbuilder create flags onto cmd.
func addCreateFlags(cmd *cobra.Command, o *createOptions) {
	cmd.Flags().StringVar(&o.name, "name", "localdev", "Name for the build (prefix for Kind cluster name, pods, etc)")
	cmd.Flags().StringVar(&o.kubeVersion, "kube-version", "v1.33.1", "Version of the Kind Kubernetes cluster to create")
	cmd.Flags().StringVar(&o.extraPorts, "extra-ports", "", `List of extra ports to expose (e.g. "22:32222,9090:39090")`)
	cmd.Flags().StringSliceVar(&o.registryConfig, "registry-config", nil, "Paths to registry config, first one that exists is used")
	cmd.Flags().BoolVar(&o.usePathRouting, "use-path-routing", false, "Expose web UIs under a single domain name")
	cmd.Flags().StringSliceVar(&o.packageCustom, "package-custom-file", nil, "Customize core packages (argocd, nginx, gitea) e.g. argocd:/tmp/argocd.yaml")
	cmd.Flags().StringVar(&o.kindConfig, "kind-config", "", "Path or URL to a kind config file")
	cmd.Flags().StringVar(&o.ingressHost, "ingress-host-name", "", "Host name used by ingresses")
	cmd.Flags().BoolVar(&o.recreate, "recreate", false, "Delete the cluster first if it already exists")
}

// idpCreateArgs builds the idpbuilder create argument list for open/reload.
// It always matches the existing cluster's dev-password setting so a
// reconcile (reload) succeeds without a teardown, while keeping
// --dev-password as the default for fresh clusters.
func idpCreateArgs(packagesDir string, o createOptions, extraArgs []string) []string {
	var args []string
	if o.name != "" {
		args = append(args, "--name="+o.name)
	}
	if o.kubeVersion != "" {
		args = append(args, "--kube-version="+o.kubeVersion)
	}
	if o.extraPorts != "" {
		args = append(args, "--extra-ports="+o.extraPorts)
	}
	for _, rc := range o.registryConfig {
		args = append(args, "--registry-config="+rc)
	}
	if o.usePathRouting {
		args = append(args, "--use-path-routing")
	}
	for _, pc := range o.packageCustom {
		args = append(args, "--package-custom-file="+pc)
	}
	if o.kindConfig != "" {
		args = append(args, "--kind-config="+o.kindConfig)
	}
	if o.ingressHost != "" {
		args = append(args, "--ingress-host-name="+o.ingressHost)
	}
	if o.recreate {
		args = append(args, "--recreate")
	}
	args = append(args, "--no-exit=false")

	enabled, determined := cluster.StaticPasswordEnabled()
	if flag := devPasswordFlag(determined, enabled); flag != "" {
		args = append([]string{flag}, args...)
	}
	if hasPackages(packagesDir) {
		args = append(args, "-p", packagesDir)
	}
	return append(args, extraArgs...)
}

func dockerEndpointIsPodman() bool {
	return strings.Contains(os.Getenv("DOCKER_HOST"), "/podman/")
}

// buildCustomImages builds custom container images (backstage-platform,
// terminal). It skips gracefully if a source directory doesn't exist.
func buildCustomImages(stepLabel, orgDir string) error {
	fmt.Printf("\n%s Building custom container images\n\n", bold(stepLabel))

	// Build backstage-platform image
	backstageDir := filepath.Join(orgDir, "repositories", "backstage-platform")
	if _, err := os.Stat(backstageDir); err != nil {
		fmt.Printf("  %s backstage-platform not found, skipping\n", warn("[warn]"))
		return nil
	}
	if err := builder.BuildBackstagePlatform(orgDir, verbose); err != nil {
		return fmt.Errorf("backstage-platform build failed: %w", err)
	}

	// Build terminal image
	terminalDir := filepath.Join(orgDir, "containers", "terminal")
	if _, err := os.Stat(terminalDir); err != nil {
		fmt.Printf("  %s terminal not found, skipping\n", warn("[warn]"))
	} else if err := builder.BuildTerminal(orgDir, verbose); err != nil {
		return fmt.Errorf("terminal build failed: %w", err)
	}

	// Build sandbox-tty image
	sandboxDir := filepath.Join(orgDir, "containers", "sandbox")
	if _, err := os.Stat(sandboxDir); err != nil {
		fmt.Printf("  %s sandbox-tty not found, skipping\n", warn("[warn]"))
	} else if err := builder.BuildSandbox(orgDir, verbose); err != nil {
		return fmt.Errorf("sandbox-tty build failed: %w", err)
	}

	fmt.Printf("  %s all custom images built\n", pass("[ok]"))
	return nil
}

// loadCustomImages loads custom container images into the Kind cluster.
// It checks if images exist locally before attempting to load them.
func loadCustomImages(stepLabel string) error {
	fmt.Printf("\n%s Loading custom images into Kind\n\n", bold(stepLabel))

	images := []string{"backstage-platform:latest", "terminal:latest", "sandbox-tty:latest"}

	for _, img := range images {
		// Check if image exists locally
		cmd := exec.Command("docker", "image", "inspect", img)
		if err := cmd.Run(); err != nil {
			fmt.Printf("  %s %s not found locally, skipping\n", warn("[warn]"), img)
			continue
		}

		if err := builder.LoadImageIntoKind(img, verbose); err != nil {
			return fmt.Errorf("failed to load %s: %w", img, err)
		}
	}

	fmt.Printf("  %s all custom images loaded\n", pass("[ok]"))
	return nil
}
