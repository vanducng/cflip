package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"golang.org/x/term"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CFLIP configuration",
	Long: `Manage CFLIP configuration settings, providers, and models.

Subcommands:
  show        - Show current configuration
  set-provider - Set active provider
  set-model   - Set active model for category
  list-models - List available models
  list-providers - List available providers
  set-api-key - Set API key for provider
  settings    - Manage global settings`,
}

func newConfigCmd() *cobra.Command {
	configCmd.AddCommand(newConfigShowCmd())
	configCmd.AddCommand(newConfigSetProviderCmd())
	configCmd.AddCommand(newConfigSetModelCmd())
	configCmd.AddCommand(newConfigListModelsCmd())
	configCmd.AddCommand(newConfigListProvidersCmd())
	configCmd.AddCommand(newConfigSetAPIKeyCmd())
	configCmd.AddCommand(newConfigSettingsCmd())
	return configCmd
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	RunE: runConfigShow,
}

func newConfigShowCmd() *cobra.Command {
	configShowCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	configShowCmd.Flags().BoolP("models", "m", false, "Show model details")
	configShowCmd.Flags().BoolP("all", "a", false, "Show all configuration details")
	return configShowCmd
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")
	showModels, _ := cmd.Flags().GetBool("models")
	showAll, _ := cmd.Flags().GetBool("all")

	tomlManager := config.NewTOMLManagerV2()
	cfg, err := tomlManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if jsonOutput {
		// TODO: Implement JSON output
		fmt.Printf("JSON output not yet implemented\n")
		return nil
	}

	// Show active configuration
	fmt.Printf("CFLIP Configuration\n")
	fmt.Printf("===================\n\n")

	// Active provider
	fmt.Printf("Active Provider: %s\n", cfg.Active.Provider)
	if provider, err := cfg.GetActiveProvider(); err == nil {
		fmt.Printf("  Name: %s\n", provider.DisplayName)
		fmt.Printf("  Description: %s\n", provider.Description)
		fmt.Printf("  Authentication: %s\n", provider.Auth.Method)
		if provider.IsAPIKeyRequired() {
			if provider.HasAPIKey() {
				fmt.Printf("  API Key: Configured ✓\n")
			} else {
				fmt.Printf("  API Key: Not configured ✗\n")
			}
		}
	}

	// Active models
	fmt.Printf("\nActive Models:\n")
	for category, modelID := range cfg.Active.ModelMapping {
		if model, err := cfg.GetModelConfig(modelID); err == nil {
			fmt.Printf("  %s: %s (%s)\n", category, model.Name, model.ID)
		}
	}

	if showModels || showAll {
		fmt.Printf("\nAvailable Models:\n")
		for _, model := range cfg.Models {
			fmt.Printf("  %s - %s\n", model.ID, model.Name)
			fmt.Printf("    Provider: %s | Category: %s\n", model.Provider, model.Category)
			if showAll {
				fmt.Printf("    Description: %s\n", model.Description)
				if model.MaxTokens > 0 {
					fmt.Printf("    Max Tokens: %d\n", model.MaxTokens)
				}
				if len(model.Capabilities) > 0 {
					fmt.Printf("    Capabilities: %s\n", strings.Join(model.Capabilities, ", "))
				}
			}
		}
	}

	if showAll {
		fmt.Printf("\nProviders:\n")
		for _, provider := range cfg.Providers {
			fmt.Printf("  %s - %s\n", provider.Name, provider.DisplayName)
			fmt.Printf("    Authentication: %s\n", provider.Auth.Method)
			if len(provider.Models) > 0 {
				fmt.Printf("    Models: %s\n", strings.Join(provider.Models, ", "))
			}
		}

		fmt.Printf("\nSettings:\n")
		fmt.Printf("  Config File: %s\n", config.GetConfigPath())
		fmt.Printf("  Backup Directory: %s\n", cfg.Settings.BackupDirectory)
		fmt.Printf("  Max Backups: %d\n", cfg.Settings.MaxBackups)
		fmt.Printf("  Auto Backup: %t\n", cfg.Settings.AutoBackup)
		fmt.Printf("  Secure Storage: %t\n", cfg.Settings.SecureStorage)
	}

	return nil
}

// configSetProviderCmd represents the config set-provider command
var configSetProviderCmd = &cobra.Command{
	Use:   "set-provider <name>",
	Short: "Set active provider",
	Args:  cobra.ExactArgs(1),
	RunE: runConfigSetProvider,
}

func newConfigSetProviderCmd() *cobra.Command {
	return configSetProviderCmd
}

