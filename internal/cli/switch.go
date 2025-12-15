package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"golang.org/x/term"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [provider]",
	Short: "Switch to a different Claude provider",
	Long: `Switch the active Claude provider. This will update your ~/.cflip/config.toml
file and generate the appropriate Claude settings for the specified provider.

Available providers:
  anthropic   - Official Anthropic Claude API (requires API key)
  claude-code - Claude Code with subscription (uses 'claude /login')
  glm         - GLM models by z.ai (requires API key)

The configuration supports:
- Multiple providers with different authentication methods
- Centralized model management
- Secure API key storage
- Custom model mappings per category

If no provider is specified, you will be prompted to choose from the available options.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func newSwitchCmd() *cobra.Command {
	return switchCmd
}

func runSwitch(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	quiet, _ := cmd.Flags().GetBool("quiet")

	tomlManager := config.NewTOMLManagerV2()
	cfg, err := tomlManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get provider name
	providerName, err := getProviderName(args, cfg, verbose)
	if err != nil {
		return err
	}

	// Check if already using this provider
	if cfg.Active.Provider == providerName {
		if !quiet {
			fmt.Printf("Already using %s provider\n", providerName)
		}
		return nil
	}

	// Get provider configuration
	provider, err := cfg.GetProviderByName(providerName)
	if err != nil {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// Handle authentication based on provider type
	if provider.IsAPIKeyRequired() {
		if err := handleAPIKeyAuthentication(provider, verbose, quiet); err != nil {
			return err
		}
	} else {
		if err := handleSubscriptionAuthentication(provider, verbose, quiet); err != nil {
			return err
		}
	}

	// Create backup of current settings
	if err := createClaudeBackup(cfg, verbose, quiet); err != nil && verbose {
		fmt.Printf("\nWarning: Failed to create backup: %v\n", err)
	}

	// Switch provider
	if err := switchProvider(tomlManager, providerName); err != nil {
		return err
	}

	// Generate Claude settings file
	if err := generateClaudeSettings(tomlManager, providerName); err != nil {
		return fmt.Errorf("failed to generate Claude settings: %w", err)
	}

	if !quiet {
		displaySwitchSuccess(provider, verbose)
	}

	return nil
}

func getProviderName(args []string, cfg *config.CFLIPConfig, verbose bool) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	return promptProviderSelection(cfg)
}

func promptProviderSelection(cfg *config.CFLIPConfig) (string, error) {
	providers := cfg.Providers
	providerNames := cfg.ListProviders()

	fmt.Println("Available providers:")
	for i, name := range providerNames {
		provider := providers[name]
		current := ""
		if cfg.Active.Provider == name {
			current = " [CURRENT]"
		}

		fmt.Printf("  %d) %s - %s%s", i+1, provider.DisplayName, provider.Description, current)

		if provider.IsAPIKeyRequired() {
			if provider.HasAPIKey() {
				fmt.Printf(" (API key configured)")
			} else {
				fmt.Printf(" (needs API key)")
			}
		} else {
			fmt.Printf(" (subscription)")
		}
		fmt.Printf("\n")
	}

	fmt.Print("\nSelect provider (1-?): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)

	// Try to convert number to provider name
	for i, name := range providerNames {
		if fmt.Sprintf("%d", i+1) == input {
			return name, nil
		}
	}

	// Check if input matches provider name directly
	for _, name := range providerNames {
		if strings.EqualFold(name, input) || strings.EqualFold(providers[name].DisplayName, input) {
			return name, nil
		}
	}

	return "", fmt.Errorf("invalid selection")
}

func handleAPIKeyAuthentication(provider *config.ProviderInfo, verbose, quiet bool) error {
	if !quiet {
		fmt.Printf("\nConfiguring %s (API Key Authentication)\n", provider.DisplayName)
	}

	// Check if we already have a valid API key
	if provider.HasAPIKey() {
		if !quiet {
			fmt.Printf("Using existing API key\n")
		}
		return nil
	}

	// Check if we can import from legacy settings
	if existingKey := getLegacyAPIKey(); existingKey != "" {
		fmt.Printf("Found existing API key in ~/.claude/settings.json\n")
		reader := bufio.NewReader(os.Stdin)
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

	// Prompt for new API key
	fmt.Printf("Enter %s API key: ", provider.DisplayName)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}
	fmt.Println() // New line after password input

	apiKey := strings.TrimSpace(string(bytePassword))
	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Basic validation
	if err := validateAPIKeyFormat(provider.Name, apiKey); err != nil && verbose {
		fmt.Printf("Warning: %v\n", err)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Continue anyway? (y/N): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			return fmt.Errorf("cancelled by user")
		}
	}

	provider.SetAPIKey(apiKey)
	if !quiet {
		fmt.Printf("✓ API key configured\n")
	}

	return nil
}

func handleSubscriptionAuthentication(provider *config.ProviderInfo, verbose, quiet bool) error {
	if !quiet {
		fmt.Printf("\nConfiguring %s (Subscription Authentication)\n", provider.DisplayName)
		fmt.Printf("This provider uses your Claude subscription for authentication.\n\n")

		if provider.Auth.SetupInstructions != "" {
			fmt.Printf("Setup Instructions:\n%s\n\n", provider.Auth.SetupInstructions)
		}

		fmt.Printf("Make sure you:\n")
		fmt.Printf("  • Have an active Claude Pro or Max subscription\n")
		fmt.Printf("  • Run 'claude /login' to authenticate\n")
		fmt.Printf("  • Run 'claude /whoami' to verify your login\n\n")
	}

	return nil
}

func createClaudeBackup(cfg *config.CFLIPConfig, verbose, quiet bool) error {
	// Use the legacy backup system for Claude settings
	legacyManager := config.NewManager()

	if !quiet {
		fmt.Print("Creating backup of current Claude settings... ")
	}

	backup, err := legacyManager.CreateBackup()
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Printf("done (ID: %s)\n", backup.ID)
	}

	return nil
}

func switchProvider(tomlManager *config.TOMLManagerV2, providerName string) error {
	return tomlManager.SetActiveProvider(providerName)
}

func generateClaudeSettings(tomlManager *config.TOMLManagerV2, providerName string) error {
	cfg, err := tomlManager.LoadConfig()
	if err != nil {
		return err
	}

	provider, err := cfg.GetProviderByName(providerName)
	if err != nil {
		return err
	}

	// Create legacy Claude settings
	legacyManager := config.NewManager()

	// Map to legacy structure
	legacySettings := &config.ClaudeSettings{
		Env: make(map[string]string),
	}

	// Set authentication based on provider type
	if provider.IsAPIKeyRequired() {
		legacySettings.Env["ANTHROPIC_AUTH_TOKEN"] = provider.GetAPIKey()
		legacySettings.Env["ANTHROPIC_BASE_URL"] = provider.Auth.BaseURL
	} else {
		// For subscription-based auth, we don't set these
		// Claude Code CLI will handle authentication
		legacySettings.Env["ANTHROPIC_BASE_URL"] = ""
	}

	// Set model mappings
	for category, modelID := range cfg.Active.ModelMapping {
		if model, err := cfg.GetModelConfig(modelID); err == nil {
			envKey := fmt.Sprintf("ANTHROPIC_DEFAULT_%s_MODEL", strings.ToUpper(category))
			legacySettings.Env[envKey] = model.ID
		}
	}

	// Set additional environment variables
	if provider.EnvVars != nil {
		for key, value := range provider.EnvVars {
			legacySettings.Env[key] = value
		}
	}

	// Set default timeout
	if _, exists := legacySettings.Env["API_TIMEOUT_MS"]; !exists {
		timeout := provider.Auth.TimeoutSeconds
		if timeout == 0 {
			timeout = 300
		}
		legacySettings.Env["API_TIMEOUT_MS"] = fmt.Sprintf("%d", timeout*1000)
	}

	return legacyManager.SaveSettings(legacySettings)
}

func displaySwitchSuccess(provider *config.ProviderInfo, verbose bool) {
	fmt.Printf("\n✓ Successfully switched to %s\n", provider.DisplayName)

	// Show active models
	cfg, _ := config.NewTOMLManagerV2().LoadConfig()
	fmt.Printf("\nActive Models:\n")

	for category, modelID := range cfg.Active.ModelMapping {
		if model, err := cfg.GetModelConfig(modelID); err == nil {
			fmt.Printf("  %s: %s", category, model.Name)
			if model.MaxTokens > 0 {
				fmt.Printf(" (max tokens: %d)", model.MaxTokens)
			}
			fmt.Printf("\n")
		}
	}

	// Show authentication method
	if provider.IsAPIKeyRequired() {
		fmt.Printf("\nAuthentication: API Key\n")
	} else {
		fmt.Printf("\nAuthentication: Subscription (Claude Code CLI)\n")
		fmt.Printf("Remember to run 'claude /login' if not already authenticated\n")
	}

	if verbose {
		fmt.Printf("\nConfiguration saved to: %s\n", config.GetConfigPath())
		fmt.Printf("Claude settings updated at: %s\n", config.GetLegacySettingsPath())
	}
}

