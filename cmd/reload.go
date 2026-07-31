package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/Unimart-For-Operations/meta/internal/builder"
	"github.com/Unimart-For-Operations/meta/internal/cluster"
	"github.com/Unimart-For-Operations/meta/internal/gitea"
	"github.com/spf13/cobra"
)

var (
	reloadSkipCreate  bool
	reloadSkipPublish bool
)

var reloadCmd = &cobra.Command{
	Use:   "reload [-- extra idpbuilder create flags]",
	Short: "Reconcile platform changes without tearing down the cluster",
	Long: `Push local changes into a running cluster — no teardown required.

Performs a 2-step reconcile sequence:
  1. Re-run idpbuilder create (idempotent — reconciles packages + ArgoCD apps)
  2. Re-publish all org repos to in-cluster Gitea (push updated branches)

Use this after editing CI workflows, ArgoCD Application YAMLs, or any repo
contents that should be reflected in the running platform without a full
unimart close && unimart open cycle.

Requires a running cluster (unimart open must have been run first).

Extra arguments after -- are passed through to idpbuilder create.`,
	RunE: runReload,
}

func init() {
	reloadCmd.Flags().BoolVar(&reloadSkipCreate, "skip-create", false, "Skip the idpbuilder create reconcile step")
	reloadCmd.Flags().BoolVar(&reloadSkipPublish, "skip-publish", false, "Skip publishing repos to Gitea")
	rootCmd.AddCommand(reloadCmd)
}

func runReload(cmd *cobra.Command, args []string) error {
	orgDir, err := resolveOrgDir()
	if err != nil {
		return err
	}
	idpDir := filepath.Join(orgDir, "idpbuilder")

	// Step 1: Re-run idpbuilder create (idempotent reconcile)
	if reloadSkipCreate {
		fmt.Printf("%s Skipping idpbuilder create (--skip-create)\n", bold("[1/2]"))
	} else {
		fmt.Printf("%s Reconciling IDP platform\n\n", bold("[1/2]"))

		packagesDir := filepath.Join(orgDir, "packages")
		createArgs := idpCreateArgs(packagesDir, args)

		if err := builder.Create(idpDir, createArgs); err != nil {
			return err
		}
		fmt.Printf("  %s platform reconciled\n", pass("[ok]"))
	}

	// Step 2: Re-publish org repos to Gitea
	if reloadSkipPublish {
		fmt.Printf("\n%s Skipping repo publish (--skip-publish)\n", bold("[2/2]"))
	} else {
		fmt.Printf("\n%s Publishing org repos to in-cluster Gitea\n\n", bold("[2/2]"))

		token, err := cluster.GetGiteaAdminToken(defaultGiteaURL)
		if err != nil {
			fmt.Printf("  %s could not discover Gitea token: %v\n", warn("[warn]"), err)
			fmt.Println("  Skipping publish — run manually: unimart freezer repos publish-to-gitea")
		} else {
			if err := gitea.EnsureOrg(defaultGiteaURL, defaultOwner, token, true); err != nil {
				fmt.Printf("  %s could not ensure org %s: %v\n", warn("[warn]"), defaultOwner, err)
			} else {
				fmt.Printf("  %s org %s ready\n", pass("[ok]"), defaultOwner)
			}

			if err := publishAllRepos(orgDir, token); err != nil {
				fmt.Printf("  %s publish failed: %v\n", warn("[warn]"), err)
				fmt.Println("  Run manually: unimart freezer repos publish-to-gitea")
			} else {
				fmt.Printf("  %s all repos published\n", pass("[ok]"))
			}
		}
	}

	fmt.Println()
	fmt.Printf("%s platform reloaded\n", pass("done"))
	fmt.Println()
	fmt.Println("  ArgoCD will reconcile changes within ~3 minutes.")
	fmt.Printf("  Dashboard: %s\n", argoCDURL)

	return nil
}