func runConfigSetProvider(cmd *cobra.Command, args []string) error {
	tomlManager := config.NewTOMLManagerV2()

	providerName := args[0]

	// Validate provider exists
	_, err := tomlManager.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// Set as active
	if err := tomlManager.SetActiveProvider(providerName); err != nil {
		return fmt.Errorf("failed to set active provider: %w", err)
	}

	fmt.Printf("✓ Active provider set to: %s\n", providerName)

	// Generate Claude settings
	if err := generateClaudeSettings(tomlManager, providerName); err != nil {
		fmt.Printf("Warning: Failed to generate Claude settings: %v\n", err)
	}

	return nil
}

// configSetModelCmd represents the config set-model command
var configSetModelCmd = &cobra.Command{
	Use:   "set-model <category> <model-id>",
	Short: "Set active model for category",
	Args:  cobra.ExactArgs(2),
	RunE: runConfigSetModel,
}

func newConfigSetModelCmd() *cobra.Command {
	return configSetModelCmd
}

func runConfigSetModel(cmd *cobra.Command, args []string) error {
	category := args[0]
	modelID := args[1]

	tomlManager := config.NewTOMLManagerV2()

	// Validate category
	validCategories := []string{"haiku", "sonnet", "opus"}
	isValidCategory := false
	for _, cat := range validCategories {
		if category == cat {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return fmt.Errorf("invalid category '%s'. Valid categories: %s",
			category, strings.Join(validCategories, ", "))
	}

	// Set model
	if err := tomlManager.SetActiveModel(category, modelID); err != nil {
		return fmt.Errorf("failed to set active model: %w", err)
	}

	fmt.Printf("✓ Active model for %s set to: %s\n", category, modelID)

	return nil
}

// configListModelsCmd represents the config list-models command
var configListModelsCmd = &cobra.Command{
	Use:   "list-models [provider]",
	Short: "List available models",
	Args:  cobra.MaximumNArgs(1),
	RunE: runConfigListModels,
}

func newConfigListModelsCmd() *cobra.Command {
	configListModelsCmd.Flags().StringP("category", "c", "", "Filter by category")
	return configListModelsCmd
}

func runConfigListModels(cmd *cobra.Command, args []string) error {
	category, _ := cmd.Flags().GetString("category")

	tomlManager := config.NewTOMLManagerV2()

	var models []config.ModelConfig
	var err error

	if len(args) > 0 {
		// List models for specific provider
		providerName := args[0]
		models, err = tomlManager.GetModelsByProvider(providerName)
		if err != nil {
			return fmt.Errorf("failed to get models for provider '%s': %w", providerName, err)
		}
		fmt.Printf("Models for provider: %s\n", providerName)
	} else if category != "" {
		// List models by category
		models, err = tomlManager.GetModelsByCategory(category)
		if err != nil {
			return fmt.Errorf("failed to get models for category '%s': %w", category, err)
		}
		fmt.Printf("Models in category: %s\n", category)
	} else {
		// List all models
		allModels, err := tomlManager.ListModels()
		if err != nil {
			return fmt.Errorf("failed to list models: %w", err)
		}
		for _, model := range allModels {
			models = append(models, model)
		}
		fmt.Printf("All Available Models\n")
	}

	fmt.Printf("\n")
	for _, model := range models {
		active := ""
		cfg, _ := tomlManager.LoadConfig()
		if cfg.Active.ModelMapping[model.Category] == model.ID {
			active = " [ACTIVE]"
		}

		fmt.Printf("  %s%s\n", model.ID, active)
		fmt.Printf("    Name: %s\n", model.Name)
		fmt.Printf("    Provider: %s | Category: %s\n", model.Provider, model.Category)
		fmt.Printf("    Description: %s\n", model.Description)
		if model.MaxTokens > 0 {
			fmt.Printf("    Max Tokens: %d | Context Window: %d\n", model.MaxTokens, model.ContextWindow)
		}
		if len(model.Capabilities) > 0 {
			fmt.Printf("    Capabilities: %s\n", strings.Join(model.Capabilities, ", "))
		}
		fmt.Printf("\n")
	}

	return nil
}

// configListProvidersCmd represents the config list-providers command
var configListProvidersCmd = &cobra.Command{
	Use:   "list-providers",
	Short: "List available providers",
	RunE: runConfigListProviders,
}

func newConfigListProvidersCmd() *cobra.Command {
	return configListProvidersCmd
}

func runConfigListProviders(cmd *cobra.Command, args []string) error {
	tomlManager := config.NewTOMLManagerV2()

	cfg, err := tomlManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Printf("Available Providers\n")
	fmt.Printf("==================\n\n")

	for _, provider := range cfg.Providers {
		current := ""
		if cfg.Active.Provider == provider.Name {
			current = " [CURRENT]"
		}

		fmt.Printf("%s%s\n", provider.DisplayName, current)
		fmt.Printf("  Name: %s\n", provider.Name)
		fmt.Printf("  Description: %s\n", provider.Description)
		fmt.Printf("  Authentication: %s\n", provider.Auth.Method)

		if provider.IsAPIKeyRequired() {
			if provider.HasAPIKey() {
				fmt.Printf("  API Key: Configured ✓\n")
			} else {
				fmt.Printf("  API Key: Not configured ✗\n")
			}
		} else {
			fmt.Printf("  Uses Claude subscription\n")
		}

		if len(provider.Models) > 0 {
			fmt.Printf("  Available Models: %d\n", len(provider.Models))
		}

		if len(provider.Tags) > 0 {
			fmt.Printf("  Tags: %s\n", strings.Join(provider.Tags, ", "))
		}

		fmt.Printf("\n")
	}

	return nil
}

// configSetAPIKeyCmd represents the config set-api-key command
var configSetAPIKeyCmd = &cobra.Command{
	Use:   "set-api-key <provider>",
	Short: "Set API key for provider",
	Args:  cobra.ExactArgs(1),
	RunE: runConfigSetAPIKey,
}

func newConfigSetAPIKeyCmd() *cobra.Command {
	return configSetAPIKeyCmd
}

func runConfigSetAPIKey(cmd *cobra.Command, args []string) error {
	providerName := args[0]

	tomlManager := config.NewTOMLManagerV2()

	provider, err := tomlManager.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	if !provider.IsAPIKeyRequired() {
		return fmt.Errorf("provider '%s' does not use API key authentication", providerName)
	}

	fmt.Printf("Setting API key for %s\n", provider.DisplayName)

	// Prompt for API key
	fmt.Printf("Enter API key: ")
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}
	fmt.Println() // New line after password input

	apiKey := strings.TrimSpace(string(bytePassword))
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Set API key
	provider.SetAPIKey(apiKey)

	if err := tomlManager.SaveProvider(providerName, provider); err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}

	fmt.Printf("✓ API key configured for %s\n", provider.DisplayName)

	// Regenerate Claude settings if this is the active provider
	cfg, _ := tomlManager.LoadConfig()
	if cfg.Active.Provider == providerName {
		if err := generateClaudeSettings(tomlManager, providerName); err != nil {
			fmt.Printf("Warning: Failed to update Claude settings: %v\n", err)
		}
	}

	return nil
}

