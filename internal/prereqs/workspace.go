package prereqs

import (
	"os"
	"path/filepath"
)

// CheckWorkspaceIdpbuilder verifies the idpbuilder repo within the org directory.
func CheckWorkspaceIdpbuilder(orgDir string) []CheckResult {
	var results []CheckResult

	// Check idpbuilder repo exists
	idpDir := filepath.Join(orgDir, "idpbuilder")
	if info, err := os.Stat(idpDir); err != nil || !info.IsDir() {
		results = append(results, CheckResult{
			Name:   "idpbuilder repo",
			Status: StatusFail,
			Detail: "not found — run: unimart freezer repos clone",
		})
		return results
	}
	results = append(results, CheckResult{
		Name:   "idpbuilder repo",
		Status: StatusPass,
	})

	// Check idpbuilder binary exists
	binaryPath := filepath.Join(idpDir, "idpbuilder")
	if _, err := os.Stat(binaryPath); err != nil {
		results = append(results, CheckResult{
			Name:   "idpbuilder binary",
			Status: StatusWarn,
			Detail: "not built — run: unimart freezer build",
		})
	} else {
		results = append(results, CheckResult{
			Name:   "idpbuilder binary",
			Status: StatusPass,
			Detail: "built",
		})
	}

	// Check Makefile exists (confirms it's a valid idpbuilder checkout)
	makefile := filepath.Join(idpDir, "Makefile")
	if _, err := os.Stat(makefile); err != nil {
		results = append(results, CheckResult{
			Name:   "idpbuilder Makefile",
			Status: StatusWarn,
			Detail: "not found — is this a valid idpbuilder checkout?",
		})
	}

	return results
}
