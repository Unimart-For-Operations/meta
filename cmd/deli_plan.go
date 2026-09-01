package cmd

import (
	"fmt"
	"strings"

	"github.com/Unimart-For-Operations/meta/internal/host"
	"github.com/spf13/cobra"
)

var deliPlanCmd = &cobra.Command{
	Use:   "plan [host]",
	Short: "Show the non-destructive onboarding plan for a host",
	Long: `Show the physical-host onboarding plan for a target host.

This command does not mutate disks, install packages, or apply configuration.
It translates cmdr host metadata into a concrete provisioning handoff that an
external installer device can target.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDeliPlan,
}

func init() {
	deliCmd.AddCommand(deliPlanCmd)
}

type planStep struct {
	Title string
	Items []string
}

type hostPlan struct {
	Host  host.Info
	Steps []planStep
}

func runDeliPlan(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	var target *host.Info
	if len(args) > 0 {
		target, err = host.GetHost(dir, args[0])
	} else {
		target, err = host.Detect(dir)
	}
	if err != nil {
		return fmt.Errorf("resolve plan target: %w\n\nRun: unimart deli hosts", err)
	}

	printHostPlan(buildHostPlan(*target))
	return nil
}

func buildHostPlan(h host.Info) hostPlan {
	steps := []planStep{
		{
			Title: "Phase 0 — Prepare Installer Media",
			Items: []string{
				"Boot the provisioner USB (NixOS minimal live ISO)",
				"Confirm network access and target architecture",
				"Have Git/SSH credentials ready for private repo access",
			},
		},
		{
			Title: "Phase 1 — Install Base OS",
			Items: []string{
				"Partition, encrypt, and mount the target disk",
				"Install base system, bootloader, networking, sudo, and initial user",
				"Reboot into the installed OS before applying desired state",
			},
		},
		{
			Title: "Phase 2 — Clone Control Plane",
			Items: []string{
				"git clone --recurse-submodules https://github.com/Unimart-For-Operations/meta.git",
				"cd meta",
				"make init",
			},
		},
		{
			Title: "Phase 3 — Converge Host",
			Items: []string{
				fmt.Sprintf("unimart deli switch %s", h.Name),
				"unimart deli doctor",
				"unimart stockroom check",
			},
		},
	}

	if hasCapability(h, "operator") {
		steps = append(steps, planStep{
			Title: "Phase 4 — Validate Operator Tooling",
			Items: []string{
				"unimart freezer doctor",
				"kubectl version --client",
				"kind version",
			},
		})
	}

	if hasCapability(h, "idp-local") {
		steps = append(steps, planStep{
			Title: "Phase 5 — Stand Up Local IDP",
			Items: []string{
				"unimart open --no-browser",
				"unimart freezer status",
			},
		})
	}

	if hasCapability(h, "desktop") {
		steps = append(steps, planStep{
			Title: "Desktop Check",
			Items: []string{
				"Start or log into the graphical session",
				"Confirm terminal, editor, browser, and Obsidian if available",
			},
		})
	}

	return hostPlan{Host: h, Steps: steps}
}

func printHostPlan(plan hostPlan) {
	h := plan.Host
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("%s\n", bold("Physical Host Provisioning Plan"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Printf("  Host:         %s\n", cyan(h.Name))
	fmt.Printf("  Platform:     %s\n", displayValue(h.Platform))
	fmt.Printf("  User:         %s\n", displayValue(h.Username))
	fmt.Printf("  Role:         %s\n", displayValue(h.Role))
	fmt.Printf("  Capabilities: %s\n", displayList(h.Capabilities))
	fmt.Println()
	fmt.Printf("%s This is a dry-run plan. No disks, packages, or configs were changed.\n", warn("[dry-run]"))

	for _, step := range plan.Steps {
		fmt.Println()
		fmt.Printf("%s\n", bold(step.Title))
		for _, item := range step.Items {
			fmt.Printf("  - %s\n", item)
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

func hasCapability(h host.Info, capability string) bool {
	for _, c := range h.Capabilities {
		if strings.EqualFold(c, capability) {
			return true
		}
	}
	return false
}