// configSettingsCmd represents the config settings command
var configSettingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Manage global settings",
	RunE: runConfigSettings,
}

func newConfigSettingsCmd() *cobra.Command {
	configSettingsCmd.Flags().StringP("backup-dir", "b", "", "Set backup directory")
	configSettingsCmd.Flags().IntP("max-backups", "m", 0, "Set maximum backups to keep")
	configSettingsCmd.Flags().BoolP("auto-backup", "a", false, "Enable/disable automatic backups")
	configSettingsCmd.Flags().BoolP("secure-storage", "s", false, "Enable/disable secure API key storage")
	return configSettingsCmd
}

func runConfigSettings(cmd *cobra.Command, args []string) error {
	backupDir, _ := cmd.Flags().GetString("backup-dir")
	maxBackups, _ := cmd.Flags().GetInt("max-backups")
	autoBackup, _ := cmd.Flags().GetBool("auto-backup")
	secureStorage, _ := cmd.Flags().GetBool("secure-storage")

	tomlManager := config.NewTOMLManagerV2()

	settings, err := tomlManager.GetSettings()
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	updated := false

	if backupDir != "" {
		settings.BackupDirectory = backupDir
		updated = true
		fmt.Printf("✓ Backup directory set to: %s\n", backupDir)
	}

	if maxBackups > 0 {
		settings.MaxBackups = maxBackups
		updated = true
		fmt.Printf("✓ Maximum backups set to: %d\n", maxBackups)
	}

	if cmd.Flags().Changed("auto-backup") {
		settings.AutoBackup = autoBackup
		updated = true
		fmt.Printf("✓ Auto backup %s\n", map[bool]string{true: "enabled", false: "disabled"}[autoBackup])
	}

	if cmd.Flags().Changed("secure-storage") {
		settings.SecureStorage = secureStorage
		updated = true
		fmt.Printf("✓ Secure storage %s\n", map[bool]string{true: "enabled", false: "disabled"}[secureStorage])
	}

	if updated {
		if err := tomlManager.UpdateSettings(*settings); err != nil {
			return fmt.Errorf("failed to update settings: %w", err)
		}
	} else {
		// Show current settings
		fmt.Printf("Current Settings\n")
		fmt.Printf("================\n")
		fmt.Printf("Backup Directory: %s\n", settings.BackupDirectory)
		fmt.Printf("Max Backups: %d\n", settings.MaxBackups)
		fmt.Printf("Auto Backup: %t\n", settings.AutoBackup)
		fmt.Printf("Secure Storage: %t\n", settings.SecureStorage)
		fmt.Printf("Default Timeout: %d seconds\n", settings.DefaultTimeout)
		fmt.Printf("Auto Validate: %t\n", settings.AutoValidate)
		fmt.Printf("Log Level: %s\n", settings.LogLevel)
	}

	return nil
}