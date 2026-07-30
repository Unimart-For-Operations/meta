package prereqs

import (
	"os"
	"path/filepath"
	"testing"
)

// --- CheckWorkspace tests (org directory only) ---

func TestCheckWorkspace_ValidOrg(t *testing.T) {
	tmpDir := t.TempDir()

	results := CheckWorkspace(tmpDir)
	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}

	// Org directory should pass
	if results[0].Status != StatusPass {
		t.Errorf("org directory check: got %v, want pass", results[0].Status)
	}
}

func TestCheckWorkspace_MissingOrg(t *testing.T) {
	results := CheckWorkspace("/nonexistent/path/12345")
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}
	if results[0].Status != StatusFail {
		t.Errorf("missing org dir: got %v, want fail", results[0].Status)
	}
}

// --- CheckWorkspaceIdpbuilder tests ---

func TestCheckWorkspaceIdpbuilder_RepoExists(t *testing.T) {
	tmpDir := t.TempDir()
	idpDir := filepath.Join(tmpDir, "idpbuilder")
	if err := os.Mkdir(idpDir, 0755); err != nil {
		t.Fatalf("failed to create idpbuilder dir: %v", err)
	}

	// Create a Makefile to simulate valid checkout
	makefile := filepath.Join(idpDir, "Makefile")
	if err := os.WriteFile(makefile, []byte("all:\n\techo ok\n"), 0644); err != nil {
		t.Fatalf("failed to create Makefile: %v", err)
	}

	results := CheckWorkspaceIdpbuilder(tmpDir)
	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}

	// idpbuilder repo should pass
	if results[0].Status != StatusPass {
		t.Errorf("idpbuilder repo check: got %v, want pass", results[0].Status)
	}
}

func TestCheckWorkspaceIdpbuilder_MissingRepo(t *testing.T) {
	tmpDir := t.TempDir()

	results := CheckWorkspaceIdpbuilder(tmpDir)
	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}

	// idpbuilder repo should fail
	if results[0].Status != StatusFail {
		t.Errorf("missing idpbuilder repo: got %v, want fail", results[0].Status)
	}
}

func TestCheckWorkspaceIdpbuilder_NoBinary(t *testing.T) {
	tmpDir := t.TempDir()
	idpDir := filepath.Join(tmpDir, "idpbuilder")
	if err := os.Mkdir(idpDir, 0755); err != nil {
		t.Fatalf("failed to create idpbuilder dir: %v", err)
	}
	// Create Makefile but no binary
	if err := os.WriteFile(filepath.Join(idpDir, "Makefile"), []byte("all:"), 0644); err != nil {
		t.Fatalf("failed to create Makefile: %v", err)
	}

	results := CheckWorkspaceIdpbuilder(tmpDir)

	// Find the binary check result
	var binaryResult *CheckResult
	for i := range results {
		if results[i].Name == "idpbuilder binary" {
			binaryResult = &results[i]
			break
		}
	}

	if binaryResult == nil {
		t.Fatal("expected a result for idpbuilder binary")
	}
	if binaryResult.Status != StatusWarn {
		t.Errorf("missing binary: got %v, want warn", binaryResult.Status)
	}
}

func TestCheckWorkspaceIdpbuilder_WithBinary(t *testing.T) {
	tmpDir := t.TempDir()
	idpDir := filepath.Join(tmpDir, "idpbuilder")
	if err := os.Mkdir(idpDir, 0755); err != nil {
		t.Fatalf("failed to create idpbuilder dir: %v", err)
	}
	// Create Makefile and binary
	if err := os.WriteFile(filepath.Join(idpDir, "Makefile"), []byte("all:"), 0644); err != nil {
		t.Fatalf("failed to create Makefile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(idpDir, "idpbuilder"), []byte("binary"), 0755); err != nil {
		t.Fatalf("failed to create binary: %v", err)
	}

	results := CheckWorkspaceIdpbuilder(tmpDir)

	// Find the binary check result
	var binaryResult *CheckResult
	for i := range results {
		if results[i].Name == "idpbuilder binary" {
			binaryResult = &results[i]
			break
		}
	}

	if binaryResult == nil {
		t.Fatal("expected a result for idpbuilder binary")
	}
	if binaryResult.Status != StatusPass {
		t.Errorf("present binary: got %v, want pass", binaryResult.Status)
	}
}
