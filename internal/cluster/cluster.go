package cluster

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ArgoApp represents a minimal ArgoCD Application for status display.
type ArgoApp struct {
	Name      string
	Namespace string
	Status    string
	Health    string
	SyncedAt  string
}

// Secret represents a key-value secret for display.
type Secret struct {
	Namespace string
	Name      string
	Key       string
	Value     string
}

// IsClusterRunning checks if a Kind cluster is running.
func IsClusterRunning() bool {
	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		return false
	}
	clusters := strings.TrimSpace(string(out))
	return clusters != ""
}

// GetClusterName returns the name of the running Kind cluster.
func GetClusterName() string {
	out, err := exec.Command("kind", "get", "clusters").Output()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	return ""
}

// GetArgoApps retrieves ArgoCD Application status from the cluster.
func GetArgoApps() ([]ArgoApp, error) {
	out, err := exec.Command("kubectl", "get", "applications", "-n", "argocd", "-o", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get ArgoCD applications: %w", err)
	}

	var result struct {
		Items []struct {
			Metadata struct {
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"metadata"`
			Status struct {
				Sync struct {
					Status string `json:"status"`
				} `json:"sync"`
				Health struct {
					Status string `json:"status"`
				} `json:"health"`
				OperationState struct {
					FinishedAt string `json:"finishedAt"`
				} `json:"operationState"`
			} `json:"status"`
		} `json:"items"`
	}

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing ArgoCD response: %w", err)
	}

	var apps []ArgoApp
	for _, item := range result.Items {
		apps = append(apps, ArgoApp{
			Name:      item.Metadata.Name,
			Namespace: item.Metadata.Namespace,
			Status:    item.Status.Sync.Status,
			Health:    item.Status.Health.Status,
			SyncedAt:  item.Status.OperationState.FinishedAt,
		})
	}
	return apps, nil
}

// GetSecrets retrieves IDP secrets (ArgoCD admin, Gitea admin).
func GetSecrets() ([]Secret, error) {
	var secrets []Secret

	// ArgoCD admin password
	out, err := exec.Command("kubectl", "get", "secret", "argocd-initial-admin-secret",
		"-n", "argocd", "-o", "jsonpath={.data.password}").Output()
	if err == nil {
		decoded, derr := decodeBase64(strings.TrimSpace(string(out)))
		if derr == nil {
			secrets = append(secrets, Secret{
				Namespace: "argocd",
				Name:      "argocd-initial-admin-secret",
				Key:       "password",
				Value:     decoded,
			})
		}
	}

	// Gitea admin password
	out, err = exec.Command("kubectl", "get", "secret", "gitea-admin-secret",
		"-n", "gitea", "-o", "jsonpath={.data.password}").Output()
	if err == nil {
		decoded, derr := decodeBase64(strings.TrimSpace(string(out)))
		if derr == nil {
			secrets = append(secrets, Secret{
				Namespace: "gitea",
				Name:      "gitea-admin-secret",
				Key:       "password",
				Value:     decoded,
			})
		}
	}

	return secrets, nil
}

// decodeBase64 decodes a base64-encoded string.
func decodeBase64(s string) (string, error) {
	out, err := exec.Command("bash", "-c", fmt.Sprintf("echo '%s' | base64 -d", s)).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// GetGiteaAdminToken attempts to read a Gitea admin token from known secret
// locations in the cluster. It prefers the token stored in the
// "gitea-credential" secret (data.token), falling back to other common
// secret names. Returns an error if the token cannot be located or decoded.
func GetGiteaAdminToken() (string, error) {
	// Try gitea-credential (used by idpbuilder manifests)
	out, err := exec.Command("kubectl", "get", "secret", "gitea-credential", "-n", "gitea", "-o", "jsonpath={.data.token}").Output()
	if err == nil {
		if len(strings.TrimSpace(string(out))) > 0 {
			decoded, derr := decodeBase64(strings.TrimSpace(string(out)))
			if derr == nil {
				return decoded, nil
			}
		}
	}

	// Fallback: older installers may use gitea-admin-secret with password only
	// In that case we cannot derive a token, so return not found.
	return "", fmt.Errorf("gitea admin token not found in cluster secrets")
}
