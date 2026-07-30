package prereqs

import "testing"

func TestCheckPodman_NotInstalled(t *testing.T) {
	// This test simply runs CheckPodman to ensure it returns a sensible result
	res := CheckPodman()
	if res.Name != "podman" {
		t.Fatalf("unexpected name: %s", res.Name)
	}
}
