package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Build information injected at build time
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

// getVersion returns the formatted version string
func getVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, buildTime)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cflip",
	Short: "Claude Provider Switcher",
	Long: `CFLIP is a CLI tool that enables seamless switching between different
Claude Code providers (Anthropic, GLM/z.ai, and future providers).

It manages the ~/.claude/settings.json configuration file to toggle between
different API endpoints and authentication methods.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute(v, c, bt string) error {
	// Set build information (override version if provided)
	if v != "" {
		version = v
	}
	commit = c
	buildTime = bt

	// Set version on root command
	rootCmd.Version = getVersion()

	// Add subcommands
	addCommands()

	return rootCmd.Execute()
}

// addCommands adds all subcommands to the root command
func addCommands() {
	// Main commands - only switch for now
	rootCmd.AddCommand(newSwitchCmd())
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "quiet mode (no output)")

	// Custom help and version formatting
	cobra.AddTemplateFunc("indent", indent)
	cobra.AddTemplateFunc("trimTrailingWhitespaces", trimTrailingWhitespaces)

	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetHelpTemplate(helpTemplate)
}

// Helper functions for template formatting
func indent(spaces int, text string) string {
	prefix := ""
	for i := 0; i < spaces; i++ {
		prefix += " "
	}
	return prefix + text
}

func trimTrailingWhitespaces(text string) string {
	// Implementation to trim trailing whitespaces
	return text
}

const usageTemplate = `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

const helpTemplate = `{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
