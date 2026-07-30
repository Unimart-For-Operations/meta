package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/idpbuilder/meta/internal/cluster"
	"github.com/idpbuilder/meta/internal/gitea"
	"github.com/idpbuilder/meta/internal/repos"
	"github.com/spf13/cobra"
)

var (
	publishGiteaURL       string
	publishOwner          string
	publishToken          string
	publishDryRun         bool
	publishNonInteractive bool
	publishUseSSH         bool
	publishSSHKeyPath     string
)

var freezerReposPublishCmd = &cobra.Command{
	Use:   "publish-to-gitea",
	Short: "Publish local org repos into an in-cluster Gitea instance",
	RunE:  runFreezerReposPublish,
}

func init() {
	freezerReposPublishCmd.Flags().StringVar(&publishGiteaURL, "gitea-url", "", "Gitea base URL (eg https://gitea.local:8443)")
	freezerReposPublishCmd.Flags().StringVar(&publishOwner, "owner", "idpbuilder", "Gitea org/owner to publish into")
	freezerReposPublishCmd.Flags().StringVar(&publishToken, "token", "", "Gitea admin token (optional; prompt/use cluster secret if empty)")
	freezerReposPublishCmd.Flags().BoolVar(&publishDryRun, "dry-run", true, "Show planned actions without executing")
	freezerReposPublishCmd.Flags().BoolVar(&publishNonInteractive, "non-interactive", false, "Run without prompts (requires --token and --use-ssh flags set as needed)")
	freezerReposPublishCmd.Flags().BoolVar(&publishUseSSH, "use-ssh", false, "Use SSH remotes for pushes (non-interactive friendly)")
	freezerReposPublishCmd.Flags().StringVar(&publishSSHKeyPath, "ssh-key-path", "", "Path to public SSH key to upload (non-interactive)")
	freezerReposCmd.AddCommand(freezerReposPublishCmd)
}

func runFreezerReposPublish(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	if publishGiteaURL == "" {
		return fmt.Errorf("--gitea-url is required")
	}

	local, sourceDir, err := repos.ListPublishable(orgDir)
	if err != nil {
		return err
	}
	if len(local) == 0 {
		fmt.Printf("No local repos found to publish in %s\n", sourceDir)
		fmt.Printf("Add repos under %s to publish by default\n", repos.RepositoriesDir(orgDir))
		return nil
	}

	// Print planned actions
	fmt.Printf("Planned repositories to publish (from %s):\n", sourceDir)
	for _, r := range local {
		fmt.Printf("  - %s\n", r.Name)
	}

	if publishDryRun {
		fmt.Println("\nDry-run mode enabled; no changes will be made. Re-run with --dry-run=false to execute.")
		return nil
	}

	// Token resolution (interactive or non-interactive)
	token := publishToken
	if token == "" {
		// try cluster discovery
		t, err := cluster.GetGiteaAdminToken()
		if err == nil {
			token = t
			if !publishNonInteractive {
				fmt.Println("Discovered admin token from cluster.")
			}
		} else if !publishNonInteractive {
			if promptYesNo("No token provided. Try to discover admin token from in-cluster secret 'gitea-credential'? (Y/n): ", true) {
				t2, err2 := cluster.GetGiteaAdminToken()
				if err2 == nil {
					token = t2
					fmt.Println("Discovered admin token from cluster.")
				} else {
					fmt.Printf("Could not discover token: %v\n", err2)
				}
			}
		}
	}

	if token == "" && !publishNonInteractive {
		tok := promptString("Paste Gitea admin token (will be used for API actions): ")
		token = strings.TrimSpace(tok)
	}

	if token == "" {
		return fmt.Errorf("no admin token provided or discovered; aborting")
	}

	if !publishNonInteractive {
		fmt.Println()
		if !promptYesNo("Proceed to create missing repos and push branches/tags to Gitea? (y/N): ", false) {
			fmt.Println("publish cancelled")
			return nil
		}
	}

	// Execute actions
	for _, r := range local {
		repoName := r.Name
		fmt.Printf("Ensuring repo %s/%s exists...\n", publishOwner, repoName)
		exists, err := gitea.RepoExists(publishGiteaURL, publishOwner, repoName, token, true)
		if err != nil {
			return err
		}
		if !exists {
			fmt.Printf("Creating repo %s/%s...\n", publishOwner, repoName)
			if err := gitea.CreateRepo(publishGiteaURL, publishOwner, repoName, token, true, true); err != nil {
				return err
			}
		}

		// set remote and push (prefer SSH if user opted in)
		repoPath := r.Path
		if repoPath == "" {
			repoPath = filepath.Join(sourceDir, repoName)
		}
		var useSSH bool
		if publishNonInteractive {
			useSSH = publishUseSSH
		} else {
			useSSH = promptYesNo("Use SSH remote for pushes? (recommended) (Y/n): ", true)
		}
		if useSSH {
			// If SSH chosen, ensure key exists on remote user; attempt to add it if missing
			keys, kerr := gitea.ListUserKeys(publishGiteaURL, publishOwner, token, true)
			if kerr == nil {
				keyPath := publishSSHKeyPath
				if keyPath == "" {
					keyPath = filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub")
				}
				pub, _ := os.ReadFile(keyPath)
				hasKey := false
				if pub != nil {
					for _, k := range keys {
						if s, ok := k["key"].(string); ok && strings.TrimSpace(s) == strings.TrimSpace(string(pub)) {
							hasKey = true
							break
						}
					}
				}
				if !hasKey {
					authUser, _ := gitea.GetAuthenticatedUser(publishGiteaURL, token, true)
					if authUser == publishOwner {
						if promptYesNo("Local SSH public key not found in your Gitea account. Upload it now to your account? (y/N): ", false) {
							title := promptString("Key title (e.g. 'laptop'): ")
							if perr := gitea.CreateOwnKey(publishGiteaURL, token, title, strings.TrimSpace(string(pub)), true); perr != nil {
								return fmt.Errorf("failed to create own key: %w", perr)
							}
							fmt.Println("Uploaded SSH key to your Gitea account.")
						} else {
							fmt.Println("Skipping SSH upload; push may fail if key not present.")
						}
					} else {
						if promptYesNo("Local SSH public key not found in Gitea user. Upload it now (requires admin token)? (y/N): ", false) {
							title := promptString("Key title (e.g. 'laptop'): ")
							if perr := gitea.CreateUserKey(publishGiteaURL, publishOwner, token, title, strings.TrimSpace(string(pub)), true); perr != nil {
								return fmt.Errorf("failed to create user key: %w", perr)
							}
							fmt.Println("Uploaded SSH key to Gitea user account.")
						} else {
							fmt.Println("Skipping SSH upload; push may fail if key not present.")
						}
					}
				}
			}
		}

		if err := repos.SetRemoteAndPush(repoPath, publishGiteaURL, publishOwner, repoName, token, useSSH); err != nil {
			return err
		}
		fmt.Printf("Published %s\n", repoName)
	}

	return nil
}
