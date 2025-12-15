package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"golang.org/x/term"
)

const (
	anthropicProvider  = "anthropic"
	claudeCodeProvider = "claude-code"
	anthropicName      = "Anthropic"
	glmProvider        = "glm"
	statusOAuth        = "OAuth"
	statusAPI          = "API"
	currentMarker      = " [CURRENT]"
	yesResponse        = "yes"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [provider]",
	Short: "Switch to a different Claude provider",
	Long: `Switch the active Claude provider. This will update your ~/.cflip/config.toml
file and generate the appropriate Claude settings for the specified provider.

Available providers:
  anthropic - Official Anthropic Claude API (optional API key, uses default endpoint)
  glm       - GLM models by z.ai (requires API key and base URL)
  custom    - Any custom provider (requires API key and base URL)

For external providers (glm, custom), you can optionally configure model mappings
to map their models to Anthropic's model categories (haiku, sonnet, opus).

If no provider is specified, you will be prompted to choose from the available options.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func newSwitchCmd() *cobra.Command {
	return switchCmd
}

// NewSwitchCmd exports the switch command
func NewSwitchCmd() *cobra.Command {
	return switchCmd
}

func runSwitch(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")
	quiet, _ := cmd.Flags().GetBool("quiet")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get provider name
	providerName, err := getProviderName(args, cfg, verbose)
	if err != nil {
		return err
	}

	// If no provider specified and in a terminal, use interactive mode
	if providerName == "" && len(args) == 0 && isTerminal() {
		if provider, err := RunInteractiveSelection(cfg); err == nil && provider != "" {
			providerName = provider
		}
		// Fall through to text selection if interactive fails
	}

	// Check if already using this provider
	if cfg.Provider == providerName {
		if !quiet {
			fmt.Printf("Already using %s provider\n", providerName)
		}
		return nil
	}

	// Configure provider if needed
	if providerName != anthropicProvider {
		if err := configureExternalProvider(cfg, providerName, verbose, quiet); err != nil {
			return err
		}
	} else {
		if err := configureAnthropicProvider(cfg, verbose, quiet); err != nil {
			return err
		}
	}

	// Switch provider
	cfg.Provider = providerName

	// Save configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Generate Claude settings file
	if err := generateClaudeSettings(cfg, quiet); err != nil {
		return fmt.Errorf("failed to generate Claude settings: %w", err)
	}

	if !quiet {
		displaySwitchSuccess(cfg, providerName, verbose)
	}

	return nil
}

func getProviderName(args []string, cfg *config.Config, verbose bool) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	return promptProviderSelection(cfg)
}

func promptProviderSelection(cfg *config.Config) (string, error) {
	// Try interactive selection first
	if provider, err := RunInteractiveSelection(cfg); err == nil && provider != "" {
		return provider, nil
	}

	// Fall back to simple text selection if interactive fails
	return promptProviderSelectionText(cfg)
}

// promptProviderSelectionText provides the original text-based selection
func promptProviderSelectionText(cfg *config.Config) (string, error) {
	// Always include anthropic as first option
	providerNames := []string{anthropicProvider}

	// Add configured external providers in sorted order
	var externalProviders []string
	for name := range cfg.Providers {
		if name != anthropicProvider {
			externalProviders = append(externalProviders, name)
		}
	}
	// Sort external providers for consistent display
	sort.Strings(externalProviders)
	providerNames = append(providerNames, externalProviders...)

	fmt.Println("Available providers:")
	for i, name := range providerNames {
		current := ""
		if cfg.Provider == name {
			current = " [CURRENT]"
		}

		displayName, statusText := getProviderDisplayInfo(name, cfg.Providers[name])

		fmt.Printf("  %d) %s%s", i+1, displayName, current)
		if statusText != "" {
			fmt.Printf(" (%s)", statusText)
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
		if strings.EqualFold(name, input) {
			return name, nil
		}
	}

	// If it's a new provider name, add it
	if input != "" {
		fmt.Printf("Creating new provider '%s'\n", input)
		return input, nil
	}

	return "", fmt.Errorf("invalid selection")
}

// getProviderDisplayInfo returns the display name and status text for a provider
func getProviderDisplayInfo(providerName string, provider config.ProviderConfig) (displayName, statusText string) {
	if providerName == anthropicProvider {
		displayName = anthropicName
		statusText = statusOAuth
		return displayName, statusText
	}

	// External providers
	switch providerName {
	case claudeCodeProvider:
		displayName = anthropicName
	case glmProvider:
		displayName = "GLM"
	default:
		displayName = providerName
	}

	statusText = statusAPI

	return displayName, statusText
}

func configureExternalProvider(cfg *config.Config, providerName string, verbose, quiet bool) error {
	if !quiet {
		fmt.Printf("\nConfiguring %s provider\n", providerName)
	}

	provider := cfg.Providers[providerName]

	// Show current configuration status
	if !quiet {
		showProviderStatus(provider)
	}

	// Configure token if needed
	if err := configureToken(&provider, providerName); err != nil {
		return err
	}

	// Configure base URL if needed
	if err := configureBaseURL(&provider, providerName); err != nil {
		return err
	}

	// Configure model mappings if requested
	if err := configureModelMappings(&provider); err != nil {
		return err
	}

	cfg.SetProviderConfig(providerName, provider)
	return nil
}

// showProviderStatus displays the current provider configuration status
func showProviderStatus(provider config.ProviderConfig) {
	if provider.Token != "" {
		fmt.Printf("Using existing API token\n")
	}
	if provider.BaseURL != "" {
		fmt.Printf("Using existing base URL\n")
	}
}

// configureToken prompts for and configures the API token
func configureToken(provider *config.ProviderConfig, providerName string) error {
	if provider.Token != "" {
		return nil // Already configured
	}

	fmt.Printf("Enter %s API token: ", providerName)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to read API token: %w", err)
	}
	fmt.Println() // New line after password input

	token := strings.TrimSpace(string(bytePassword))
	if token == "" {
		return fmt.Errorf("API token cannot be empty")
	}
	provider.Token = token
	return nil
}

// configureBaseURL prompts for and configures the base URL
func configureBaseURL(provider *config.ProviderConfig, providerName string) error {
	if provider.BaseURL != "" {
		return nil // Already configured
	}

	fmt.Printf("Enter %s base URL: ", providerName)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return fmt.Errorf("base URL cannot be empty")
	}
	provider.BaseURL = input
	return nil
}

// configureModelMappings prompts for and configures model mappings
func configureModelMappings(provider *config.ProviderConfig) error {
	fmt.Printf("\nConfigure model mappings? (Y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "" && input != "y" && input != yesResponse {
		return nil // User declined
	}

	if provider.ModelMap == nil {
		provider.ModelMap = make(map[string]string)
	}

	// Prompt for each category
	categories := []string{"haiku", "sonnet", "opus"}
	for _, category := range categories {
		fmt.Printf("Enter model for %s category (optional): ", category)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input != "" {
			provider.ModelMap[category] = input
		}
	}
	return nil
}

// detectCurrentProvider determines the current provider from settings
func detectCurrentProvider(settings *ClaudeSettings) string {
	if settings.Env == nil {
		return "unknown"
	}

	if baseURL, exists := settings.Env["ANTHROPIC_BASE_URL"]; exists && baseURL != nil && baseURL != "" {
		// External provider has base URL, try to identify which one
		baseURLStr := fmt.Sprintf("%v", baseURL)
		if strings.Contains(baseURLStr, "z.ai") || strings.Contains(baseURLStr, "glm") {
			return "glm"
		} else if strings.Contains(baseURLStr, "openai") || strings.Contains(baseURLStr, "azure") {
			return "openai"
		} else {
			// Generic external provider
			return "external"
		}
	}

	// No base URL means Anthropic
	return "anthropic"
}

func configureAnthropicProvider(cfg *config.Config, verbose, quiet bool) error {
	// No configuration needed for Anthropic subscription plan
	// Users can optionally configure an API key later if needed

	if !quiet && verbose {
		fmt.Println("\nNote: Using Anthropic subscription plan")
		fmt.Println("No API key required - will use your Claude Code subscription")
	}

	return nil
}

func generateClaudeSettings(cfg *config.Config, quiet bool) error {
	// Claude settings path
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Load current settings with all attributes
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load current settings: %w", err)
	}

	// Create snapshot before switching (always, even if user edited manually)
	cflipDir := filepath.Dir(settingsPath)
	snapshotsDir := filepath.Join(cflipDir, "snapshots")

	// Determine the current provider from existing settings
	currentProvider := detectCurrentProvider(settings)

	// Create snapshot with current provider name
	if err := CreateSnapshot(settingsPath, snapshotsDir, currentProvider); err != nil {
		// Don't fail if snapshot fails, just log it
		if !quiet {
			fmt.Printf("Warning: Failed to create snapshot: %v\n", err)
		}
	}

	// Clean up old snapshots (keep last 5)
	if err := CleanupOldSnapshots(snapshotsDir, 5); err != nil {
		fmt.Printf("Warning: Failed to cleanup old snapshots: %v\n", err)
	}

	// Clear existing Claude-related env vars
	keysToDelete := []string{
		"ANTHROPIC_AUTH_TOKEN",
		"ANTHROPIC_BASE_URL",
		"ANTHROPIC_DEFAULT_HAIKU_MODEL",
		"ANTHROPIC_DEFAULT_SONNET_MODEL",
		"ANTHROPIC_DEFAULT_OPUS_MODEL",
	}
	for _, key := range keysToDelete {
		delete(settings.Env, key)
	}

	// Configure based on provider
	if cfg.Provider == anthropicProvider {
		provider := cfg.Providers[anthropicProvider]

		// Only set API key if provided
		if provider.Token != "" {
			settings.Env["ANTHROPIC_AUTH_TOKEN"] = provider.Token
		}

		// Do NOT set ANTHROPIC_BASE_URL - use Claude Code default
		// Do NOT set model mappings - use defaults
	} else {
		// External provider
		provider := cfg.Providers[cfg.Provider]

		// Set required fields
		settings.Env["ANTHROPIC_AUTH_TOKEN"] = provider.Token
		settings.Env["ANTHROPIC_BASE_URL"] = provider.BaseURL

		// Set model mappings if available
		if len(provider.ModelMap) > 0 {
			if haikuModel, exists := provider.ModelMap["haiku"]; exists {
				settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = haikuModel
			}
			if sonnetModel, exists := provider.ModelMap["sonnet"]; exists {
				settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = sonnetModel
			}
			if opusModel, exists := provider.ModelMap["opus"]; exists {
				settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = opusModel
			}
		}
	}

	// Save settings preserving all other fields
	return SaveSettings(settingsPath, settings)
}

func displaySwitchSuccess(cfg *config.Config, providerName string, verbose bool) {
	fmt.Printf("\n✓ Successfully switched to %s\n", providerName)

	if providerName == anthropicProvider {
		fmt.Printf("\nConfiguration: Using Anthropic with default endpoint\n")
		if cfg.Providers[anthropicProvider].Token != "" {
			fmt.Printf("Authentication: API Key configured\n")
		} else {
			fmt.Printf("Authentication: No API key (will use Claude Code subscription)\n")
		}
	} else {
		provider := cfg.Providers[providerName]
		fmt.Printf("\nConfiguration:\n")
		fmt.Printf("  Base URL: %s\n", provider.BaseURL)
		if len(provider.ModelMap) > 0 {
			fmt.Printf("  Model Mappings:\n")
			for category, model := range provider.ModelMap {
				fmt.Printf("    %s: %s\n", category, model)
			}
		}
		fmt.Printf("\nAuthentication: API Key\n")
	}

	if verbose {
		fmt.Printf("\nConfiguration saved to: %s\n", config.GetConfigPath())
		homeDir, _ := os.UserHomeDir()
		fmt.Printf("Claude settings updated at: %s\n", filepath.Join(homeDir, ".claude", "settings.json"))
	}
}
