package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveOrgDir(t *testing.T) {
	// Create a temp directory that looks like an org dir
	tmpDir := t.TempDir()
	for _, marker := range []string{".gitmodules", "go.mod", "flake.nix"} {
		if err := os.WriteFile(filepath.Join(tmpDir, marker), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
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

	t.Run("via walk-up from nested cwd", func(t *testing.T) {
		orgDir = ""
		t.Setenv("UNIMART_ORG_DIR", "")

		nested := filepath.Join(tmpDir, "cmdr", "home", "02-hosts")
		if err := os.MkdirAll(nested, 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(nested)

		dir, err := resolveOrgDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Resolve symlinks — macOS TempDir lives under /private
		wantResolved, _ := filepath.EvalSymlinks(tmpDir)
		gotResolved, _ := filepath.EvalSymlinks(dir)
		if gotResolved != wantResolved {
			t.Errorf("expected %q, got %q", wantResolved, gotResolved)
		}
	})

	t.Run("not found", func(t *testing.T) {
		orgDir = ""
		t.Setenv("UNIMART_ORG_DIR", "")

		outside := t.TempDir()
		t.Chdir(outside)

		if _, err := resolveOrgDir(); err == nil {
			t.Error("expected error when no org dir markers exist")
		}
	})
}

func TestIsOrgDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Not an org dir — missing .gitmodules
	if isOrgDir(tmpDir) {
		t.Error("empty dir should not be detected as org dir")
	}

	// A partial marker set is not sufficient.
	os.WriteFile(filepath.Join(tmpDir, ".gitmodules"), []byte("[submodule]"), 0o644)
	if isOrgDir(tmpDir) {
		t.Error("dir with only .gitmodules should not be detected as org dir")
	}

	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example"), 0o644)
	os.WriteFile(filepath.Join(tmpDir, "flake.nix"), []byte("{}"), 0o644)
	if !isOrgDir(tmpDir) {
		t.Error("dir with all meta repository markers should be detected as org dir")
	}
}
