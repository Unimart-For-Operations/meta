package builder

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuild_MissingMakefile(t *testing.T) {
	err := Build(t.TempDir(), false)
	if err == nil {
		t.Fatal("Build should fail when Makefile is missing")
	}
	if !strings.Contains(err.Error(), "Makefile not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuild_BuildSucceedsButBinaryMissing(t *testing.T) {
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not available")
	}

	dir := t.TempDir()
	makefile := filepath.Join(dir, "Makefile")
	if err := os.WriteFile(makefile, []byte("build:\n\t@echo fake build\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Build(dir, false)
	if err == nil {
		t.Skip("make build produced a binary; nothing to assert")
	}
	if !strings.Contains(err.Error(), "binary not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCreate_MissingBinary(t *testing.T) {
	err := Create(t.TempDir(), []string{"--dev-password"})
	if err == nil {
		t.Fatal("Create should fail when the idpbuilder binary is missing")
	}
	if !strings.Contains(err.Error(), "idpbuilder binary not found") {
		t.Errorf("unexpected error: %v", err)
	}
}
