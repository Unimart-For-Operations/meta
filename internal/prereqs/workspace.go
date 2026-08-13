package prereqs

import (
	"os"
	"path/filepath"
)

// CheckWorkspaceIdpbuilder verifies the idpbuilder source tree (nested Go
// module) is present within the org directory. It is compiled into unimart,
// so no separate binary build is needed.
func CheckWorkspaceIdpbuilder(orgDir string) []CheckResult {
	idpDir := filepath.Join(orgDir, "idpbuilder")
	if info, err := os.Stat(idpDir); err != nil || !info.IsDir() {
		return []CheckResult{{
			Name:   "idpbuilder source",
			Status: StatusFail,
			Detail: "not found — the meta checkout is incomplete",
		}}
	}
	return []CheckResult{{
		Name:   "idpbuilder source",
		Status: StatusPass,
	}}
}
