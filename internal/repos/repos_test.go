package repos

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
}

// makeRepo creates a git repository at dir, optionally with an untracked file.
func makeRepo(t *testing.T, dir string, dirty bool) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "init", "-q", "-b", "main")
	if dirty {
		if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRepositoriesDir(t *testing.T) {
	org := filepath.Join(t.TempDir(), "org")
	got := RepositoriesDir(org)
	want := filepath.Join(org, "repositories")
	if got != want {
		t.Errorf("RepositoriesDir(%q) = %q, want %q", org, got, want)
	}
}

func TestListLocal(t *testing.T) {
	orgDir := t.TempDir()
	makeRepo(t, filepath.Join(orgDir, "clean-repo"), false)
	makeRepo(t, filepath.Join(orgDir, "dirty-repo"), true)
	if err := os.MkdirAll(filepath.Join(orgDir, "not-a-git-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orgDir, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	repos, err := ListLocal(orgDir)
	if err != nil {
		t.Fatalf("ListLocal: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}

	var cleanRepo, dirtyRepo *LocalRepo
	for i := range repos {
		switch repos[i].Name {
		case "clean-repo":
			cleanRepo = &repos[i]
		case "dirty-repo":
			dirtyRepo = &repos[i]
		}
	}

	if cleanRepo == nil || dirtyRepo == nil {
		t.Fatalf("repos = %v, want clean-repo and dirty-repo", repos)
	}

	if cleanRepo.Branch != "main" {
		t.Errorf("clean-repo.Branch = %q, want main", cleanRepo.Branch)
	}
	if !cleanRepo.Clean {
		t.Error("clean-repo.Clean = false, want true")
	}
	if dirtyRepo.Clean {
		t.Error("dirty-repo.Clean = true, want false")
	}
	if filepath.Base(cleanRepo.Path) != "clean-repo" {
		t.Errorf("clean-repo.Path = %q", cleanRepo.Path)
	}
}

func TestListLocal_MissingDir(t *testing.T) {
	_, err := ListLocal(filepath.Join(t.TempDir(), "nope"))
	if err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestListPublishable_WithRepositoriesDir(t *testing.T) {
	orgDir := t.TempDir()
	makeRepo(t, filepath.Join(orgDir, "repositories", "repo-a"), false)
	// A repo in org root should NOT be discovered when repositories/ exists.
	makeRepo(t, filepath.Join(orgDir, "stray"), false)

	repos, source, err := ListPublishable(orgDir)
	if err != nil {
		t.Fatalf("ListPublishable: %v", err)
	}
	if source != RepositoriesDir(orgDir) {
		t.Errorf("source = %q, want %q", source, RepositoriesDir(orgDir))
	}
	if len(repos) != 1 || repos[0].Name != "repo-a" {
		t.Errorf("repos = %v, want [repo-a]", repos)
	}
}

func TestListPublishable_FallbackToOrgDir(t *testing.T) {
	orgDir := t.TempDir()
	makeRepo(t, filepath.Join(orgDir, "repo-b"), false)

	repos, source, err := ListPublishable(orgDir)
	if err != nil {
		t.Fatalf("ListPublishable: %v", err)
	}
	if source != orgDir {
		t.Errorf("source = %q, want %q", source, orgDir)
	}
	if len(repos) != 1 || repos[0].Name != "repo-b" {
		t.Errorf("repos = %v, want [repo-b]", repos)
	}
}

func TestListPublishable_RepositoriesDirIsFile(t *testing.T) {
	orgDir := t.TempDir()
	file := RepositoriesDir(orgDir)
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := ListPublishable(orgDir)
	if err == nil {
		t.Fatal("expected error when repositories is a file")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClone_AlreadyExists(t *testing.T) {
	orgDir := t.TempDir()
	dest := filepath.Join(orgDir, "existing")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}

	err := Clone(orgDir, "Unimart-For-Operations", "existing")
	if err == nil {
		t.Fatal("expected error for existing destination")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGitOutput(t *testing.T) {
	dir := t.TempDir()
	makeRepo(t, dir, false)

	branch, err := gitOutput(dir, "branch", "--show-current")
	if err != nil {
		t.Fatalf("gitOutput(branch): %v", err)
	}
	if branch != "main" {
		t.Errorf("branch = %q, want main", branch)
	}

	status, err := gitOutput(dir, "status", "--porcelain")
	if err != nil {
		t.Fatalf("gitOutput(status): %v", err)
	}
	if status != "" {
		t.Errorf("status = %q, want empty for clean repo", status)
	}
}
