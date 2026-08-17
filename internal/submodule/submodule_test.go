package submodule

import (
	"os"
	"path/filepath"
	"testing"
)

// gitmodulesContent is a realistic .gitmodules matching the org repo.
const gitmodulesContent = `[submodule "cmdr"]
	path = cmdr
	url = git@github.com:Unimart-For-Operations/cmdr.git
	ignore = dirty
[submodule "idpbuilder"]
	path = idpbuilder
	url = git@github.com:Unimart-For-Operations/idpbuilder.git
	ignore = dirty
[submodule "idpctl"]
	path = idpctl
	url = git@github.com:Unimart-For-Operations/idpctl.git
	ignore = dirty
[submodule "docs"]
	path = docs
	url = git@github.com:Unimart-For-Operations/docs.git
	ignore = dirty
[submodule "unimart-employee-handbooks/cdc"]
	path = unimart-employee-handbooks/cdc
	url = git@github.com:idpbuilder/cdc.git
	ignore = dirty
`

func TestParseGitmodules(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitmodules"), []byte(gitmodulesContent), 0o644); err != nil {
		t.Fatal(err)
	}

	subs, err := ParseGitmodules(tmpDir)
	if err != nil {
		t.Fatalf("ParseGitmodules: %v", err)
	}

	if len(subs) != 5 {
		t.Fatalf("expected 5 submodules, got %d", len(subs))
	}

	// Verify each entry
	expected := []struct {
		name, path, url string
	}{
		{"cmdr", "cmdr", "git@github.com:Unimart-For-Operations/cmdr.git"},
		{"idpbuilder", "idpbuilder", "git@github.com:Unimart-For-Operations/idpbuilder.git"},
		{"idpctl", "idpctl", "git@github.com:Unimart-For-Operations/idpctl.git"},
		{"docs", "docs", "git@github.com:Unimart-For-Operations/docs.git"},
		{"unimart-employee-handbooks/cdc", "unimart-employee-handbooks/cdc", "git@github.com:idpbuilder/cdc.git"},
	}

	for i, e := range expected {
		if subs[i].Name != e.name {
			t.Errorf("sub[%d].Name = %q, want %q", i, subs[i].Name, e.name)
		}
		if subs[i].Path != e.path {
			t.Errorf("sub[%d].Path = %q, want %q", i, subs[i].Path, e.path)
		}
		if subs[i].URL != e.url {
			t.Errorf("sub[%d].URL = %q, want %q", i, subs[i].URL, e.url)
		}
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"cmdr", "cmdr"},
		{"idpbuilder", "idpbuilder"},
		{"unimart-employee-handbooks/cdc", "cdc"},
		{"deep/nested/path/vault", "vault"},
	}

	for _, tc := range tests {
		s := Submodule{Path: tc.path}
		if got := s.DisplayName(); got != tc.want {
			t.Errorf("DisplayName(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestMaxDisplayWidth(t *testing.T) {
	subs := []Submodule{
		{Path: "cmdr"},
		{Path: "idpbuilder"},
		{Path: "unimart-employee-handbooks/cdc"},
	}

	got := MaxDisplayWidth(subs)
	// "idpbuilder" is 10 chars, longest among "cmdr"(4), "idpbuilder"(10), "cdc"(3)
	if got != 10 {
		t.Errorf("MaxDisplayWidth = %d, want 10", got)
	}
}

func TestIsSourceModule(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a submodule dir with docs/
	withDocs := filepath.Join(tmpDir, "cmdr", "docs")
	if err := os.MkdirAll(withDocs, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a submodule dir without docs/
	withoutDocs := filepath.Join(tmpDir, "unimart-employee-handbooks", "cdc")
	if err := os.MkdirAll(withoutDocs, 0o755); err != nil {
		t.Fatal(err)
	}

	t.Run("has docs", func(t *testing.T) {
		s := Submodule{Path: "cmdr"}
		if !s.IsSourceModule(tmpDir) {
			t.Error("expected IsSourceModule=true for cmdr (has docs/)")
		}
	})

	t.Run("no docs", func(t *testing.T) {
		s := Submodule{Path: "unimart-employee-handbooks/cdc"}
		if s.IsSourceModule(tmpDir) {
			t.Error("expected IsSourceModule=false for cdc (no docs/)")
		}
	})
}

func TestPaths(t *testing.T) {
	subs := []Submodule{
		{Path: "cmdr"},
		{Path: "unimart-employee-handbooks/cdc"},
	}
	got := Paths(subs)
	if len(got) != 2 || got[0] != "cmdr" || got[1] != "unimart-employee-handbooks/cdc" {
		t.Errorf("Paths = %v, want [cmdr unimart-employee-handbooks/cdc]", got)
	}
}

func TestParseGitmodules_Missing(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := ParseGitmodules(tmpDir)
	if err == nil {
		t.Error("expected error for missing .gitmodules, got nil")
	}
}

func TestParseGitmodules_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitmodules"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	subs, err := ParseGitmodules(tmpDir)
	if err != nil {
		t.Fatalf("ParseGitmodules: %v", err)
	}
	if len(subs) != 0 {
		t.Errorf("expected 0 submodules from empty file, got %d", len(subs))
	}
}
