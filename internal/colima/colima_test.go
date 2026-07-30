package colima

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.CPU != 4 {
		t.Errorf("DefaultConfig().CPU = %d, want 4", cfg.CPU)
	}
	if cfg.Memory != 8 {
		t.Errorf("DefaultConfig().Memory = %d, want 8", cfg.Memory)
	}
	if cfg.Disk != 60 {
		t.Errorf("DefaultConfig().Disk = %d, want 60", cfg.Disk)
	}
}

func TestSocketPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("could not get home dir: %v", err)
	}

	want := filepath.Join(home, ".config", "colima", "default", "docker.sock")
	got := SocketPath()
	if got != want {
		t.Errorf("SocketPath() = %q, want %q", got, want)
	}
}

func TestEnsureDockerHost_AlreadySet(t *testing.T) {
	// Create a temp socket file so Stat doesn't fail
	tmpDir := t.TempDir()
	sockFile := filepath.Join(tmpDir, "docker.sock")
	if err := os.WriteFile(sockFile, nil, 0600); err != nil {
		t.Fatalf("failed to create temp socket: %v", err)
	}

	// Save and restore env
	origHost := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", origHost)

	// When DOCKER_HOST already equals the expected value, ensureDockerHost
	// should be a no-op. We can't easily test this without mocking SocketPath,
	// but we can test the exported wrapper doesn't panic.
	expected := "unix://" + SocketPath()
	os.Setenv("DOCKER_HOST", expected)

	// This will either succeed (socket exists) or return an error (socket doesn't exist)
	// Either way it should not panic
	_ = EnsureDockerHost()

	// DOCKER_HOST should still be the same
	if got := os.Getenv("DOCKER_HOST"); got != expected {
		t.Errorf("DOCKER_HOST changed unexpectedly: got %q, want %q", got, expected)
	}
}

func TestEnsureDockerHost_SocketMissing(t *testing.T) {
	// Save and restore env
	origHost := os.Getenv("DOCKER_HOST")
	defer os.Setenv("DOCKER_HOST", origHost)

	// Set DOCKER_HOST to something wrong so ensureDockerHost tries to fix it
	os.Setenv("DOCKER_HOST", "unix:///nonexistent/path.sock")

	// EnsureDockerHost should return an error if the Colima socket doesn't exist
	// (unless Colima is actually running, in which case the socket does exist)
	err := EnsureDockerHost()

	sock := SocketPath()
	if _, statErr := os.Stat(sock); statErr != nil {
		// Socket doesn't exist — should get an error
		if err == nil {
			t.Error("EnsureDockerHost() should return error when socket doesn't exist")
		}
	} else {
		// Socket exists (Colima running) — should succeed
		if err != nil {
			t.Errorf("EnsureDockerHost() returned error with existing socket: %v", err)
		}
		// Verify DOCKER_HOST was updated
		expected := "unix://" + sock
		if got := os.Getenv("DOCKER_HOST"); got != expected {
			t.Errorf("DOCKER_HOST = %q, want %q", got, expected)
		}
	}
}
