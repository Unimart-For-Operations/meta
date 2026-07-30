package repos

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultOrg = "idpbuilder"

// repositoriesDirName is the designated directory under orgDir that contains
// repos intended for local Gitea publish flows.
const repositoriesDirName = "repositories"

// Repo represents a GitHub repository in the org.
type Repo struct {
	Name        string
	Description string
	CloneURL    string
	SSHURL      string
	IsPrivate   bool
	IsFork      bool
}

// LocalRepo represents a locally cloned repo.
type LocalRepo struct {
	Name   string
	Path   string
	Branch string
	Clean  bool
	Ahead  int
	Behind int
}

// ListRemote lists all repos in the GitHub organization using gh CLI.
func ListRemote(org string) ([]Repo, error) {
	if org == "" {
		org = defaultOrg
	}

	out, err := exec.Command("gh", "repo", "list", org,
		"--json", "name,description,url,sshUrl,isPrivate,isFork",
		"--limit", "100",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("listing repos for %s: %w", org, err)
	}

	var ghRepos []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		URL         string `json:"url"`
		SSHURL      string `json:"sshUrl"`
		IsPrivate   bool   `json:"isPrivate"`
		IsFork      bool   `json:"isFork"`
	}

	if err := json.Unmarshal(out, &ghRepos); err != nil {
		return nil, fmt.Errorf("parsing repo list: %w", err)
	}

	var repos []Repo
	for _, r := range ghRepos {
		repos = append(repos, Repo{
			Name:        r.Name,
			Description: r.Description,
			CloneURL:    r.URL,
			SSHURL:      r.SSHURL,
			IsPrivate:   r.IsPrivate,
			IsFork:      r.IsFork,
		})
	}
	return repos, nil
}

// ListLocal scans the org directory for cloned repos and reports their git status.
func ListLocal(orgDir string) ([]LocalRepo, error) {
	return listLocalInDir(orgDir)
}

// RepositoriesDir returns the designated publish directory under orgDir.
func RepositoriesDir(orgDir string) string {
	return filepath.Join(orgDir, repositoriesDirName)
}

// ListPublishable discovers repos for local Gitea publish flows.
//
// Behavior:
// - If <orgDir>/repositories exists, only repos under that directory are used.
// - If it does not exist, falls back to scanning orgDir (legacy behavior).
//
// It returns both the discovered repos and the source directory used.
func ListPublishable(orgDir string) ([]LocalRepo, string, error) {
	publishDir := RepositoriesDir(orgDir)
	if info, err := os.Stat(publishDir); err == nil {
		if !info.IsDir() {
			return nil, publishDir, fmt.Errorf("%s exists but is not a directory", publishDir)
		}
		repos, err := listLocalInDir(publishDir)
		return repos, publishDir, err
	} else if err != nil && !os.IsNotExist(err) {
		return nil, publishDir, fmt.Errorf("checking %s: %w", publishDir, err)
	}

	// Legacy fallback: publish from org root when designated directory doesn't exist.
	repos, err := listLocalInDir(orgDir)
	return repos, orgDir, err
}

// listLocalInDir scans a directory for child git repos and reports their status.
func listLocalInDir(dir string) ([]LocalRepo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	var repos []LocalRepo
	for _, entry := range entries {
		repoPath := filepath.Join(dir, entry.Name())

		// Use os.Stat (follows symlinks) so symlinked directories are included.
		// DirEntry.IsDir() does not follow symlinks and would skip them.
		info, err := os.Stat(repoPath)
		if err != nil || !info.IsDir() {
			continue
		}

		gitDir := filepath.Join(repoPath, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			continue // Not a git repo
		}

		repo := LocalRepo{
			Name: entry.Name(),
			Path: repoPath,
		}

		// Get current branch
		if branch, err := gitOutput(repoPath, "branch", "--show-current"); err == nil {
			repo.Branch = branch
		}

		// Check if clean
		if status, err := gitOutput(repoPath, "status", "--porcelain"); err == nil {
			repo.Clean = status == ""
		}

		repos = append(repos, repo)
	}

	return repos, nil
}

