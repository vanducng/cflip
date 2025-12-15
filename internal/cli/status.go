package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current provider status",
	Long: `Display the currently active Claude provider and its configuration.
Shows the provider name, authentication method, models, and API endpoint being used.`,
	RunE: runStatus,
}

func newStatusCmd() *cobra.Command {
	return statusCmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	quiet, _ := cmd.Flags().GetBool("quiet")

	tomlManager := config.NewTOMLManagerV2()
	cfg, err := tomlManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get current provider
	provider, err := cfg.GetActiveProvider()
	if err != nil {
		if !quiet {
			fmt.Printf("Error: Could not determine current provider: %v\n", err)
		}
		return err
	}

	if !quiet {
		fmt.Printf("Current provider: %s (%s)\n", provider.DisplayName, provider.Name)
		fmt.Printf("Authentication: %s\n", getAuthMethodDisplay(provider))

		// Display configuration table
		if err := displayConfigurationTable(cfg, provider); err != nil {
			return err
		}

		// Display API key status
		displayAPIKeyStatus(provider)

		// Display active models
		displayActiveModels(cfg)

		// Display provider info
		displayProviderInfo(provider, verbose)

		// Display additional info
		displayAdditionalInfo(cfg, verbose)
	}

	return nil
}

func getAuthMethodDisplay(provider *config.ProviderInfo) string {
	if provider.IsAPIKeyRequired() {
		if provider.HasAPIKey() {
			return "API Key ✓"
		}
		return "API Key (not configured)"
	}
	return "Subscription (Claude Code CLI)"
}

func displayConfigurationTable(cfg *config.CFLIPConfig, provider *config.ProviderInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "CONFIGURATION\tVALUE"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Base URL
	if provider.IsAPIKeyRequired() {
		if _, err := fmt.Fprintf(w, "Base URL\t%s\n", provider.Auth.BaseURL); err != nil {
			return fmt.Errorf("failed to write base URL: %w", err)
		}
	} else {
		if _, err := fmt.Fprintln(w, "Base URL\tN/A (uses Claude Code CLI)"); err != nil {
			return fmt.Errorf("failed to write base URL: %w", err)
		}
	}

	// Timeout
	timeout := provider.Auth.TimeoutSeconds
	if timeout == 0 {
		timeout = 300 // Default
	}
	if _, err := fmt.Fprintf(w, "Timeout\t%d seconds\n", timeout); err != nil {
		return fmt.Errorf("failed to write timeout: %w", err)
	}

	// Last switched
	if !cfg.Active.LastSwitched.IsZero() {
		if _, err := fmt.Fprintf(w, "Last switched\t%s\n", cfg.Active.LastSwitched.Format(time.RFC3339)); err != nil {
			return fmt.Errorf("failed to write last switched: %w", err)
		}
	}

	return w.Flush()
}

func displayAPIKeyStatus(provider *config.ProviderInfo) {
	fmt.Printf("\nAuthentication Status:\n")
	if provider.IsAPIKeyRequired() {
		if provider.HasAPIKey() {
			fmt.Printf("  API Key: Configured ✓\n")
			if provider.Auth.LastValidated.After(time.Time{}) {
				fmt.Printf("  Last Validated: %s\n", provider.Auth.LastValidated.Format(time.RFC3339))
			}
		} else {
			fmt.Printf("  API Key: Not configured ✗\n")
			fmt.Printf("  To configure: cflip config set-api-key %s\n", provider.Name)
		}
	} else {
		fmt.Printf("  Method: Subscription-based\n")
		fmt.Printf("  To authenticate: claude /login\n")
	}
}

func displayActiveModels(cfg *config.CFLIPConfig) {
	fmt.Printf("\nActive Models:\n")
	if len(cfg.Active.ModelMapping) == 0 {
		fmt.Printf("  No models configured\n")
		return
	}

	for category, modelID := range cfg.Active.ModelMapping {
		if model, err := cfg.GetModelConfig(modelID); err == nil {
			fmt.Printf("  %s: %s", category, model.Name)
			if model.MaxTokens > 0 {
				fmt.Printf(" (max tokens: %d)", model.MaxTokens)
			}
			fmt.Printf("\n")
		}
	}
}

func displayProviderInfo(provider *config.ProviderInfo, verbose bool) {
	if provider.Auth.RequiresSetup {
		fmt.Printf("\nSetup Required: Yes\n")
		if verbose && provider.Auth.SetupInstructions != "" {
			fmt.Printf("\nSetup Instructions:\n%s\n", provider.Auth.SetupInstructions)
		}
	}

	if verbose {
		if len(provider.Tags) > 0 {
			fmt.Printf("\nTags: %v\n", provider.Tags)
		}
		if provider.Website != "" {
			fmt.Printf("Website: %s\n", provider.Website)
		}
	}
}

func displayAdditionalInfo(cfg *config.CFLIPConfig, verbose bool) {
	fmt.Printf("\nConfiguration file: %s\n", config.GetConfigPath())
	fmt.Printf("Claude settings file: %s\n", config.GetLegacySettingsPath())

	if verbose {
		fmt.Printf("\nSettings:\n")
		fmt.Printf("  Backup directory: %s\n", cfg.Settings.BackupDirectory)
		fmt.Printf("  Max backups: %d\n", cfg.Settings.MaxBackups)
		fmt.Printf("  Auto backup: %t\n", cfg.Settings.AutoBackup)
		fmt.Printf("  Secure storage: %t\n", cfg.Settings.SecureStorage)

		// List all configured providers
		fmt.Printf("\nConfigured Providers: %d\n", len(cfg.Providers))
		for name := range cfg.Providers {
			marker := ""
			if name == cfg.Active.Provider {
				marker = " (active)"
			}
			fmt.Printf("  • %s%s\n", name, marker)
		}
	}
}