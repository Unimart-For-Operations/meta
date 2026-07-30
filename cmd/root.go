package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	// Flags
	orgDir  string
	verbose bool

	// Colors (shared across all commands)
	pass = color.New(color.FgGreen).SprintFunc()
	fail = color.New(color.FgRed).SprintFunc()
	warn = color.New(color.FgYellow).SprintFunc()
	bold = color.New(color.Bold).SprintFunc()
	cyan = color.New(color.FgCyan).SprintFunc()
	dim  = color.New(color.FgHiBlack).SprintFunc()
)

var rootCmd = &cobra.Command{
	Use:           "unimart",
	Short:         "idpbuilder organization CLI — your one-stop shop",
	Long:          "", // set in init()
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.Long = buildRootLong()
	rootCmd.SetHelpTemplate(helpTemplate())

	rootCmd.PersistentFlags().StringVar(&orgDir, "org-dir", "", "path to the idpbuilder org directory (default: auto-detect)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

// buildRootLong returns the colored long description for the root command.
func buildRootLong() string {
	aisle := color.New(color.FgCyan, color.Bold).SprintFunc()
	desc := color.New(color.FgHiBlack).SprintFunc()
	title := color.New(color.Bold).SprintFunc()

	return title("unimart") + " — your one-stop shop for the idpbuilder organization.\n\n" +
		title("Browse the aisles:") + "\n" +
		"  " + aisle("deli") + "        " + desc("Workstation configuration (Nix/Home Manager)") + "\n" +
		"  " + aisle("freezer") + "     " + desc("IDP platform lifecycle (clusters, builds, repos)") + "\n" +
		"  " + aisle("stockroom") + "   " + desc("Cross-repo validation (contract checks)")
}

// helpTemplate returns a cobra help template with colored section headers.
func helpTemplate() string {
	h := color.New(color.Bold).SprintFunc()
	hint := color.New(color.FgHiBlack).SprintFunc()

	return `{{with .Long}}{{. | trimRightSpace}}

{{end}}` + h("Usage:") + `
  {{.UseLine}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

` + h("Aliases:") + `
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

` + h("Examples:") + `
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

` + h("Available Commands:") + `{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

` + h("Flags:") + `
{{.LocalFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasAvailableInheritedFlags}}

` + h("Global Flags:") + `
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

` + h("Additional Help Topics:") + `{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

` + hint(`Use "{{.CommandPath}} [command] --help" for more information about a command.`) + `
{{end}}`
}

// Execute runs the root command. This is the only exported function.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s %s\n", fail("[fail]"), err)
		return err
	}
	return nil
}

// resolveOrgDir returns the org directory, resolving from flag, env, or auto-detection.
func resolveOrgDir() (string, error) {
	// 1. --org-dir flag
	if orgDir != "" {
		return filepath.Abs(orgDir)
	}

	// 2. UNIMART_ORG_DIR environment variable
	if env := os.Getenv("UNIMART_ORG_DIR"); env != "" {
		return filepath.Abs(env)
	}

	// 3. Auto-detect: walk up from CWD looking for the meta repo markers
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	dir := cwd
	for {
		if isOrgDir(dir) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not detect org directory (no .gitmodules with cmdr submodule found)\nUse --org-dir or set UNIMART_ORG_DIR")
}

// isOrgDir checks if a directory looks like the meta repo root.
func isOrgDir(dir string) bool {
	// Must have .gitmodules and cmdr/ submodule
	gitmodules := filepath.Join(dir, ".gitmodules")
	if _, err := os.Stat(gitmodules); err != nil {
		return false
	}
	cmdrDir := filepath.Join(dir, "cmdr", "flake.nix")
	if _, err := os.Stat(cmdrDir); err != nil {
		return false
	}
	return true
}

// idpbuilderDir returns the path to the idpbuilder repo within the org dir.
func idpbuilderDir() string {
	dir, err := resolveOrgDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "idpbuilder")
}

// mustResolveOrgDir returns the org directory or exits with an error.
func mustResolveOrgDir() (string, error) {
	return resolveOrgDir()
}

// promptString reads a line from stdin after printing the given prompt.
func promptString(prompt string) string {
	r := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	s, _ := r.ReadString('\n')
	return strings.TrimSpace(s)
}

// promptYesNo asks a yes/no question and returns the answer.
func promptYesNo(prompt string, defaultYes bool) bool {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(prompt)
		s, _ := r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return defaultYes
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
		fmt.Println("Please answer y or n")
	}
}