// Clone clones a repo into the org directory.
func Clone(orgDir, org, repoName string) error {
	if org == "" {
		org = defaultOrg
	}

	dest := filepath.Join(orgDir, repoName)
	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("%s already exists at %s", repoName, dest)
	}

	fmt.Printf("  Cloning %s/%s...\n", org, repoName)
	cmd := exec.Command("gh", "repo", "clone", fmt.Sprintf("%s/%s", org, repoName), dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cloning %s: %w", repoName, err)
	}

	return nil
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// remoteName is the git remote used for Gitea pushes. Using a dedicated name
// avoids clobbering the "origin" remote that points to GitHub.
const remoteName = "gitea"

// SetRemoteAndPush configures a "gitea" remote on the repo and pushes all
// branches and tags. For HTTPS (useSSH=false), token is embedded in the push
// URL for authentication and TLS verification is disabled (self-signed cert).
// For SSH, the token is unused for push; authentication relies on SSH keys.
func SetRemoteAndPush(repoPath, giteaBaseURL, owner, repo, token string, useSSH bool) error {
	// Build the remote URL for storage in .git/config (no credentials).
	// For HTTPS pushes we embed credentials in the push command directly.
	var remoteURL string
	if useSSH {
		u, err := url.Parse(giteaBaseURL)
		if err != nil {
			return fmt.Errorf("invalid gitea url for ssh remote: %w", err)
		}
		host := u.Hostname()

		sshPort := getGiteaSSHNodePort()
		if sshPort == "" {
			port := u.Port()
			if port == "" || port == "22" || port == "443" {
				remoteURL = fmt.Sprintf("git@%s:%s/%s.git", host, owner, repo)
			} else {
				remoteURL = fmt.Sprintf("ssh://git@%s:%s/%s/%s.git", host, port, owner, repo)
			}
		} else {
			remoteURL = fmt.Sprintf("ssh://git@%s:%s/%s/%s.git", host, sshPort, owner, repo)
		}
	} else {
		remoteURL = fmt.Sprintf("%s/%s/%s.git", strings.TrimRight(giteaBaseURL, "/"), owner, repo)
	}

	// Set or add the "gitea" remote — always scoped to repoPath via gitOutput.
	if _, err := gitOutput(repoPath, "remote", "get-url", remoteName); err == nil {
		if _, err := gitOutput(repoPath, "remote", "set-url", remoteName, remoteURL); err != nil {
			return fmt.Errorf("set remote %s: %w", remoteName, err)
		}
	} else {
		if _, err := gitOutput(repoPath, "remote", "add", remoteName, remoteURL); err != nil {
			return fmt.Errorf("add remote %s: %w", remoteName, err)
		}
	}

	// For HTTPS pushes, build a URL with embedded credentials so git does not
	// prompt. The stored remote URL stays credential-free.
	pushURL := remoteName
	if !useSSH && token != "" {
		u, err := url.Parse(giteaBaseURL)
		if err == nil {
			u.User = url.UserPassword(owner, token)
			u.Path = fmt.Sprintf("/%s/%s.git", owner, repo)
			pushURL = u.String()
		}
	}

	// Environment for git commands: skip TLS verification for self-signed certs,
	// inherit current environment for PATH etc.
	gitEnv := append(os.Environ(), "GIT_SSL_NO_VERIFY=true", "GIT_TERMINAL_PROMPT=0")

	// Push all branches (--no-verify skips pre-push hooks — this is a mirror push)
	cmd := exec.Command("git", "push", "--no-verify", pushURL, "--all")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pushing branches: %w", err)
	}

	// Push tags
	cmd = exec.Command("git", "push", "--no-verify", pushURL, "--tags")
	cmd.Dir = repoPath
	cmd.Env = gitEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pushing tags: %w", err)
	}

	return nil
}

// getGiteaSSHNodePort retrieves the NodePort for the gitea SSH service
func getGiteaSSHNodePort() string {
	out, err := exec.Command("kubectl", "get", "service", "-n", "gitea", "my-gitea-ssh", "-o", "jsonpath={.spec.ports[?(@.name==\"ssh\")].nodePort}").Output()
	if err != nil {
		// Fallback to checking if the service exists with different naming
		out, err = exec.Command("kubectl", "get", "service", "-n", "gitea", "-o", "jsonpath={.items[?(@.metadata.name==\"my-gitea-ssh\")].spec.ports[?(@.name==\"ssh\")].nodePort}").Output()
		if err != nil {
			return "" // Could not determine SSH port
		}
	}
	return strings.TrimSpace(string(out))
}
