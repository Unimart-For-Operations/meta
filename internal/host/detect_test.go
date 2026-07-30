package host

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func testPlatform() string {
	if runtime.GOOS == "darwin" {
		return "macos"
	}
	return "arch" // first Linux search dir
}

func TestDetectForUser(t *testing.T) {
	// Create a temporary directory structure mimicking cmdr/home/02-hosts/
	// Use a platform that matches the current OS so DetectForUser finds it.
	platform := testPlatform()
	tmpDir := t.TempDir()
	cmdrDir := filepath.Join(tmpDir, "cmdr", "home", "02-hosts", platform, "test-host")
	if err := os.MkdirAll(cmdrDir, 0o755); err != nil {
		t.Fatal(err)
	}

	metaNix := filepath.Join(cmdrDir, "meta.nix")
	content := `{
  username = "testuser";
  role = "tty-engineer";
  capabilities = [ "baseline" "terminal-dev" ];
  hostname = "test-host";
}`
	if err := os.WriteFile(metaNix, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	// Also need cmdr/flake.nix for org detection
	flakeNix := filepath.Join(tmpDir, "cmdr", "flake.nix")
	if err := os.WriteFile(flakeNix, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("match", func(t *testing.T) {
		info, err := DetectForUser(tmpDir, "testuser")
		if err != nil {
			t.Fatalf("expected match, got error: %v", err)
		}
		if info.Name != "test-host" {
			t.Errorf("expected name 'test-host', got %q", info.Name)
		}
		if info.Platform != platform {
			t.Errorf("expected platform %q, got %q", platform, info.Platform)
		}
		if info.Username != "testuser" {
			t.Errorf("expected username 'testuser', got %q", info.Username)
		}
		if info.Role != "tty-engineer" {
			t.Errorf("expected role 'tty-engineer', got %q", info.Role)
		}
		if got, want := info.Capabilities, []string{"baseline", "terminal-dev"}; !sameStrings(got, want) {
			t.Errorf("expected capabilities %v, got %v", want, got)
		}
	})

	t.Run("no match", func(t *testing.T) {
		_, err := DetectForUser(tmpDir, "nobody")
		if err == nil {
			t.Fatal("expected error for non-existent user, got nil")
		}
	})
}

func TestListHosts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two hosts on different platforms
	for _, tc := range []struct {
		platform string
		host     string
		username string
	}{
		{"macos", "macbook", "alice"},
		{"arch", "desktop", "bob"},
	} {
		hostDir := filepath.Join(tmpDir, "cmdr", "home", "02-hosts", tc.platform, tc.host)
		if err := os.MkdirAll(hostDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := `{ username = "` + tc.username + `"; role = "developer-workstation"; capabilities = [ "baseline" "operator" ]; }`
		if err := os.WriteFile(filepath.Join(hostDir, "meta.nix"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	hosts, err := ListHosts(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(hosts))
	}

	// Hosts are returned in platform order (arch before macos)
	found := map[string]bool{}
	for _, h := range hosts {
		found[h.Name] = true
	}
	if !found["macbook"] || !found["desktop"] {
		t.Errorf("expected macbook and desktop, got %v", hosts)
	}
}

func TestGetHost(t *testing.T) {
	tmpDir := t.TempDir()
	hostDir := filepath.Join(tmpDir, "cmdr", "home", "02-hosts", "arch", "desktop")
	if err := os.MkdirAll(hostDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := `{ username = "bob"; role = "developer-workstation"; capabilities = [ "baseline" ]; }`
	if err := os.WriteFile(filepath.Join(hostDir, "meta.nix"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	h, err := GetHost(tmpDir, "desktop")
	if err != nil {
		t.Fatalf("GetHost: %v", err)
	}
	if h.Name != "desktop" || h.Role != "developer-workstation" {
		t.Fatalf("unexpected host: %+v", h)
	}

	if _, err := GetHost(tmpDir, "missing"); err == nil {
		t.Fatal("expected missing host error")
	}
}

func TestParseMetaFields(t *testing.T) {
	data := []byte(`{
  description = "Example";
  username = "alice";
  role = "platform-operator";
  capabilities = [
    "baseline"
    "terminal-dev"
    "idp-local"
  ];
}`)

	fields := parseMetaFields(data)
	if fields.Strings["username"] != "alice" {
		t.Fatalf("username = %q, want alice", fields.Strings["username"])
	}
	if fields.Strings["role"] != "platform-operator" {
		t.Fatalf("role = %q, want platform-operator", fields.Strings["role"])
	}
	if got, want := fields.Lists["capabilities"], []string{"baseline", "terminal-dev", "idp-local"}; !sameStrings(got, want) {
		t.Fatalf("capabilities = %v, want %v", got, want)
	}
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
