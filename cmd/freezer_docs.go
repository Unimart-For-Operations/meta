package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/cluster"
	"github.com/Unimart-For-Operations/meta/internal/container"
	"github.com/Unimart-For-Operations/meta/internal/gitea"
	"github.com/Unimart-For-Operations/meta/internal/platform"
	"github.com/Unimart-For-Operations/meta/internal/repos"
	"github.com/spf13/cobra"
)

const (
	docsRepoName    = "docs-service"
	docsImage       = "docs-service:latest"
	docsNamespace   = "docs-service"
	docsSecretName  = "docs-service-secrets"
	docsURL         = "https://docs.cnoe.localtest.me:8443"
	docsGiteaInURL  = "http://my-gitea-http.gitea.svc.cluster.local:3000"
	docsArgoAppName = "docs-service"
)

var docsSkipBuild bool

var freezerDocsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Docs microservice lifecycle (deliberate deploy, dogfooding)",
	Long: `The docs service is intentionally not part of 'unimart open'.
Bring it up when you want a real microservice on the platform to observe —
and good docs are a nice side effect.

  up      Build image, publish repo to Gitea, deploy via ArgoCD
  status  Deployment + ArgoCD application health
  down    Remove the service from the cluster
  open    Open the docs UI in a browser`,
}

var freezerDocsUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Build, publish, and deploy the docs service",
	RunE:  runFreezerDocsUp,
}

var freezerDocsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show docs service deployment status",
	RunE:  runFreezerDocsStatus,
}

var freezerDocsDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Remove the docs service from the cluster",
	RunE:  runFreezerDocsDown,
}

var freezerDocsOpenCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the docs UI in a browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := platform.OpenBrowser(docsURL); err != nil {
			return err
		}
		fmt.Printf("%s opened %s\n", pass("[ok]"), docsURL)
		return nil
	},
}

func init() {
	freezerDocsUpCmd.Flags().BoolVar(&docsSkipBuild, "skip-build", false, "Skip the container image build + load")
	freezerDocsCmd.AddCommand(freezerDocsUpCmd, freezerDocsStatusCmd, freezerDocsDownCmd, freezerDocsOpenCmd)
	freezerCmd.AddCommand(freezerDocsCmd)
}

// docsServiceDir returns the docs-service source directory, preferring the
// repositories/ symlink (publish source of truth).
func docsServiceDir(orgDir string) (string, error) {
	candidates := []string{
		filepath.Join(repos.RepositoriesDir(orgDir), docsRepoName),
		filepath.Join(orgDir, docsRepoName),
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "Dockerfile")); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("docs-service repo not found under %s", repos.RepositoriesDir(orgDir))
}

