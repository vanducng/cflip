package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"golang.org/x/term"
)

const (
	anthropicProvider = "anthropic"
	yesResponse       = "yes"
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
	if err := generateClaudeSettings(cfg); err != nil {
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
	// Always include anthropic as an option
	providerNames := []string{anthropicProvider}

	// Add configured external providers
	for name := range cfg.Providers {
		if name != anthropicProvider {
			providerNames = append(providerNames, name)
		}
	}

	fmt.Println("Available providers:")
	for i, name := range providerNames {
		current := ""
		if cfg.Provider == name {
			current = " [CURRENT]"
		}

		displayName := name
		if name == anthropicProvider {
			displayName = "Anthropic (Official)"
		}

		provider := cfg.Providers[name]
		fmt.Printf("  %d) %s%s", i+1, displayName, current)

		if name != anthropicProvider {
			if provider.Token != "" {
				fmt.Printf(" (configured)")
			} else {
				fmt.Printf(" (needs configuration)")
			}
		} else {
			if provider.Token != "" {
				fmt.Printf(" (API key configured)")
			} else {
				fmt.Printf(" (no API key)")
			}
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

func configureExternalProvider(cfg *config.Config, providerName string, verbose, quiet bool) error {
	if !quiet {
		fmt.Printf("\nConfiguring %s provider\n", providerName)
	}

	provider := cfg.Providers[providerName]

	// Show current configuration status
	if !quiet {
		if provider.Token != "" {
			fmt.Printf("Using existing API token\n")
		}
	}

	// Prompt for token if not configured
	if provider.Token == "" {
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
	}

	// Show base URL status
	if !quiet && provider.BaseURL != "" {
		fmt.Printf("Using existing base URL\n")
	}

	// Prompt for base URL if not configured
	if provider.BaseURL == "" {
		fmt.Printf("Enter %s base URL: ", providerName)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			return fmt.Errorf("base URL cannot be empty")
		}
		provider.BaseURL = input
	}

	// Optionally configure model mappings
	fmt.Printf("\nConfigure model mappings? (Y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" || input == "y" || input == yesResponse {
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
	}

	cfg.SetProviderConfig(providerName, provider)
	return nil
}

func configureAnthropicProvider(cfg *config.Config, verbose, quiet bool) error {
	provider := cfg.Providers[anthropicProvider]

	// Optionally configure API key for Anthropic
	if !quiet && provider.Token == "" {
		fmt.Printf("\nConfigure API key for Anthropic? (optional, Y/n): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "" || input == "y" || input == yesResponse {
			fmt.Printf("Enter Anthropic API key (optional): ")
			bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return fmt.Errorf("failed to read API key: %w", err)
			}
			fmt.Println() // New line after password input

			token := strings.TrimSpace(string(bytePassword))
			if token != "" {
				provider.Token = token
				cfg.SetProviderConfig("anthropic", provider)
			}
		}
	}

	return nil
}

func generateClaudeSettings(cfg *config.Config) error {
	// Claude settings path
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Read existing settings
	var settings map[string]interface{}
	if _, err := os.ReadFile(settingsPath); err == nil {
		// Parse existing JSON (simple parsing)
		settings = make(map[string]interface{})
		// For simplicity, we'll recreate the file structure

		// Ensure env map exists
		if settings["env"] == nil {
			settings["env"] = make(map[string]string)
		}
	} else {
		// Create new settings structure
		settings = map[string]interface{}{
			"env": make(map[string]string),
		}
	}

	// Get env map
	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		env = make(map[string]interface{})
		settings["env"] = env
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
		delete(env, key)
	}

	// Configure based on provider
	if cfg.Provider == anthropicProvider {
		provider := cfg.Providers[anthropicProvider]

		// Only set API key if provided
		if provider.Token != "" {
			env["ANTHROPIC_AUTH_TOKEN"] = provider.Token
		}

		// Do NOT set ANTHROPIC_BASE_URL - use Claude Code default
		// Do NOT set model mappings - use defaults
	} else {
		// External provider
		provider := cfg.Providers[cfg.Provider]

		// Set required fields
		env["ANTHROPIC_AUTH_TOKEN"] = provider.Token
		env["ANTHROPIC_BASE_URL"] = provider.BaseURL

		// Set model mappings if available
		if len(provider.ModelMap) > 0 {
			if haikuModel, exists := provider.ModelMap["haiku"]; exists {
				env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = haikuModel
			}
			if sonnetModel, exists := provider.ModelMap["sonnet"]; exists {
				env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = sonnetModel
			}
			if opusModel, exists := provider.ModelMap["opus"]; exists {
				env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = opusModel
			}
		}
	}

	// For simplicity, write basic JSON structure
	output := fmt.Sprintf(`{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "env": {
%s
  }
}`, formatEnvMap(env))

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0750); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Write settings
	if err := os.WriteFile(settingsPath, []byte(output), 0600); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	return nil
}

func formatEnvMap(env map[string]interface{}) string {
	var lines []string
	for k, v := range env {
		lines = append(lines, fmt.Sprintf(`    %q: %q`, k, v))
	}
	return strings.Join(lines, ",\n")
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