package builder

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// KindDeleteCluster deletes Kind clusters directly via the kind CLI.
// It finds running clusters and deletes each one.
func KindDeleteCluster() error {
	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		return fmt.Errorf("could not list Kind clusters: %w", err)
	}

	clusters := strings.TrimSpace(string(out))
	if clusters == "" {
		fmt.Println("  No Kind clusters found")
		return nil
	}

	for _, name := range strings.Split(clusters, "\n") {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		fmt.Printf("  Deleting Kind cluster: %s\n", name)
		cmd := exec.Command("kind", "delete", "cluster", "--name", name)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("kind delete cluster %s failed: %w", name, err)
		}
	}
	return nil
}

// BuildBackstagePlatform builds the backstage-platform Docker image from source.
// It requires Docker to be running and the backstage-platform source to exist
// at repositories/backstage-platform/.
func BuildBackstagePlatform(orgDir string, verbose bool) error {
	backstageDir := filepath.Join(orgDir, "repositories", "backstage-platform")

	// Check if directory exists
	if _, err := os.Stat(backstageDir); err != nil {
		return fmt.Errorf("backstage-platform source not found at %s", backstageDir)
	}

	// Check for Dockerfile
	dockerfile := filepath.Join(backstageDir, "packages", "backend", "Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("Dockerfile not found at %s", dockerfile)
	}

	fmt.Println("  Building backstage-platform:latest...")

	// Build the image
	args := []string{
		"build",
		"-t", "backstage-platform:latest",
		"-f", "packages/backend/Dockerfile",
		".",
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = backstageDir
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	fmt.Println("  backstage-platform:latest built successfully")
	return nil
}

// BuildTerminal builds the terminal (ttyd + kubectl) Docker image from
// containers/terminal. It requires Docker to be running.
func BuildTerminal(orgDir string, verbose bool) error {
	terminalDir := filepath.Join(orgDir, "containers", "terminal")

	if _, err := os.Stat(terminalDir); err != nil {
		return fmt.Errorf("terminal source not found at %s", terminalDir)
	}

	dockerfile := filepath.Join(terminalDir, "Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("Dockerfile not found at %s", dockerfile)
	}

	fmt.Println("  Building terminal:latest...")

	args := []string{
		"build",
		"-t", "terminal:latest",
		"-f", "Dockerfile",
		".",
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = terminalDir
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	fmt.Println("  terminal:latest built successfully")
	return nil
}

// BuildSandbox builds the sandbox-tty Docker image from containers/sandbox.
// It requires Docker to be running. Before building, it copies the user's
// nvim-astro config from the cmdr submodule into the build context so the
// sandbox's editor matches the workstation; the copy is skipped (with a
// warning) when cmdr or the config directory is absent.
func BuildSandbox(orgDir string, verbose bool) error {
	sandboxDir := filepath.Join(orgDir, "containers", "sandbox")

	if _, err := os.Stat(sandboxDir); err != nil {
		return fmt.Errorf("sandbox source not found at %s", sandboxDir)
	}

	dockerfile := filepath.Join(sandboxDir, "Dockerfile")
	if _, err := os.Stat(dockerfile); err != nil {
		return fmt.Errorf("Dockerfile not found at %s", dockerfile)
	}

	if err := syncNvimAstro(orgDir, sandboxDir); err != nil {
		fmt.Printf("  [warn]        %v\n", err)
	}

	fmt.Println("  Building sandbox-tty:latest...")

	args := []string{
		"build",
		"-t", "sandbox-tty:latest",
		"-f", "Dockerfile",
		".",
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = sandboxDir
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build failed: %w", err)
	}

	fmt.Println("  sandbox-tty:latest built successfully")
	return nil
}

// syncNvimAstro copies the cmdr nvim-astro config into containers/sandbox/hm
// so the Docker build context always reflects the current workstation editor
// config. It returns nil when the copy is not needed.
func syncNvimAstro(orgDir, sandboxDir string) error {
	src := filepath.Join(orgDir, "cmdr", "home", "04-modules", "tui", "graduated", "nvim", "nvim-astro")
	dstRoot := filepath.Join(sandboxDir, "hm", "nvim-astro")

	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("nvim-astro not found at %s (skipping config sync)", src)
	}

	if err := os.RemoveAll(dstRoot); err != nil {
		return fmt.Errorf("could not clear stale nvim-astro copy: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(dstRoot), 0o755); err != nil {
		return fmt.Errorf("could not create %s: %w", filepath.Dir(dstRoot), err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("could not read nvim-astro source: %w", err)
	}
	for _, e := range entries {
		if err := copyTree(filepath.Join(src, e.Name()), filepath.Join(dstRoot, e.Name())); err != nil {
			return fmt.Errorf("could not copy %s: %w", e.Name(), err)
		}
	}

	fmt.Println("  [ok]   synced cmdr nvim-astro into image build context")
	return nil
}

// copyTree recursively copies a file or directory tree from src to dst.
func copyTree(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, data, info.Mode().Perm())
	}
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

// LoadImageIntoKind loads a Docker image into the Kind cluster.
// It auto-detects the Kind cluster name (defaults to "localdev").
func LoadImageIntoKind(image string, verbose bool) error {
	// Detect Kind cluster name
	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		return fmt.Errorf("could not list kind clusters: %w", err)
	}
	clusters := strings.Fields(strings.TrimSpace(string(out)))
	if len(clusters) == 0 {
		return fmt.Errorf("no Kind clusters found — create one first with: unimart freezer up")
	}
	clusterName := clusters[0]

	fmt.Printf("  Loading %s into Kind cluster %s...\n", image, clusterName)

	cmd := exec.Command("kind", "load", "docker-image", image, "--name", clusterName)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kind load docker-image failed: %w", err)
	}

	fmt.Printf("  %s loaded successfully\n", image)
	return nil
}
