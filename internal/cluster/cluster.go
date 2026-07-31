package cluster

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
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

// GetGiteaAdminCredentials reads the Gitea admin username and password from
// the "gitea-credential" secret in the gitea namespace.
func GetGiteaAdminCredentials() (username, password string, err error) {
	readSecret := func(key string) (string, error) {
		out, err := exec.Command("kubectl", "get", "secret", "gitea-credential", "-n", "gitea", "-o", fmt.Sprintf("jsonpath={.data.%s}", key)).Output()
		if err != nil {
			return "", err
		}
		if len(strings.TrimSpace(string(out))) == 0 {
			return "", fmt.Errorf("key %s not present in gitea-credential secret", key)
		}
		decoded, derr := decodeBase64(strings.TrimSpace(string(out)))
		if derr != nil {
			return "", derr
		}
		return decoded, nil
	}

	username, err = readSecret("username")
	if err != nil {
		return "", "", err
	}
	password, err = readSecret("password")
	if err != nil {
		return "", "", err
	}
	return username, password, nil
}

// mintGiteaTokenWithURL creates an all-scope access token for the given user
// via basic auth (POST /api/v1/users/{username}/tokens) and returns the
// token value. Idempotent: an existing "unimart" token is deleted first.
func mintGiteaTokenWithURL(username, password, baseURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid gitea url: %w", err)
	}

	// Remove a previously created "unimart" token so we don't accumulate.
	if existing, err := listGiteaTokens(client, base, username, password); err == nil {
		for _, t := range existing {
			if t.Name == "unimart" {
				_ = deleteGiteaToken(client, base, username, password, t.ID)
			}
		}
	}

	u := *base
	u.Path = fmt.Sprintf("/api/v1/users/%s/tokens", username)
	body := map[string]interface{}{
		"name":   "unimart",
		"scopes": []string{"all"},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(username, password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		rb, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("minting gitea token failed: %d %s", resp.StatusCode, string(rb))
	}

	var out struct {
		Token string `json:"sha1"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Token, nil
}

// listGiteaTokens lists access tokens for a user via basic auth.
func listGiteaTokens(client *http.Client, base *url.URL, username, password string) ([]struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}, error) {
	u := *base
	u.Path = fmt.Sprintf("/api/v1/users/%s/tokens", username)
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list user tokens failed: %d", resp.StatusCode)
	}
	var tokens []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

// deleteGiteaToken removes a user's access token via basic auth.
func deleteGiteaToken(client *http.Client, base *url.URL, username, password string, id int64) error {
	u := *base
	u.Path = fmt.Sprintf("/api/v1/users/%s/tokens/%d", username, id)
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(username, password)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete user token failed: %d", resp.StatusCode)
	}
	return nil
}

// GetGiteaAdminToken attempts to obtain a Gitea admin token. It prefers the
// token stored in the "gitea-credential" secret (data.token). If the secret
// only carries username+password (some idpbuilder versions), it mints an
// all-scope token on demand via the Gitea API using basic auth. baseURL is
// the Gitea instance used for token minting. Returns an error if neither
// path succeeds.
func GetGiteaAdminToken(baseURL string) (string, error) {
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

	// Fallback: mint a token from the admin username+password using basic auth.
	username, password, err := GetGiteaAdminCredentials()
	if err != nil {
		return "", fmt.Errorf("gitea admin token not found in cluster secrets: %w", err)
	}
	return mintGiteaTokenWithURL(username, password, baseURL)
}
