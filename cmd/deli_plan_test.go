package cmd

import (
	"testing"

	"github.com/Unimart-For-Operations/meta/internal/host"
)

func TestBuildHostPlanBaseline(t *testing.T) {
	plan := buildHostPlan(host.Info{
		Name:         "tty-box",
		Platform:     "arch",
		Username:     "alice",
		Role:         "tty-engineer",
		Capabilities: []string{"baseline", "terminal-dev"},
	})

	if len(plan.Steps) != 4 {
		t.Fatalf("expected 4 baseline steps, got %d", len(plan.Steps))
	}
	if hasStep(plan, "Phase 5 — Stand Up Local IDP") {
		t.Fatal("baseline plan should not include local IDP phase")
	}
}

func TestBuildHostPlanCapabilityPhases(t *testing.T) {
	plan := buildHostPlan(host.Info{
		Name:         "battle-unit",
		Platform:     "arch",
		Username:     "alice",
		Role:         "platform-operator",
		Capabilities: []string{"baseline", "operator", "idp-local", "desktop"},
	})

	for _, title := range []string{
		"Phase 4 — Validate Operator Tooling",
		"Phase 5 — Stand Up Local IDP",
		"Desktop Check",
	} {
		if !hasStep(plan, title) {
			t.Fatalf("expected plan step %q", title)
		}
	}
}

func TestHasCapabilityCaseInsensitive(t *testing.T) {
	h := host.Info{Capabilities: []string{"IDP-Local"}}
	if !hasCapability(h, "idp-local") {
		t.Fatal("expected case-insensitive capability match")
	}
}

func hasStep(plan hostPlan, title string) bool {
	for _, step := range plan.Steps {
		if step.Title == title {
			return true
		}
	}
	return false
}
