package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"golang.org/x/term"
)

// getLegacyAPIKey retrieves API key from legacy settings
func getLegacyAPIKey() string {
	legacyPath := config.GetLegacySettingsPath()
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return ""
	}

	// Try to read legacy settings
	legacyManager := config.NewManager()
	settings, err := legacyManager.LoadSettings()
	if err != nil {
		return ""
	}

	return settings.Env["ANTHROPIC_AUTH_TOKEN"]
}

// validateAPIKeyFormat validates API key format for different providers
func validateAPIKeyFormat(providerName, apiKey string) error {
	switch providerName {
	case "anthropic":
		if !strings.HasPrefix(apiKey, "sk-ant-") {
			return fmt.Errorf("Anthropic API keys usually start with 'sk-ant-'")
		}
		if len(apiKey) < 50 {
			return fmt.Errorf("API key appears to be too short")
		}
	case "glm":
		if !strings.HasPrefix(apiKey, "zai-") {
			return fmt.Errorf("GLM API keys usually start with 'zai-'")
		}
		if len(apiKey) < 40 {
			return fmt.Errorf("API key appears to be too short")
		}
	}
	return nil
}

// onboardCmd represents the onboard command
var onboardCmd = &cobra.Command{
	Use:   "onboard",
	Short: "Interactive setup for CFLIP configuration",
	Long: `Interactive setup wizard that helps you configure CFLIP for the first time.
It will guide you through:
1. Choosing your preferred Claude provider
2. Setting up authentication (API key or subscription)
3. Configuring model preferences
4. Setting up backup preferences

This command is typically run once after installing CFLIP.`,
	RunE: runOnboard,
}

func newOnboardCmd() *cobra.Command {
	return onboardCmd
}

func runOnboard(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	quiet, _ := cmd.Flags().GetBool("quiet")

	if !quiet {
		printWelcome()
	}

	tomlManager := config.NewTOMLManagerV2()
	cfg, err := tomlManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if already configured
	if isAlreadyConfigured(cfg) && !quiet {
		if !promptReconfigure() {
			fmt.Println("Onboarding cancelled. Your configuration remains unchanged.")
			return nil
		}
	}

	// Step 1: Choose provider
	providerName, err := chooseProvider(cfg, verbose)
	if err != nil {
		return err
	}

	// Step 2: Configure provider
	provider := cfg.Providers[providerName]
	if provider.IsAPIKeyRequired() {
		if err := configureAPIKeyProvider(&provider, verbose, quiet); err != nil {
			return err
		}
	} else {
		configureSubscriptionProvider(&provider, verbose, quiet)
	}

	// Update provider in config
	cfg.Providers[providerName] = provider

	// Step 3: Configure active models
	if err := configureActiveModels(tomlManager, cfg, providerName, verbose); err != nil {
		return err
	}

	// Step 4: Configure settings
	if err := configureSettings(tomlManager, &cfg.Settings, verbose); err != nil {
		return err
	}

	// Save configuration
	if err := tomlManager.SetActiveProvider(providerName); err != nil {
		return fmt.Errorf("failed to set active provider: %w", err)
	}

	if err := tomlManager.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Step 5: Test configuration
	if !quiet {
		fmt.Printf("\n✓ Configuration saved successfully!\n")
		if promptTestConnection(&provider) {
			if err := testProviderConnection(&provider); err != nil {
				fmt.Printf("⚠ Warning: Connection test failed: %v\n", err)
				fmt.Printf("  You may need to check your API key or network connection.\n")
			} else {
				fmt.Printf("✓ Connection test successful!\n")
			}
		}
		printNextSteps(&provider)
	}

	return nil
}

