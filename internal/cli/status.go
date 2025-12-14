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

	configManager := config.NewManager()
	currentProvider, err := configManager.GetCurrentProvider()
	if err != nil {
		if !quiet {
			fmt.Printf("Error: Could not determine current provider: %v\n", err)
		}
		return err
	}

	settings, err := configManager.LoadSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	provider, err := providers.GetGlobalRegistry().Get(currentProvider)
	if err != nil {
		fmt.Printf("Current provider: %s (custom configuration)\n", currentProvider)
	} else {
		fmt.Printf("Current provider: %s\n", provider.DisplayName())
	}

	if !quiet {
		if err := displayConfigurationTable(settings); err != nil {
			return err
		}
		displayAPIKeyStatus(settings)
		displayProviderInfo(provider, verbose)
		displayAdditionalInfo(configManager, verbose)
	}

	return nil
}

func displayConfigurationTable(settings *config.ClaudeSettings) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "CONFIGURATION\tVALUE"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Base URL
	if baseURL, exists := settings.Env["ANTHROPIC_BASE_URL"]; exists {
		if _, err := fmt.Fprintf(w, "Base URL\t%s\n", baseURL); err != nil {
			return fmt.Errorf("failed to write base URL: %w", err)
		}
	} else {
		if _, err := fmt.Fprintln(w, "Base URL\thttps://api.anthropic.com (default)"); err != nil {
			return fmt.Errorf("failed to write default base URL: %w", err)
		}
	}

	// Models
	models := map[string]string{
		"ANTHROPIC_DEFAULT_HAIKU_MODEL":  "Haiku Model",
		"ANTHROPIC_DEFAULT_SONNET_MODEL": "Sonnet Model",
		"ANTHROPIC_DEFAULT_OPUS_MODEL":   "Opus Model",
	}
	for envVar, displayName := range models {
		if model, exists := settings.Env[envVar]; exists {
			if _, err := fmt.Fprintf(w, "%s\t%s\n", displayName, model); err != nil {
				return fmt.Errorf("failed to write %s: %w", displayName, err)
			}
		}
	}

	// API Timeout
	if timeout, exists := settings.Env["API_TIMEOUT_MS"]; exists {
		if _, err := fmt.Fprintf(w, "API Timeout\t%s ms\n", timeout); err != nil {
			return fmt.Errorf("failed to write API timeout: %w", err)
		}
	}

	return w.Flush()
}

func displayAPIKeyStatus(settings *config.ClaudeSettings) {
	if _, exists := settings.Env["ANTHROPIC_AUTH_TOKEN"]; exists {
		fmt.Printf("\nAPI Key: Configured ✓\n")
	} else {
		fmt.Printf("\nAPI Key: Not configured ✗\n")
	}
}

func displayProviderInfo(provider providers.Provider, verbose bool) {
	if provider == nil {
		return
	}

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

func displayAdditionalInfo(configManager *config.Manager, verbose bool) {
	fmt.Printf("\nSettings file: %s\n", configManager.GetSettingsPath())

	if verbose {
		backups, err := configManager.ListBackups()
		if err == nil {
			fmt.Printf("\nAvailable backups: %d\n", len(backups))
		}
	}
}