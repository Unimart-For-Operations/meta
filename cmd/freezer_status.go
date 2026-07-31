package cmd

import (
	"fmt"

	"github.com/Unimart-For-Operations/meta/internal/cluster"
	"github.com/Unimart-For-Operations/meta/internal/colima"
	"github.com/Unimart-For-Operations/meta/internal/prereqs"
	"github.com/spf13/cobra"
)

var showSecrets bool

var freezerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show IDP platform status",
	Long:  `Displays cluster health, ArgoCD application status, service URLs, and optionally credentials.`,
	RunE:  runFreezerStatus,
}

func init() {
	freezerStatusCmd.Flags().BoolVar(&showSecrets, "secrets", false, "Show credentials (passwords)")
	freezerCmd.AddCommand(freezerStatusCmd)
}

func runFreezerStatus(cmd *cobra.Command, args []string) error {
	// On macOS, ensure DOCKER_HOST points to Colima's socket before any
	// Docker/Kind checks.
	if prereqs.IsDarwin() && prereqs.CommandExists("colima") && colima.IsRunning() {
		if err := colima.EnsureDockerHost(); err != nil {
			fmt.Printf("  %s could not set DOCKER_HOST: %v\n", warn("[--]"), err)
		}
	}

	// Docker / Colima status
	if prereqs.IsDarwin() {
		fmt.Println(bold("Colima:"))
		if prereqs.CommandExists("colima") && colima.IsRunning() {
			fmt.Printf("  %s running\n", pass("[ok]"))
		} else if prereqs.CommandExists("colima") {
			fmt.Printf("  %s not running\n", warn("[--]"))
		} else {
			fmt.Printf("  %s not installed\n", fail("[!!]"))
		}
	} else {
		fmt.Println(bold("Docker:"))
		dockerCheck := prereqs.CheckDocker()
		switch dockerCheck.Status {
		case prereqs.StatusPass:
			fmt.Printf("  %s daemon reachable\n", pass("[ok]"))
		case prereqs.StatusWarn:
			fmt.Printf("  %s %s\n", warn("[--]"), dockerCheck.Detail)
		default:
			fmt.Printf("  %s %s\n", fail("[!!]"), dockerCheck.Detail)
		}
	}

	// Cluster status
	fmt.Println()
	fmt.Println(bold("Cluster:"))
	clusterName := cluster.GetClusterName()
	if clusterName != "" {
		fmt.Printf("  %s Kind cluster: %s\n", pass("[ok]"), clusterName)
	} else {
		fmt.Printf("  %s no Kind cluster running\n", fail("[!!]"))
		return nil
	}

	// ArgoCD Applications
	fmt.Println()
	fmt.Println(bold("ArgoCD Applications:"))
	apps, err := cluster.GetArgoApps()
	if err != nil {
		fmt.Printf("  %s could not retrieve applications: %v\n", warn("[--]"), err)
	} else if len(apps) == 0 {
		fmt.Printf("  %s no applications found\n", warn("[--]"))
	} else {
		for _, app := range apps {
			statusIcon := pass("[ok]")
			if app.Status != "Synced" || app.Health != "Healthy" {
				statusIcon = warn("[--]")
			}
			if app.Health == "Degraded" || app.Health == "Missing" {
				statusIcon = fail("[!!]")
			}
			fmt.Printf("  %s %-30s sync=%-10s health=%s\n",
				statusIcon, app.Name, app.Status, app.Health)
		}
	}

	// URLs
	fmt.Println()
	fmt.Println(bold("Service URLs:"))
	fmt.Println("  ArgoCD:     https://argocd.cnoe.localtest.me:8443")
	fmt.Println("  Gitea:      https://gitea.cnoe.localtest.me:8443")

	// Secrets
	if showSecrets {
		fmt.Println()
		fmt.Println(bold("Credentials:"))
		secrets, err := cluster.GetSecrets()
		if err != nil {
			fmt.Printf("  %s could not retrieve secrets: %v\n", warn("[--]"), err)
		} else {
			for _, s := range secrets {
				fmt.Printf("  %-40s %s\n", s.Namespace+"/"+s.Name, s.Value)
			}
		}
	} else {
		fmt.Println()
		fmt.Println("  (use --secrets to show credentials)")
	}

	return nil
}
