package cmd

import "github.com/spf13/cobra"

var freezerGiteaCmd = &cobra.Command{
	Use:   "gitea",
	Short: "Manage the in-cluster Gitea instance",
	Long:  `Configure and manage the Gitea instance running in the IDP cluster (users, keys, orgs).`,
}

func init() {
	freezerCmd.AddCommand(freezerGiteaCmd)
}
