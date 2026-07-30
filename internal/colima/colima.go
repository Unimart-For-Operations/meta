package colima

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds Colima VM resource settings.
type Config struct {
	CPU    int
	Memory int
	Disk   int
}

// DefaultConfig returns the recommended Colima settings for running Kind + IDP.
func DefaultConfig() Config {
	return Config{
		CPU:    4,
		Memory: 8,
		Disk:   60,
	}
}

// SocketPath returns the path to Colima's Docker socket.
func SocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "colima", "default", "docker.sock")
}

// IsRunning checks if Colima is currently running.
// Note: colima status writes structured log output to stderr, not stdout,
// so we must use CombinedOutput() to capture it.
func IsRunning() bool {
	out, err := exec.Command("colima", "status").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "running")
}

// Start starts Colima with the given resource configuration.
// After starting, it ensures DOCKER_HOST points to Colima's socket
// so that all child processes (docker, kind, idpbuilder) connect correctly.
func Start(cfg Config) error {
	if IsRunning() {
		fmt.Println("  Colima is already running")
		return ensureDockerHost()
	}

	fmt.Printf("  Starting Colima (cpu=%d, memory=%dGB, disk=%dGB)...\n", cfg.CPU, cfg.Memory, cfg.Disk)

	cmd := exec.Command("colima", "start",
		"--cpu", strconv.Itoa(cfg.CPU),
		"--memory", strconv.Itoa(cfg.Memory),
		"--disk", strconv.Itoa(cfg.Disk),
		"--runtime", "docker",
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("starting Colima: %w", err)
	}

	fmt.Println("  Colima started")
	return ensureDockerHost()
}

// ensureDockerHost sets DOCKER_HOST to Colima's socket for this process and
// all child processes. This is necessary because:
//   - DOCKER_HOST env var overrides Docker CLI contexts
//   - The user's shell may have DOCKER_HOST pointing elsewhere (e.g. Docker Desktop)
//   - Child processes (kind, idpbuilder) inherit the env and need the right socket
func ensureDockerHost() error {
	sock := SocketPath()
	expected := "unix://" + sock

	if os.Getenv("DOCKER_HOST") == expected {
		return nil
	}

	// Verify socket exists
	if _, err := os.Stat(sock); err != nil {
		return fmt.Errorf("Colima socket not found at %s — is Colima running?", sock)
	}

	fmt.Printf("  Setting DOCKER_HOST=%s\n", expected)
	os.Setenv("DOCKER_HOST", expected)

	// Also switch Docker context as a convenience for post-unimart commands
	_ = exec.Command("docker", "context", "use", "colima").Run()

	return nil
}

// Stop stops Colima.
func Stop() error {
	if !IsRunning() {
		fmt.Println("  Colima is not running")
		return nil
	}

	fmt.Println("  Stopping Colima...")
	cmd := exec.Command("colima", "stop")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("stopping Colima: %w", err)
	}

	fmt.Println("  Colima stopped")
	return nil
}

// Status returns a human-readable Colima status string.
func Status() string {
	out, err := exec.Command("colima", "status").CombinedOutput()
	if err != nil {
		return "not running"
	}
	return strings.TrimSpace(string(out))
}

// EnsureDockerHost sets DOCKER_HOST to Colima's socket for this process
// and all child processes. Exported so that commands like `status` can
// call it before shelling out to docker/kind.
func EnsureDockerHost() error {
	return ensureDockerHost()
}
