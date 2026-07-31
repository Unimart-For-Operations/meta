package cmd

import (
	"fmt"

	"github.com/Unimart-For-Operations/meta/internal/repos"
	"github.com/spf13/cobra"
)

var reposOrg string

var freezerReposCmd = &cobra.Command{
	Use:   "repos",
	Short: "Manage organization repositories",
	Long:  `List, clone, and check status of repositories in the idpbuilder GitHub organization.`,
}

var freezerReposListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all repos in the organization",
	RunE:  runFreezerReposList,
}

var freezerReposCloneCmd = &cobra.Command{
	Use:   "clone [repo-name]",
	Short: "Clone a repo from the organization",
	Long:  `Clone a specific repo, or all repos if no name is given.`,
	RunE:  runFreezerReposClone,
}

var freezerReposStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show git status of all local repos",
	RunE:  runFreezerReposStatus,
}

func init() {
	freezerReposCmd.PersistentFlags().StringVar(&reposOrg, "org", "idpbuilder", "GitHub organization name")
	freezerReposCmd.AddCommand(freezerReposListCmd)
	freezerReposCmd.AddCommand(freezerReposCloneCmd)
	freezerReposCmd.AddCommand(freezerReposStatusCmd)
	// publish-to-gitea command is added by freezer_repos_publish.go during package init
	freezerCmd.AddCommand(freezerReposCmd)
}

func runFreezerReposList(cmd *cobra.Command, args []string) error {
	fmt.Printf("%s Repositories in %s:\n\n", bold(">>"), reposOrg)

	remote, err := repos.ListRemote(reposOrg)
	if err != nil {
		return err
	}

	if len(remote) == 0 {
		fmt.Println("  No repositories found")
		return nil
	}

	for _, r := range remote {
		flags := ""
		if r.IsFork {
			flags += " (fork)"
		}
		if r.IsPrivate {
			flags += " (private)"
		}
		desc := r.Description
		if desc == "" {
			desc = dim("no description")
		}
		fmt.Printf("  %-30s %s%s\n", r.Name, desc, flags)
	}

	return nil
}

func runFreezerReposClone(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	if len(args) > 0 {
		// Clone specific repo
		for _, name := range args {
			if err := repos.Clone(orgDir, reposOrg, name); err != nil {
				return err
			}
			fmt.Printf("  %s %s cloned\n", pass("[ok]"), name)
		}
		return nil
	}

	// Clone all repos
	remote, err := repos.ListRemote(reposOrg)
	if err != nil {
		return err
	}

	cloned := 0
	for _, r := range remote {
		if err := repos.Clone(orgDir, reposOrg, r.Name); err != nil {
			// Already exists is not a fatal error
			fmt.Printf("  [--] %s: %v\n", r.Name, err)
			continue
		}
		fmt.Printf("  %s %s cloned\n", pass("[ok]"), r.Name)
		cloned++
	}

	fmt.Printf("\n%s Cloned %d repo(s)\n", pass("done"), cloned)
	return nil
}

func runFreezerReposStatus(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	local, sourceDir, err := repos.ListPublishable(orgDir)
	if err != nil {
		return err
	}

	fmt.Printf("%s Repositories ready for publish (from %s):\n\n", bold(">>"), sourceDir)

	if len(local) == 0 {
		fmt.Println("  No git repositories found")
		return nil
	}

	for _, r := range local {
		status := pass("[clean]")
		if !r.Clean {
			status = warn("[dirty]")
		}
		branch := r.Branch
		if branch == "" {
			branch = "detached"
		}
		fmt.Printf("  %s %-25s branch=%-15s\n", status, r.Name, branch)
	}

	return nil
}