func runFreezerDocsUp(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}
	docsDir, err := docsServiceDir(orgDir)
	if err != nil {
		return err
	}

	if !cluster.IsClusterRunning() {
		return fmt.Errorf("no Kind cluster running — bring the platform up first: unimart open")
	}

	// Step 1: Build the container image and load it into Kind
	fmt.Printf("%s Building container image\n\n", bold("[1/4]"))
	if docsSkipBuild {
		fmt.Println("  Skipping build (--skip-build)")
	} else {
		if err := platform.RunVisibleDir(docsDir, "docker", "build", "-t", docsImage, "."); err != nil {
			return fmt.Errorf("docker build failed: %w", err)
		}
		if err := container.LoadImageIntoKind("docker", docsImage); err != nil {
			return fmt.Errorf("loading image into Kind: %w", err)
		}
		fmt.Printf("  %s %s built and loaded into Kind\n", pass("[ok]"), docsImage)
	}

	// Step 2: Publish the repo to in-cluster Gitea
	fmt.Printf("\n%s Publishing %s to Gitea\n\n", bold("[2/4]"), docsRepoName)
	token, err := cluster.GetGiteaAdminToken(defaultGiteaURL)
	if err != nil {
		return fmt.Errorf("could not discover Gitea admin token: %w", err)
	}
	if err := gitea.EnsureOrg(defaultGiteaURL, defaultOwner, token, true); err != nil {
		return fmt.Errorf("ensuring org %s: %w", defaultOwner, err)
	}
	exists, err := gitea.RepoExists(defaultGiteaURL, defaultOwner, docsRepoName, token, true)
	if err != nil {
		return err
	}
	if !exists {
		if err := gitea.CreateRepo(defaultGiteaURL, defaultOwner, docsRepoName, token, true, true); err != nil {
			return err
		}
	}
	if err := repos.SetRemoteAndPush(docsDir, defaultGiteaURL, defaultOwner, docsRepoName, token, false); err != nil {
		return fmt.Errorf("pushing to Gitea: %w", err)
	}
	fmt.Printf("  %s %s/%s pushed\n", pass("[ok]"), defaultOwner, docsRepoName)

	// Step 3: Secrets — app secret + ArgoCD repo credentials (repo is private)
	fmt.Printf("\n%s Configuring secrets\n\n", bold("[3/4]"))
	if err := ensureDocsSecrets(token); err != nil {
		return err
	}

	// Step 4: Apply the ArgoCD Application
	fmt.Printf("\n%s Deploying via ArgoCD\n\n", bold("[4/4]"))
	appManifest := filepath.Join(docsDir, "argocd", "application.yaml")
	if err := kubectlApplyFile(appManifest); err != nil {
		return fmt.Errorf("applying ArgoCD application: %w", err)
	}
	fmt.Printf("  %s ArgoCD application %s applied\n", pass("[ok]"), docsArgoAppName)

	// Restart so imagePullPolicy: Never picks up a freshly loaded image.
	// Ignore errors — on first deploy the Deployment may not exist yet.
	if !docsSkipBuild {
		_ = exec.Command("kubectl", "rollout", "restart",
			"deployment/"+docsRepoName, "-n", docsNamespace).Run()
	}

	fmt.Println()
	fmt.Printf("%s docs service deployed\n", pass("done"))
	fmt.Println()
	fmt.Printf("  Docs UI:  %s (may take a minute to sync)\n", docsURL)
	fmt.Println("  Status:   unimart freezer docs status")
	return nil
}

// ensureDocsSecrets creates the app namespace + secret and the ArgoCD repo
// credential secret so ArgoCD can pull the private Gitea repo.
func ensureDocsSecrets(giteaToken string) error {
	// Namespace must exist before the secret (ArgoCD also uses CreateNamespace,
	// but the secret is unmanaged and can arrive first).
	if err := kubectlApplyStdin(fmt.Sprintf(
		"apiVersion: v1\nkind: Namespace\nmetadata:\n  name: %s\n", docsNamespace)); err != nil {
		return fmt.Errorf("creating namespace: %w", err)
	}

	// Preserve SECRET_KEY_BASE across runs so restarts don't churn sessions.
	secretKeyBase, err := existingSecretValue(docsNamespace, docsSecretName, "SECRET_KEY_BASE")
	if err != nil || secretKeyBase == "" {
		buf := make([]byte, 48)
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("generating SECRET_KEY_BASE: %w", err)
		}
		secretKeyBase = hex.EncodeToString(buf)
	}

	if err := kubectlApplySecret(docsNamespace, docsSecretName, map[string]string{
		"SECRET_KEY_BASE": secretKeyBase,
		"GITEA_TOKEN":     giteaToken,
	}, nil); err != nil {
		return fmt.Errorf("creating app secret: %w", err)
	}
	fmt.Printf("  %s secret %s/%s ready\n", pass("[ok]"), docsNamespace, docsSecretName)

	// ArgoCD repository credentials for the private Gitea repo.
	username, password, err := cluster.GetGiteaAdminCredentials()
	if err != nil {
		return fmt.Errorf("discovering Gitea admin credentials: %w", err)
	}
	repoURL := fmt.Sprintf("%s/%s/%s.git", docsGiteaInURL, defaultOwner, docsRepoName)
	if err := kubectlApplySecret("argocd", "docs-service-repo", map[string]string{
		"type":     "git",
		"url":      repoURL,
		"username": username,
		"password": password,
	}, map[string]string{"argocd.argoproj.io/secret-type": "repository"}); err != nil {
		return fmt.Errorf("creating ArgoCD repo credential: %w", err)
	}
	fmt.Printf("  %s ArgoCD repo credentials ready\n", pass("[ok]"))
	return nil
}