func printWelcome() {
	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║                    Welcome to CFLIP!                         ║
║              Claude Provider Switcher for CLI                 ║
╚══════════════════════════════════════════════════════════════╝

This wizard will help you configure CFLIP to work with your preferred
Claude provider. Let's get started!

`)
}

func isAlreadyConfigured(cfg *config.CFLIPConfig) bool {
	// Check if any provider has an API key configured
	for _, provider := range cfg.Providers {
		if provider.IsAPIKeyRequired() && provider.HasAPIKey() {
			return true
		}
	}
	return false
}

func promptReconfigure() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Configuration already exists. Would you like to reconfigure? (y/N): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func chooseProvider(cfg *config.CFLIPConfig, verbose bool) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nChoose your Claude provider:\n")

	providerNames := cfg.ListProviders()

	for i, name := range providerNames {
		provider := cfg.Providers[name]
		if verbose {
			fmt.Printf("  %d) %s - %s", len(providerNames), provider.DisplayName, provider.Description)
			if provider.IsAPIKeyRequired() {
				fmt.Printf(" (requires API key)")
			} else {
				fmt.Printf(" (uses Claude subscription)")
			}
			fmt.Printf("\n")
		} else {
			fmt.Printf("  %d) %s\n", i+1, provider.DisplayName)
		}
	}

	for {
		fmt.Printf("\nSelect provider (1-%d): ", len(providerNames))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if choice, err := strconv.Atoi(input); err == nil {
			if choice >= 1 && choice <= len(providerNames) {
				return providerNames[choice-1], nil
			}
		}

		// Check if input matches provider name
		for _, name := range providerNames {
			if strings.EqualFold(input, name) {
				return name, nil
			}
		}

		fmt.Printf("Invalid selection. Please try again.\n")
	}
}

func configureAPIKeyProvider(provider *config.ProviderInfo, verbose, quiet bool) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\nAPI Key Configuration for %s\n", provider.DisplayName)
	fmt.Printf("----------------------------------------\n")

	// Check if we can import from legacy settings
	if existingKey := getLegacyAPIKey(); existingKey != "" {
		fmt.Printf("Found existing API key in ~/.claude/settings.json\n")
		fmt.Print("Use existing API key? (Y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" || input == "y" || input == "yes" {
			provider.SetAPIKey(existingKey)
			if !quiet {
				fmt.Printf("✓ Imported existing API key\n")
			}
			return nil
		}
	}

	fmt.Printf("Setup Instructions:\n")
	fmt.Printf("1. Get your API key from %s\n", provider.Website)
	if provider.Name == "anthropic" {
		fmt.Printf("2. Ensure you have credits or a subscription\n")
	}
	fmt.Printf("3. Enter your API key when prompted\n\n")

	for {
		fmt.Printf("Enter %s API key: ", provider.DisplayName)
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("failed to read API key: %w", err)
		}
		fmt.Println() // New line after password input

		apiKey := strings.TrimSpace(string(bytePassword))
		if apiKey == "" {
			fmt.Printf("API key cannot be empty. Please try again.\n")
			continue
		}

		// Basic validation
		if err := validateAPIKeyFormat(provider.Name, apiKey); err != nil && verbose {
			fmt.Printf("Warning: %v\n", err)
			fmt.Print("Continue anyway? (y/N): ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))
			if input != "y" && input != "yes" {
				continue
			}
		}

		provider.SetAPIKey(apiKey)
		if !quiet {
			fmt.Printf("✓ API key configured\n")
		}
		break
	}

	return nil
}

func configureSubscriptionProvider(provider *config.ProviderInfo, verbose, quiet bool) {
	fmt.Printf("\nSubscription Authentication for %s\n", provider.DisplayName)
	fmt.Printf("-------------------------------------------\n")
	fmt.Printf("This provider uses your Claude subscription for authentication.\n\n")

	if provider.Auth.SetupInstructions != "" {
		fmt.Printf("Setup Instructions:\n%s\n\n", provider.Auth.SetupInstructions)
	}

	fmt.Printf("Note: Make sure you have:\n")
	fmt.Printf("  • An active Claude Pro or Max subscription\n")
	fmt.Printf("  • Run 'claude /login' to authenticate\n")
	fmt.Printf("  • Run 'claude /whoami' to verify your login\n\n")

	if !quiet {
		fmt.Printf("✓ Subscription provider configured\n")
		fmt.Printf("  Remember to run 'claude /login' before using Claude!\n")
	}
}

func configureActiveModels(tomlManager *config.TOMLManagerV2, cfg *config.CFLIPConfig, providerName string, verbose bool) error {
	reader := bufio.NewReader(os.Stdin)

	provider := cfg.Providers[providerName]

	fmt.Printf("\nModel Configuration\n")
	fmt.Printf("-------------------\n")

	// If using subscription, models will be selected automatically
	if provider.Auth.Method == config.AuthMethodSubscription {
		fmt.Printf("Using subscription authentication, models will be selected automatically based on your subscription.\n")
		fmt.Printf("You can specify custom models later using 'cflip config set-model <category> <model-id>'.\n")
		return nil
	}

	// For API key providers, let user choose models
	categories := []string{"haiku", "sonnet", "opus"}

	for _, category := range categories {
		// Get available models for this category from the provider
		var availableModels []config.ModelConfig
		for _, modelID := range provider.Models {
			if model, exists := cfg.Models[modelID]; exists && model.Category == category {
				availableModels = append(availableModels, model)
			}
		}

		if len(availableModels) == 0 {
			fmt.Printf("No %s models available for %s\n", category, provider.DisplayName)
			continue
		}

		// Show current active model
		currentModelID := cfg.Active.ModelMapping[category]
		currentModelName := ""
		if currentModelID != "" {
			if model, exists := cfg.Models[currentModelID]; exists {
				currentModelName = model.Name
			}
		}

		fmt.Printf("\n%s models:\n", category)
		for i, model := range availableModels {
			marker := ""
			if model.ID == currentModelID {
				marker = " [CURRENT]"
			}
			fmt.Printf("  %d) %s%s\n", i+1, model.Name, marker)
		}

		fmt.Printf("Select %s model (1-%d) [current: %s]: ",
			category, len(availableModels), currentModelName)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			// Keep current selection
			continue
		}

		if choice, err := strconv.Atoi(input); err == nil {
			if choice >= 1 && choice <= len(availableModels) {
				selectedModel := availableModels[choice-1]
				if err := tomlManager.SetActiveModel(category, selectedModel.ID); err != nil && verbose {
					fmt.Printf("Warning: Failed to set model: %v\n", err)
				} else {
					fmt.Printf("✓ Selected %s for %s\n", selectedModel.Name, category)
				}
			}
		}
	}

	return nil
}

func configureSettings(tomlManager *config.TOMLManagerV2, settings *config.SettingsConfig, verbose bool) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\nSettings Configuration\n")
	fmt.Printf("----------------------\n")

	// Backup directory
	fmt.Printf("Backup directory [%s]: ", settings.BackupDirectory)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		settings.BackupDirectory = input
	}

	// Max backups
	fmt.Printf("Maximum backups to keep [%d]: ", settings.MaxBackups)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		if num, err := strconv.Atoi(input); err == nil && num > 0 {
			settings.MaxBackups = num
		}
	}

	// Auto backup
	fmt.Printf("Enable automatic backups before switching? (Y/n): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" || input == "y" || input == "yes" {
		settings.AutoBackup = true
	} else {
		settings.AutoBackup = false
	}

	// Secure storage
	fmt.Printf("Enable secure storage for API keys? (Y/n): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" || input == "y" || input == "yes" {
		settings.SecureStorage = true
	} else {
		settings.SecureStorage = false
	}

	if err := tomlManager.UpdateSettings(*settings); err != nil && verbose {
		fmt.Printf("Warning: Failed to update settings: %v\n", err)
	} else {
		fmt.Printf("✓ Settings configured\n")
	}

	return nil
}

func promptTestConnection(provider *config.ProviderInfo) bool {
	if provider.Auth.Method == config.AuthMethodSubscription {
		return false // Skip test for subscription-based auth
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Test connection to verify configuration? (Y/n): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}

func testProviderConnection(provider *config.ProviderInfo) error {
	// TODO: Implement connection test
	// For now, just validate the API key format
	if !provider.HasAPIKey() {
		return fmt.Errorf("no API key configured")
	}

	switch provider.Name {
	case "anthropic":
		if !strings.HasPrefix(provider.GetAPIKey(), "sk-ant-") {
			return fmt.Errorf("invalid API key format")
		}
	case "glm":
		if !strings.HasPrefix(provider.GetAPIKey(), "zai-") {
			return fmt.Errorf("invalid API key format")
		}
	}

	return nil
}

func printNextSteps(provider *config.ProviderInfo) {
	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║                        Next Steps                            ║
╚══════════════════════════════════════════════════════════════╝

Your CFLIP configuration is complete!

Commands you can use:
`)

	if provider.Auth.Method == config.AuthMethodSubscription {
		fmt.Printf("  • claude /login      - Authenticate with your subscription\n")
		fmt.Printf("  • claude /whoami     - Verify your authentication\n")
	}

	fmt.Printf(`  • cflip switch       - Switch between providers
  • cflip config show  - Show current configuration
  • cflip config       - Manage configuration
  • cflip list         - List all available providers
  • cflip status       - Check current status

For more help, run: cflip --help

`)
}