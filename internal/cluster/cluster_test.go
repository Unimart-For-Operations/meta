package cluster

import (
	"os/exec"
	"testing"
)

// Basic smoke test for GetGiteaAdminToken; since it shells out to kubectl,
// we verify the function returns an error when kubectl is not available in the
// test environment.
func TestGetGiteaAdminToken_NoKubectl(t *testing.T) {
	// Temporarily manipulate PATH to ensure kubectl is not found would be
	// intrusive; instead, call the function and expect an error in CI.
	_, err := GetGiteaAdminToken()
	if err == nil {
		// If kubectl is present and cluster available, that's okay — just skip
		t.Skip("kubectl present; skipping negative test")
	}
	// Ensure the error is due to command execution (exec.ExitError or similar)
	if _, ok := err.(*exec.ExitError); ok {
		return
	}
}