func runFreezerDocsStatus(cmd *cobra.Command, args []string) error {
	if !cluster.IsClusterRunning() {
		return fmt.Errorf("no Kind cluster running")
	}

	fmt.Printf("%s docs service status\n\n", bold(">>"))

	// ArgoCD application health
	apps, err := cluster.GetArgoApps()
	if err != nil {
		return err
	}
	found := false
	for _, a := range apps {
		if a.Name == docsArgoAppName {
			found = true
			mark := pass("[ok]")
			if a.Health != "Healthy" || a.Status != "Synced" {
				mark = warn("[warn]")
			}
			fmt.Printf("  %s ArgoCD app: sync=%s health=%s\n", mark, a.Status, a.Health)
		}
	}
	if !found {
		fmt.Printf("  %s ArgoCD app not found — deploy with: unimart freezer docs up\n", warn("[warn]"))
		return nil
	}

	// Deployment readiness
	out, err := exec.Command("kubectl", "get", "deployment", docsRepoName,
		"-n", docsNamespace,
		"-o", "jsonpath={.status.readyReplicas}/{.status.replicas}").Output()
	if err != nil {
		fmt.Printf("  %s deployment not found yet (still syncing?)\n", warn("[warn]"))
	} else {
		ready := strings.TrimSpace(string(out))
		mark := pass("[ok]")
		if !strings.HasPrefix(ready, "1/") || ready == "/1" {
			mark = warn("[warn]")
		}
		fmt.Printf("  %s deployment replicas ready: %s\n", mark, ready)
	}

	fmt.Printf("\n  Docs UI: %s\n", docsURL)
	return nil
}

func runFreezerDocsDown(cmd *cobra.Command, args []string) error {
	if !cluster.IsClusterRunning() {
		return fmt.Errorf("no Kind cluster running")
	}

	fmt.Printf("%s Removing docs service\n\n", bold(">>"))

	// Delete the Application; automated prune removes the workload resources.
	if err := exec.Command("kubectl", "delete", "application", docsArgoAppName,
		"-n", "argocd", "--ignore-not-found").Run(); err != nil {
		return fmt.Errorf("deleting ArgoCD application: %w", err)
	}
	fmt.Printf("  %s ArgoCD application removed\n", pass("[ok]"))

	_ = exec.Command("kubectl", "delete", "secret", "docs-service-repo",
		"-n", "argocd", "--ignore-not-found").Run()
	_ = exec.Command("kubectl", "delete", "namespace", docsNamespace,
		"--ignore-not-found").Run()
	fmt.Printf("  %s namespace %s removed\n", pass("[ok]"), docsNamespace)

	fmt.Printf("\n%s docs service closed\n", pass("done"))
	return nil
}

// --- kubectl helpers ---

func kubectlApplyFile(path string) error {
	out, err := exec.Command("kubectl", "apply", "-f", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func kubectlApplyStdin(manifest string) error {
	cmd := exec.Command("kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// kubectlApplySecret idempotently applies an Opaque secret via
// `kubectl create --dry-run=client -o yaml | kubectl apply`.
func kubectlApplySecret(namespace, name string, data map[string]string, labels map[string]string) error {
	args := []string{"create", "secret", "generic", name, "-n", namespace,
		"--dry-run=client", "-o", "yaml"}
	for k, v := range data {
		args = append(args, fmt.Sprintf("--from-literal=%s=%s", k, v))
	}
	dryRun := exec.Command("kubectl", args...)
	manifest, err := dryRun.Output()
	if err != nil {
		return fmt.Errorf("rendering secret %s/%s: %w", namespace, name, err)
	}

	rendered := string(manifest)
	if len(labels) > 0 {
		var lines []string
		lines = append(lines, "  labels:")
		for k, v := range labels {
			lines = append(lines, fmt.Sprintf("    %s: %q", k, v))
		}
		rendered = strings.Replace(rendered,
			"metadata:\n", "metadata:\n"+strings.Join(lines, "\n")+"\n", 1)
	}

	return kubectlApplyStdin(rendered)
}

// existingSecretValue returns the decoded value of a key in an existing
// secret, or empty string if the secret/key doesn't exist.
func existingSecretValue(namespace, name, key string) (string, error) {
	out, err := exec.Command("kubectl", "get", "secret", name, "-n", namespace,
		"-o", fmt.Sprintf("jsonpath={.data.%s}", key)).Output()
	if err != nil {
		return "", err
	}
	return decodeSecretB64(strings.TrimSpace(string(out)))
}

func decodeSecretB64(s string) (string, error) {
	if s == "" {
		return "", nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
