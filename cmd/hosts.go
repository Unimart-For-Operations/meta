package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Unimart-For-Operations/meta/internal/host"
	"github.com/spf13/cobra"
)

var hostsCmd = &cobra.Command{
	Use:   "hosts",
	Short: "List available host configurations",
	RunE:  runHosts,
}

func init() {
	deliCmd.AddCommand(hostsCmd)
}

func runHosts(cmd *cobra.Command, args []string) error {
	dir, err := resolveOrgDir()
	if err != nil {
		return err
	}

	hosts, err := host.ListHosts(dir)
	if err != nil {
		return fmt.Errorf("list hosts: %w", err)
	}

	if len(hosts) == 0 {
		fmt.Println("No host configurations found.")
		fmt.Println("Run: unimart deli bootstrap  to set up this machine.")
		return nil
	}

	// Detect current host for highlighting
	current, _ := host.Detect(dir)

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n", bold("HOST"), bold("PLATFORM"), bold("USER"), bold("ROLE"), bold("CAPABILITIES"))
	for _, h := range hosts {
		marker := "  "
		if current != nil && h.Name == current.Name {
			marker = cyan("→ ")
		}
		fmt.Fprintf(w, "%s%s\t%s\t%s\t%s\t%s\t\n", marker, h.Name, h.Platform, h.Username, displayValue(h.Role), displayList(h.Capabilities))
	}
	w.Flush()

	return nil
}

func displayValue(value string) string {
	if value == "" {
		return "-"
	}
	return value
}

func displayList(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
