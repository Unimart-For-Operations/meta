package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Build runs `make build` in the idpbuilder directory.
func Build(idpbuilderDir string, verbose bool) error {
	makefile := filepath.Join(idpbuilderDir, "Makefile")
	if _, err := os.Stat(makefile); err != nil {
		return fmt.Errorf("Makefile not found in %s — is this a valid idpbuilder checkout?", idpbuilderDir)
	}

	fmt.Printf("  Building idpbuilder in %s...\n", idpbuilderDir)

	cmd := exec.Command("make", "build")
	cmd.Dir = idpbuilderDir
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("make build failed: %w", err)
	}

	// Verify binary was produced
	binary := filepath.Join(idpbuilderDir, "idpbuilder")
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("build completed but binary not found at %s", binary)
	}

	fmt.Println("  Build complete")
	return nil
}

// Create runs `./idpbuilder create` in the idpbuilder directory.
func Create(idpbuilderDir string, extraArgs []string) error {
	binary := filepath.Join(idpbuilderDir, "idpbuilder")
	if _, err := os.Stat(binary); err != nil {
		return fmt.Errorf("idpbuilder binary not found — run: unimart freezer build")
	}

	args := append([]string{"create"}, extraArgs...)
	fmt.Println("  Running: ./idpbuilder create")

	cmd := exec.Command(binary, args...)
	cmd.Dir = idpbuilderDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("idpbuilder create failed: %w", err)
	}

	return nil
}

// Delete tears down the IDP cluster. It prefers `./idpbuilder delete` for a
// clean teardown but falls back to `kind delete cluster` if the binary isn't built.
func Delete(idpbuilderDir string) error {
	binary := filepath.Join(idpbuilderDir, "idpbuilder")
	if _, err := os.Stat(binary); err == nil {
		fmt.Println("  Running: ./idpbuilder delete")

		cmd := exec.Command(binary, "delete")
		cmd.Dir = idpbuilderDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("idpbuilder delete failed: %w", err)
		}
		return nil
	}

	// Fallback: use kind directly
	fmt.Println("  idpbuilder binary not built — falling back to: kind delete cluster")
	return KindDeleteCluster()
}

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
