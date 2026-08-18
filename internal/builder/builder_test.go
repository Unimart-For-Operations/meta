package builder

import (
	"os"
	"os/exec"
	"path/filepath"
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

func TestCopyTree_FileAndDir(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "out")

	if err := os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(src, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "sub", "b.lua"), []byte("print(1)"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := copyTree(src, dst); err != nil {
		t.Fatalf("copyTree failed: %v", err)
	}

	for _, p := range []string{"a.txt", "sub/b.lua"} {
		got, err := os.ReadFile(filepath.Join(dst, p))
		if err != nil {
			t.Fatalf("missing copied file %s: %v", p, err)
		}
		if p == "a.txt" && string(got) != "hello" {
			t.Errorf("a.txt content mismatch: %q", got)
		}
	}
	if err := os.RemoveAll(dst); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("expected dst remove to succeed (re-copy path), got %v", err)
	}
	if err := copyTree(src, dst); err != nil {
		t.Fatalf("copyTree onto removed dst should succeed, got %v", err)
	}
}

func TestSyncNvimAstro_MissingSource(t *testing.T) {
	org := t.TempDir()
	sandbox := filepath.Join(org, "containers", "sandbox")
	if err := os.MkdirAll(sandbox, 0o755); err != nil {
		t.Fatal(err)
	}
	err := syncNvimAstro(org, sandbox)
	if err == nil {
		t.Fatal("syncNvimAstro should fail when cmdr nvim-astro is missing")
	}
	if !strings.Contains(err.Error(), "nvim-astro not found") {
		t.Errorf("unexpected error: %v", err)
	}
}
