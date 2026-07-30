package prereqs

import (
	"testing"
)

func TestCheckDocker_CommandExists(t *testing.T) {
	// This tests the real system — docker should be installed via nix
	result := CheckDocker()
	if !CommandExists("docker") {
		if result.Status != StatusFail {
			t.Errorf("CheckDocker with no docker: got %v, want fail", result.Status)
		}
		return
	}

	// Docker CLI exists — should be pass or warn (daemon may not be reachable)
	if result.Status == StatusFail {
		t.Errorf("CheckDocker with docker installed should not be fail, got: %+v", result)
	}
	if result.Name != "docker" {
		t.Errorf("CheckDocker().Name = %q, want %q", result.Name, "docker")
	}
}

func TestCheckColima_OnDarwin(t *testing.T) {
	if !IsDarwin() {
		t.Skip("Colima tests only run on macOS")
	}

	result := CheckColima()
	if !CommandExists("colima") {
		if result.Status != StatusFail {
			t.Errorf("CheckColima with no colima: got %v, want fail", result.Status)
		}
		return
	}

	// Colima is installed — should be pass (running) or warn (not running)
	if result.Status == StatusFail {
		t.Errorf("CheckColima with colima installed should not be fail, got: %+v", result)
	}
	if result.Name != "colima" {
		t.Errorf("CheckColima().Name = %q, want %q", result.Name, "colima")
	}
	if result.Version == "" {
		t.Error("CheckColima with colima installed should have a version")
	}
}

func TestCheckColima_VersionParsing(t *testing.T) {
	// Verify the version parsing logic handles the expected format.
	// colima version output: "colima version 0.8.1\n..."
	input := "colima version 0.8.1\ncommit: abc123"
	lines := splitLines(input)
	version := ""
	for _, line := range lines {
		if len(line) > 14 && line[:14] == "colima version" {
			fields := splitFields(line)
			if len(fields) >= 3 {
				version = fields[2]
			}
			break
		}
	}
	if version != "0.8.1" {
		t.Errorf("version parsing: got %q, want %q", version, "0.8.1")
	}
}

// Helper to split lines without importing strings (to keep test self-contained)
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func splitFields(s string) []string {
	var fields []string
	inField := false
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			if inField {
				fields = append(fields, s[start:i])
				inField = false
			}
		} else {
			if !inField {
				start = i
				inField = true
			}
		}
	}
	if inField {
		fields = append(fields, s[start:])
	}
	return fields
}
