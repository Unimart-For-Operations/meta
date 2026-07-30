package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHumanAge(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"just now", now.Add(-30 * time.Second), "just now"},
		{"minutes", now.Add(-30 * time.Minute), "30 minutes ago"},
		{"hours", now.Add(-5 * time.Hour), "5 hours ago"},
		{"days", now.Add(-3 * 24 * time.Hour), "3 days ago"},
		{"weeks", now.Add(-21 * 24 * time.Hour), "3 weeks ago"},
		{"months", now.Add(-90 * 24 * time.Hour), "3 months ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := humanAge(tt.t); got != tt.want {
				t.Errorf("humanAge() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseMakefileTargets(t *testing.T) {
	dir := t.TempDir()
	makefile := filepath.Join(dir, "Makefile")
	content := `# comment line
CYAN := \033[0;36m
.DEFAULT_GOAL := help

help: ## Show help
	@echo help

hooks: deps ## Install hooks
	@echo hooks

VAR = value
$(BINARY): main.go
	go build

.PHONY: help hooks
`
	if err := os.WriteFile(makefile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	targets, err := parseMakefileTargets(makefile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"help", "hooks"} {
		if !targets[want] {
			t.Errorf("expected target %q to be found", want)
		}
	}

	for _, notWant := range []string{"CYAN", "VAR", "$(BINARY)", ".PHONY", ".DEFAULT_GOAL"} {
		if targets[notWant] {
			t.Errorf("did not expect %q to be a target", notWant)
		}
	}
}

func TestParseMakefileTargetsMissingFile(t *testing.T) {
	if _, err := parseMakefileTargets(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("expected error for missing file")
	}
}

func TestDisplayValue(t *testing.T) {
	if got := displayValue(""); got != "-" {
		t.Errorf("displayValue(\"\") = %q, want \"-\"", got)
	}
	if got := displayValue("operator"); got != "operator" {
		t.Errorf("displayValue(\"operator\") = %q", got)
	}
}

func TestDisplayList(t *testing.T) {
	if got := displayList(nil); got != "-" {
		t.Errorf("displayList(nil) = %q, want \"-\"", got)
	}
	if got := displayList([]string{"a", "b"}); got != "a,b" {
		t.Errorf("displayList = %q, want \"a,b\"", got)
	}
}

func TestHasPackages(t *testing.T) {
	t.Run("missing dir", func(t *testing.T) {
		if hasPackages(filepath.Join(t.TempDir(), "nope")) {
			t.Error("expected false for missing dir")
		}
	})

	t.Run("empty dir", func(t *testing.T) {
		if hasPackages(t.TempDir()) {
			t.Error("expected false for empty dir")
		}
	})

	t.Run("only gitkeep", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, ".gitkeep"), nil, 0o644)
		if hasPackages(dir) {
			t.Error("expected false for dir with only .gitkeep")
		}
	})

	t.Run("only subdirs", func(t *testing.T) {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
		if hasPackages(dir) {
			t.Error("expected false for dir with only subdirectories")
		}
	})

	t.Run("with package file", func(t *testing.T) {
		dir := t.TempDir()
		os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("kind: Application"), 0o644)
		if !hasPackages(dir) {
			t.Error("expected true for dir with a yaml file")
		}
	})
}
