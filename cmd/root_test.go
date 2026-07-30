package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOrgDir(t *testing.T) {
	// Create a temp directory that looks like an org dir
	tmpDir := t.TempDir()
	cmdrDir := filepath.Join(tmpDir, "cmdr")
	if err := os.MkdirAll(cmdrDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create .gitmodules
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitmodules"), []byte("[submodule]"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create cmdr/flake.nix
	if err := os.WriteFile(filepath.Join(cmdrDir, "flake.nix"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("via flag", func(t *testing.T) {
		orgDir = tmpDir
		defer func() { orgDir = "" }()
		dir, err := resolveOrgDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dir != tmpDir {
			t.Errorf("expected %q, got %q", tmpDir, dir)
		}
	})

	t.Run("via env", func(t *testing.T) {
		orgDir = ""
		t.Setenv("UNIMART_ORG_DIR", tmpDir)
		dir, err := resolveOrgDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if dir != tmpDir {
			t.Errorf("expected %q, got %q", tmpDir, dir)
		}
	})
}

func TestIsOrgDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Not an org dir — missing .gitmodules
	if isOrgDir(tmpDir) {
		t.Error("empty dir should not be detected as org dir")
	}

	// Add .gitmodules but no cmdr/flake.nix
	os.WriteFile(filepath.Join(tmpDir, ".gitmodules"), []byte("[submodule]"), 0o644)
	if isOrgDir(tmpDir) {
		t.Error("dir with only .gitmodules should not be detected as org dir")
	}

	// Add cmdr/flake.nix
	cmdrDir := filepath.Join(tmpDir, "cmdr")
	os.MkdirAll(cmdrDir, 0o755)
	os.WriteFile(filepath.Join(cmdrDir, "flake.nix"), []byte("{}"), 0o644)
	if !isOrgDir(tmpDir) {
		t.Error("dir with .gitmodules and cmdr/flake.nix should be detected as org dir")
	}
}
