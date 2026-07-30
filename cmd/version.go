package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Version info — injected via ldflags at build time.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := color.New(color.FgCyan, color.Bold).SprintFunc()
		ver := color.New(color.Bold).SprintFunc()
		label := color.New(color.FgHiBlack).SprintFunc()
		fmt.Printf("%s %s\n", name("unimart"), ver(Version))
		fmt.Printf("  %s %s\n", label("commit:"), GitCommit)
		fmt.Printf("  %s %s\n", label("built: "), BuildDate)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
