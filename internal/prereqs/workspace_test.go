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

	results := CheckWorkspaceIdpbuilder(tmpDir)
	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}

	// idpbuilder source should pass
	if results[0].Status != StatusPass {
		t.Errorf("idpbuilder source check: got %v, want pass", results[0].Status)
	}
}

func TestCheckWorkspaceIdpbuilder_MissingRepo(t *testing.T) {
	tmpDir := t.TempDir()

	results := CheckWorkspaceIdpbuilder(tmpDir)
	if len(results) < 1 {
		t.Fatalf("expected at least 1 result, got %d", len(results))
	}

	// idpbuilder source should fail
	if results[0].Status != StatusFail {
		t.Errorf("missing idpbuilder source: got %v, want fail", results[0].Status)
	}
}
