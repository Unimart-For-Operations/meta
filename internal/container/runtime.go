package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadImageIntoKind loads a local image into a kind cluster. runtime may be
// "docker" or "podman". For docker we use `kind load docker-image` (requires
// image present in the docker daemon). For podman we `podman save` ->
// `kind load image-archive`.
func LoadImageIntoKind(runtime, image string) error {
	cluster, err := detectKindCluster()
	if err != nil {
		return err
	}

	switch runtime {
	case "docker":
		// kind can load directly from the docker daemon
		cmd := exec.Command("kind", "load", "docker-image", image, "--name", cluster)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "podman":
		// save to a temp file, then load archive into kind
		tmpf, err := os.CreateTemp("", "kind-image-*.tar")
		if err != nil {
			return fmt.Errorf("creating temp file: %w", err)
		}
		tmpPath := tmpf.Name()
		tmpf.Close()
		defer os.Remove(tmpPath)

		saveCmd := exec.Command("podman", "save", "-o", tmpPath, image)
		saveCmd.Stdout = os.Stdout
		saveCmd.Stderr = os.Stderr
		if err := saveCmd.Run(); err != nil {
			return fmt.Errorf("podman save failed: %w", err)
		}

		loadCmd := exec.Command("kind", "load", "image-archive", tmpPath, "--name", cluster)
		loadCmd.Stdout = os.Stdout
		loadCmd.Stderr = os.Stderr
		if err := loadCmd.Run(); err != nil {
			return fmt.Errorf("kind load image-archive failed: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported runtime: %s", runtime)
	}
}

func detectKindCluster() (string, error) {
	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		return "", fmt.Errorf("could not list kind clusters: %w", err)
	}
	clusters := strings.Fields(strings.TrimSpace(string(out)))
	if len(clusters) == 0 {
		return "localdev", nil // fallback name used by idpbuilder
	}
	return clusters[0], nil
}

// GatherImagesFromIdpbuilder scans the idpbuilder repository for manifest files
// and extracts image references. It returns a de-duplicated slice of images.
func GatherImagesFromIdpbuilder(idpDir string) ([]string, error) {
	// Preferred path where the bootstrap argocd install lives.
	candidates := []string{
		filepath.Join(idpDir, "pkg/controllers/localbuild/resources/argo/install.yaml"),
	}

	images := map[string]struct{}{}

	// If the preferred file exists, parse it; otherwise, scan resources/ for yaml files.
	var files []string
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			files = append(files, c)
		}
	}
	if len(files) == 0 {
		// Walk the idpbuilder tree for resource yaml files as a fallback.
		_ = filepath.Walk(filepath.Join(idpDir), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
				if strings.Contains(path, string(filepath.Separator)+"resources"+string(filepath.Separator)) {
					files = append(files, path)
				}
			}
			return nil
		})
	}

	if len(files) == 0 {
		// No manifests found; return a small default list to attempt bootstrap.
		return []string{"quay.io/argoproj/argocd:v3.1.7"}, nil
	}

	// regex to capture image: values, allowing quoted or unquoted
	re := regexp.MustCompile(`(?m)^\s*image:\s*(?:"([^"]+)"|'([^']+)'|([^\s]+))`)

	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		s := string(b)
		matches := re.FindAllStringSubmatch(s, -1)
		for _, m := range matches {
			var img string
			if m[1] != "" {
				img = m[1]
			} else if m[2] != "" {
				img = m[2]
			} else {
				img = m[3]
			}
			img = strings.TrimSpace(img)
			if img != "" {
				images[img] = struct{}{}
			}
		}
	}

	out := make([]string, 0, len(images))
	for k := range images {
		out = append(out, k)
	}
	return out, nil
}
