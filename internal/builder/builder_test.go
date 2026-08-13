package builder

import (
	"os/exec"
	"strings"
	"testing"
)

func TestBuildBackstagePlatform_MissingSource(t *testing.T) {
	err := BuildBackstagePlatform(t.TempDir(), false)
	if err == nil {
		t.Fatal("BuildBackstagePlatform should fail when source dir is missing")
	}
	if !strings.Contains(err.Error(), "backstage-platform source not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLoadImageIntoKind_NoClusters(t *testing.T) {
	if _, err := exec.LookPath("kind"); err != nil {
		t.Skip("kind not available")
	}
	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		t.Skipf("could not query kind clusters: %v", err)
	}
	if strings.TrimSpace(string(out)) != "" {
		t.Skip("kind clusters exist; skipping no-cluster assertion")
	}

	err = LoadImageIntoKind("some-image:latest", false)
	if err == nil {
		t.Fatal("LoadImageIntoKind should fail with no kind clusters")
	}
	if !strings.Contains(err.Error(), "no Kind clusters found") {
		t.Errorf("unexpected error: %v", err)
	}
}
