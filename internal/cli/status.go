package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"github.com/vanducng/cflip/internal/providers"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current provider status",
	Long: `Display the currently active Claude provider and its configuration.
Shows the provider name, models, and API endpoint being used.`,
	RunE: runStatus,
}

func newStatusCmd() *cobra.Command {
	return statusCmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	quiet, _ := cmd.Flags().GetBool("quiet")

	// Get current provider
	configManager := config.NewManager()
	currentProvider, err := configManager.GetCurrentProvider()
	if err != nil {
		if !quiet {
			fmt.Printf("Error: Could not determine current provider: %v\n", err)
		}
		return err
	}

	// Load current settings
	settings, err := configManager.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Get provider details
	registry := providers.GetGlobalRegistry()
	provider, err := registry.Get(currentProvider)
	if err != nil {
		fmt.Printf("Current provider: %s (custom configuration)\n", currentProvider)
	} else {
		fmt.Printf("Current provider: %s\n", provider.DisplayName())
	}

	if !quiet {
		// Create tabwriter for nice formatting
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "CONFIGURATION\tVALUE")

		// Base URL
		if baseURL, exists := settings.Env["ANTHROPIC_BASE_URL"]; exists {
			fmt.Fprintf(w, "Base URL\t%s\n", baseURL)
		} else {
			fmt.Fprintln(w, "Base URL\thttps://api.anthropic.com (default)")
		}

		// Models
		if haiku, exists := settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"]; exists {
			fmt.Fprintf(w, "Haiku Model\t%s\n", haiku)
		}
		if sonnet, exists := settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"]; exists {
			fmt.Fprintf(w, "Sonnet Model\t%s\n", sonnet)
		}
		if opus, exists := settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"]; exists {
			fmt.Fprintf(w, "Opus Model\t%s\n", opus)
		}

		// API Timeout
		if timeout, exists := settings.Env["API_TIMEOUT_MS"]; exists {
			fmt.Fprintf(w, "API Timeout\t%s ms\n", timeout)
		}

		w.Flush()

		// Check if API key is set
		if _, exists := settings.Env["ANTHROPIC_AUTH_TOKEN"]; exists {
			fmt.Printf("\nAPI Key: Configured ✓\n")
		} else {
			fmt.Printf("\nAPI Key: Not configured ✗\n")
		}

		// Provider-specific information
		if provider != nil {
			if provider.RequiresSetup() {
				fmt.Printf("\nSetup required: Yes\n")
				if verbose {
					fmt.Printf("\nSetup Instructions:\n%s\n", provider.SetupInstructions())
				}
			}

			// Show provider features if GLM
			if provider.Name() == "glm" {
				if glmProvider, ok := provider.(*providers.GLMProvider); ok {
					fmt.Printf("\nAvailable Features:\n")
					for _, feature := range glmProvider.GetFeatureList() {
						fmt.Printf("  • %s\n", feature)
					}
				}
			}
		}

		// Settings file location
		fmt.Printf("\nSettings file: %s\n", configManager.GetSettingsPath())

		// Backup information
		if verbose {
			backups, err := configManager.ListBackups()
			if err == nil {
				fmt.Printf("\nAvailable backups: %d\n", len(backups))
			}
		}
	}

	return nil
}