package prereqs

import "strings"

// CheckKubectl verifies that kubectl is installed.
func CheckKubectl() CheckResult {
	if !CommandExists("kubectl") {
		return CheckResult{
			Name:   "kubectl",
			Status: StatusFail,
			Detail: "not installed",
		}
	}

	version := "installed"

	// Try JSON output and extract gitVersion
	out, err := CommandOutput("kubectl", "version", "--client", "-o", "json")
	if err == nil {
		// Quick parse: look for "gitVersion": "v1.30.5"
		for _, line := range strings.Split(out, "\n") {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "gitVersion") {
				// Extract value between quotes after the colon
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					v := strings.TrimSpace(parts[1])
					v = strings.Trim(v, `",`)
					version = v
				}
				break
			}
		}
	}

	return CheckResult{
		Name:    "kubectl",
		Status:  StatusPass,
		Version: version,
	}
}
