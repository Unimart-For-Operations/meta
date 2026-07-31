package platform

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestIsDarwinIsLinux(t *testing.T) {
	if IsDarwin() == IsLinux() {
		t.Errorf("IsDarwin() and IsLinux() should be mutually exclusive, got %v/%v", IsDarwin(), IsLinux())
	}
	if IsDarwin() && runtime.GOOS != "darwin" {
		t.Errorf("IsDarwin() = true but runtime.GOOS = %q", runtime.GOOS)
	}
	if IsLinux() && runtime.GOOS != "linux" {
		t.Errorf("IsLinux() = true but runtime.GOOS = %q", runtime.GOOS)
	}
}

func TestArch(t *testing.T) {
	if got := Arch(); got != runtime.GOARCH {
		t.Errorf("Arch() = %q, want %q", got, runtime.GOARCH)
	}
}

func TestCommandExists(t *testing.T) {
	if !CommandExists("sh") {
		t.Error("CommandExists(sh) = false, want true")
	}
	if CommandExists("definitely-not-a-real-command-xyz") {
		t.Error("CommandExists(bogus) = true, want false")
	}
}

func TestCommandOutput(t *testing.T) {
	out, err := CommandOutput("echo", "hello world")
	if err != nil {
		t.Fatalf("CommandOutput: %v", err)
	}
	if out != "hello world" {
		t.Errorf("CommandOutput = %q, want %q", out, "hello world")
	}
}

func TestCommandOutput_Trimmed(t *testing.T) {
	out, err := CommandOutput("printf", "  spaced  \n")
	if err != nil {
		t.Fatalf("CommandOutput: %v", err)
	}
	if out != "spaced" {
		t.Errorf("CommandOutput = %q, want %q", out, "spaced")
	}
}

func TestCommandOutput_MissingCommand(t *testing.T) {
	_, err := CommandOutput("definitely-not-a-real-command-xyz")
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestCommandOutputSilent(t *testing.T) {
	out, err := CommandOutputSilent("echo", "quiet")
	if err != nil {
		t.Fatalf("CommandOutputSilent: %v", err)
	}
	if out != "quiet" {
		t.Errorf("CommandOutputSilent = %q, want %q", out, "quiet")
	}
}

func TestRunSilent(t *testing.T) {
	if err := RunSilent("true"); err != nil {
		t.Errorf("RunSilent(true) = %v, want nil", err)
	}
	if err := RunSilent("false"); err == nil {
		t.Error("RunSilent(false) = nil, want error")
	}
}

func TestRunVisible(t *testing.T) {
	if err := RunVisible("true"); err != nil {
		t.Errorf("RunVisible(true) = %v, want nil", err)
	}
}

func TestRunVisibleDir(t *testing.T) {
	if err := RunVisibleDir(t.TempDir(), "true"); err != nil {
		t.Errorf("RunVisibleDir(true) = %v, want nil", err)
	}
}

func TestOSVersion(t *testing.T) {
	v := OSVersion()
	if v == "" {
		t.Fatal("OSVersion() returned empty string")
	}
	if runtime.GOOS == "linux" {
		want := linuxPrettyName(t)
		if v != want {
			t.Errorf("OSVersion() = %q, want %q", v, want)
		}
	}
}

func linuxPrettyName(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		t.Skip("no /etc/os-release")
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}
	return "Linux"
}
