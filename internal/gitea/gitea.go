package gitea

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// RepoExists checks whether a repository exists in the given Gitea instance.
// baseURL should be the scheme+host (eg. https://gitea.local:8443).
func RepoExists(baseURL, owner, repo, token string, insecure bool) (bool, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false, fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = fmt.Sprintf("/api/v1/repos/%s/%s", owner, repo)

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, err
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status checking repo: %d %s", resp.StatusCode, string(body))
}

// CreateRepo creates a repository under an organization. Requires an admin token
// with repo creation permissions.
func CreateRepo(baseURL, owner, repo, token string, private bool, insecure bool) error {
	if token == "" {
		return fmt.Errorf("token required to create repo")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid gitea url: %w", err)
	}
	// Attempt to create under organization first
	orgPath := fmt.Sprintf("/api/v1/orgs/%s/repos", owner)
	adminPath := fmt.Sprintf("/api/v1/admin/users/%s/repos", owner)

	body := map[string]interface{}{
		"name":    repo,
		"private": private,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	// helper to POST to a given path
	doPost := func(path string) (*http.Response, error) {
		u2 := *u
		u2.Path = path
		req, err := http.NewRequest("POST", u2.String(), bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
		return client.Do(req)
	}

	resp, err := doPost(orgPath)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}

	// If organization create fails with NotFound, attempt admin user create
	if resp.StatusCode == http.StatusNotFound {
		// try admin user create (admin token required)
		resp2, err := doPost(adminPath)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		if resp2.StatusCode == http.StatusCreated || resp2.StatusCode == http.StatusOK {
			return nil
		}
		bresp, _ := io.ReadAll(resp2.Body)
		return fmt.Errorf("create repo failed (admin path): %d %s", resp2.StatusCode, string(bresp))
	}

	bresp, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("create repo failed: %d %s", resp.StatusCode, string(bresp))
}

// OrgExists checks whether an organization exists in the given Gitea instance.
func OrgExists(baseURL, orgName, token string, insecure bool) (bool, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return false, fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = fmt.Sprintf("/api/v1/orgs/%s", orgName)

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return false, err
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status checking org: %d %s", resp.StatusCode, string(body))
}

// CreateOrg creates an organization in Gitea. The authenticated user becomes
// the org owner.
func CreateOrg(baseURL, orgName, token string, visibility string, insecure bool) error {
	if token == "" {
		return fmt.Errorf("token required to create org")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = "/api/v1/orgs"

	body := map[string]interface{}{
		"username":   orgName,
		"visibility": visibility,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}
	bresp, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("create org failed: %d %s", resp.StatusCode, string(bresp))
}

// EnsureOrg creates an organization if it doesn't already exist.
func EnsureOrg(baseURL, orgName, token string, insecure bool) error {
	exists, err := OrgExists(baseURL, orgName, token, insecure)
	if err != nil {
		return fmt.Errorf("checking org %s: %w", orgName, err)
	}
	if exists {
		return nil
	}
	return CreateOrg(baseURL, orgName, token, "private", insecure)
}

// ListUserKeys lists SSH keys for a given user (admin endpoint).
func ListUserKeys(baseURL, username, token string, insecure bool) ([]map[string]interface{}, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = fmt.Sprintf("/api/v1/admin/users/%s/keys", username)

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list keys failed: %d %s", resp.StatusCode, string(b))
	}

	var keys []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// CreateUserKey adds an SSH key to the specified user (admin endpoint).
func CreateUserKey(baseURL, username, token, title, key string, insecure bool) error {
	if token == "" {
		return fmt.Errorf("token required to create user key")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = fmt.Sprintf("/api/v1/admin/users/%s/keys", username)

	body := map[string]interface{}{
		"key":   key,
		"title": title,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}
	bresp, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("create user key failed: %d %s", resp.StatusCode, string(bresp))
}

// GetAuthenticatedUser returns the username of the token owner (GET /api/v1/user).
func GetAuthenticatedUser(baseURL, token string, insecure bool) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = "/api/v1/user"

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get user failed: %d %s", resp.StatusCode, string(b))
	}

	var out struct {
		Login string `json:"login"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Login, nil
}

// ListOwnKeys lists SSH keys for the authenticated user (GET /api/v1/user/keys).
func ListOwnKeys(baseURL, token string, insecure bool) ([]map[string]interface{}, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = "/api/v1/user/keys"

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list own keys failed: %d %s", resp.StatusCode, string(b))
	}

	var keys []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}
	return keys, nil
}

// CreateOwnKey adds an SSH key for the authenticated user (POST /api/v1/user/keys).
func CreateOwnKey(baseURL, token, title, key string, insecure bool) error {
	if token == "" {
		return fmt.Errorf("token required to create own key")
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("invalid gitea url: %w", err)
	}
	u.Path = "/api/v1/user/keys"

	body := map[string]interface{}{
		"key":   key,
		"title": title,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	if insecure {
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		client.Transport = tr
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return nil
	}
	bresp, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("create own key failed: %d %s", resp.StatusCode, string(bresp))
}
