package builder

import (
	"fmt"
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
