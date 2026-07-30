package prereqs

import (
	"runtime"
	"testing"
)

func TestPlatform(t *testing.T) {
	got := Platform()
	if got != runtime.GOOS {
		t.Errorf("Platform() = %q, want %q", got, runtime.GOOS)
	}
}

func TestArch(t *testing.T) {
	got := Arch()
	if got != runtime.GOARCH {
		t.Errorf("Arch() = %q, want %q", got, runtime.GOARCH)
	}
}

func TestIsDarwin(t *testing.T) {
	want := runtime.GOOS == "darwin"
	if got := IsDarwin(); got != want {
		t.Errorf("IsDarwin() = %v, want %v", got, want)
	}
}

func TestCommandExists(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"go should exist", "go", true},
		{"echo should exist", "echo", true},
		{"nonexistent command", "this-command-does-not-exist-12345", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CommandExists(tt.cmd); got != tt.want {
				t.Errorf("CommandExists(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestCommandPath(t *testing.T) {
	// A command that exists should return a non-empty path
	path := CommandPath("go")
	if path == "" {
		t.Error("CommandPath(\"go\") returned empty string, expected a path")
	}

	// A nonexistent command should return empty
	path = CommandPath("this-command-does-not-exist-12345")
	if path != "" {
		t.Errorf("CommandPath for nonexistent command returned %q, want empty", path)
	}
}

func TestCommandOutput(t *testing.T) {
	// echo writes to stdout — should be captured
	out, err := CommandOutput("echo", "hello world")
	if err != nil {
		t.Fatalf("CommandOutput(echo) error: %v", err)
	}
	if out != "hello world" {
		t.Errorf("CommandOutput(echo) = %q, want %q", out, "hello world")
	}
}

func TestCommandOutputCombined(t *testing.T) {
	// Use a shell command that writes to stderr
	out, err := CommandOutputCombined("bash", "-c", "echo stdout-msg; echo stderr-msg >&2")
	if err != nil {
		t.Fatalf("CommandOutputCombined error: %v", err)
	}
	// Both stdout and stderr should appear in output
	if out == "" {
		t.Error("CommandOutputCombined returned empty output")
	}
	// Should contain both messages
	if !containsStr(out, "stdout-msg") {
		t.Errorf("output missing stdout-msg: %q", out)
	}
	if !containsStr(out, "stderr-msg") {
		t.Errorf("output missing stderr-msg: %q", out)
	}
}

func TestCommandOutput_Error(t *testing.T) {
	_, err := CommandOutput("this-command-does-not-exist-12345")
	if err == nil {
		t.Error("CommandOutput with nonexistent command should return error")
	}
}

func TestCommandOutputCombined_Error(t *testing.T) {
	// Command that exits non-zero
	_, err := CommandOutputCombined("bash", "-c", "exit 1")
	if err == nil {
		t.Error("CommandOutputCombined with failing command should return error")
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{StatusPass, "pass"},
		{StatusFail, "fail"},
		{StatusWarn, "warn"},
		{Status(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.status.String(); got != tt.want {
				t.Errorf("Status(%d).String() = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestHasNix(t *testing.T) {
	// On the test machine nix should be available (nix-managed system)
	// This is a reasonable assumption for this project's CI
	if got := HasNix(); !got {
		t.Log("HasNix() returned false — nix not on PATH in test environment")
	}
}

func TestRunCommand(t *testing.T) {
	// Successful command
	if err := RunCommand("echo", "test"); err != nil {
		t.Errorf("RunCommand(echo) should succeed, got: %v", err)
	}

	// Failing command
	if err := RunCommand("bash", "-c", "exit 1"); err == nil {
		t.Error("RunCommand with exit 1 should return error")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
