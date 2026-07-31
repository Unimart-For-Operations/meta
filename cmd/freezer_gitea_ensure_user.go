package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/cluster"
	"github.com/Unimart-For-Operations/meta/internal/gitea"
	"github.com/spf13/cobra"
)

var (
	ensureUsername string
	ensureEmail    string
	ensurePassword string
	ensureSSHKey   string
	ensureOrg      string
	ensureGiteaURL string
	ensurePublish  bool
)

var freezerGiteaEnsureUserCmd = &cobra.Command{
	Use:   "ensure-user",
	Short: "Create (or update) a personal Gitea user account",
	Long: `Ensures a personal user account exists in the in-cluster Gitea instance.

Creates the user (if missing) with the given password, uploads your SSH
public key, creates an access token for the user, and makes the user an
owner of the target organization. With --publish, local org repos are also
published into the organization via SSH.

Runs against the currently running IDP cluster.`,
	RunE: runFreezerGiteaEnsureUser,
}

func init() {
	freezerGiteaEnsureUserCmd.Flags().StringVar(&ensureUsername, "username", "andrcmdr", "Gitea username to ensure")
	freezerGiteaEnsureUserCmd.Flags().StringVar(&ensureEmail, "email", "andrcmdr@protonmail.com", "Email for the Gitea user")
	freezerGiteaEnsureUserCmd.Flags().StringVar(&ensurePassword, "password", "developer", "Password for the Gitea user")
	freezerGiteaEnsureUserCmd.Flags().StringVar(&ensureSSHKey, "ssh-key", filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519.pub"), "Path to the public SSH key to upload")
	freezerGiteaEnsureUserCmd.Flags().StringVar(&ensureOrg, "org", "Unimart-For-Operations", "Gitea organization the user should own")
	freezerGiteaEnsureUserCmd.Flags().StringVar(&ensureGiteaURL, "gitea-url", "https://gitea.cnoe.localtest.me:8443", "Gitea base URL")
	freezerGiteaEnsureUserCmd.Flags().BoolVar(&ensurePublish, "publish", false, "Publish local org repos into the Gitea organization after ensuring the user")
	freezerGiteaCmd.AddCommand(freezerGiteaEnsureUserCmd)
}

func runFreezerGiteaEnsureUser(cmd *cobra.Command, args []string) error {
	// Resolve admin token (from cluster secret, minted on demand if needed).
	adminToken, err := cluster.GetGiteaAdminToken(ensureGiteaURL)
	if err != nil {
		return fmt.Errorf("could not obtain Gitea admin token: %w", err)
	}

	// 1. Ensure the user exists.
	fmt.Printf("%s Ensuring user %s\n", bold(">>"), ensureUsername)
	created, err := gitea.EnsureUser(ensureGiteaURL, adminToken, ensureUsername, ensurePassword, ensureEmail, true)
	if err != nil {
		return err
	}
	if created {
		fmt.Printf("  %s user %s created\n", pass("[ok]"), ensureUsername)
	} else {
		fmt.Printf("  %s user %s already exists\n", dim("[--]"), ensureUsername)
	}

	// 2. Upload SSH key if missing.
	pub, err := os.ReadFile(ensureSSHKey)
	if err != nil {
		fmt.Printf("  %s could not read SSH key %s: %v\n", warn("[warn]"), ensureSSHKey, err)
	} else {
		keys, kerr := gitea.ListUserKeys(ensureGiteaURL, ensureUsername, adminToken, true)
		if kerr != nil {
			return fmt.Errorf("listing keys for %s: %w", ensureUsername, kerr)
		}
		hasKey := false
		for _, k := range keys {
			if s, ok := k["key"].(string); ok && strings.TrimSpace(s) == strings.TrimSpace(string(pub)) {
				hasKey = true
				break
			}
		}
		if hasKey {
			fmt.Printf("  %s SSH key already uploaded\n", dim("[--]"))
		} else {
			title := defaultKeyTitle()
			if err := gitea.CreateUserKey(ensureGiteaURL, ensureUsername, adminToken, title, strings.TrimSpace(string(pub)), true); err != nil {
				return fmt.Errorf("uploading SSH key: %w", err)
			}
			fmt.Printf("  %s SSH key uploaded (%s)\n", pass("[ok]"), title)
		}
	}

	// 3. Create an access token for the user so they act as org owner.
	fmt.Printf("%s Ensuring access token for %s\n", bold(">>"), ensureUsername)
	userToken, err := gitea.CreateUserToken(ensureGiteaURL, ensureUsername, ensurePassword, true)
	if err != nil {
		return fmt.Errorf("creating token for %s: %w", ensureUsername, err)
	}

	// 4. Ensure the organization exists and the user owns it.
	fmt.Printf("%s Ensuring org %s\n", bold(">>"), ensureOrg)
	if err := gitea.EnsureOrg(ensureGiteaURL, ensureOrg, userToken, true); err != nil {
		// Org may already exist without the user as owner; fall back to admin token.
		if err2 := gitea.EnsureOrg(ensureGiteaURL, ensureOrg, adminToken, true); err2 != nil {
			return fmt.Errorf("ensuring org %s: %w", ensureOrg, err2)
		}
	}
	if err := gitea.AddOrgOwner(ensureGiteaURL, ensureOrg, ensureUsername, adminToken, true); err != nil {
		fmt.Printf("  %s could not make %s an org owner: %v\n", warn("[warn]"), ensureUsername, err)
	} else {
		fmt.Printf("  %s %s is an owner of %s\n", pass("[ok]"), ensureUsername, ensureOrg)
	}

	fmt.Println()
	fmt.Println("  Gitea login:")
	fmt.Printf("    URL:      %s\n", ensureGiteaURL)
	fmt.Printf("    Username: %s\n", ensureUsername)
	fmt.Printf("    Password: %s\n", ensurePassword)
	fmt.Println()

	// 5. Optionally publish repos into the org via SSH as the user.
	if ensurePublish {
		fmt.Printf("%s Publishing org repos into %s\n\n", bold(">>"), ensureOrg)
		publishGiteaURL = ensureGiteaURL
		publishOwner = ensureOrg
		publishToken = userToken
		publishDryRun = false
		publishNonInteractive = true
		publishUseSSH = true
		publishSSHKeyPath = ensureSSHKey
		publishSSHUser = ensureUsername
		if err := runFreezerReposPublish(nil, nil); err != nil {
			return fmt.Errorf("publish step failed: %w", err)
		}
	}

	return nil
}
