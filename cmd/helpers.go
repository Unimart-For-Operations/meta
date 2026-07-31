package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/builder"
	"github.com/Unimart-For-Operations/meta/internal/cluster"
	"github.com/Unimart-For-Operations/meta/internal/colima"
	"github.com/Unimart-For-Operations/meta/internal/prereqs"
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

// buildIdpbuilder builds idpbuilder from source. If skipBuild is true,
// the build is skipped. The stepLabel is printed before the action.
func buildIdpbuilder(idpDir, stepLabel string, skipBuild bool) error {
	fmt.Printf("\n%s Building idpbuilder\n\n", bold(stepLabel))

	if _, err := os.Stat(idpDir); err != nil {
		return fmt.Errorf("idpbuilder directory not found at %s", idpDir)
	}

	if skipBuild {
		fmt.Println("  Skipping build (--skip-build)")
		return nil
	}
	return builder.Build(idpDir, verbose)
}

func createIDP(idpDir string, args []string) error {
	return builder.Create(idpDir, args)
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

// idpCreateArgs builds the idpbuilder create argument list for open/reload.
// It always matches the existing cluster's dev-password setting so a
// reconcile (reload) succeeds without a teardown, while keeping
// --dev-password as the default for fresh clusters.
func idpCreateArgs(packagesDir string, extraArgs []string) []string {
	enabled, determined := cluster.StaticPasswordEnabled()
	args := []string{"--no-exit=false"}
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
