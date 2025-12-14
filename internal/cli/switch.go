package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"github.com/vanducng/cflip/internal/providers"
	"golang.org/x/term"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch [provider]",
	Short: "Switch to a different Claude provider",
	Long: `Switch the active Claude provider. This will update your ~/.claude/settings.json
file to use the specified provider's API endpoint and models.

Available providers:
  anthropic - Official Anthropic Claude API
  glm       - GLM models by z.ai

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

	// Get the provider registry
	registry := providers.GetGlobalRegistry()

	var providerName string

	if len(args) > 0 {
		providerName = args[0]
	} else {
		// Interactive mode
		providerList := registry.List()
		if len(providerList) == 0 {
			return fmt.Errorf("no providers available")
		}

		fmt.Println("Available providers:")
		for i, provider := range providerList {
			fmt.Printf("  %d) %s - %s\n", i+1, provider.DisplayName(), provider.Description())
		}

		fmt.Print("\nSelect provider (1-?): ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		// Convert number to provider name
		for i, provider := range providerList {
			if fmt.Sprintf("%d", i+1) == input {
				providerName = provider.Name()
				break
			}
		}

		if providerName == "" {
			// Check if input matches provider name directly
			for _, provider := range providerList {
				if strings.EqualFold(provider.Name(), input) {
					providerName = provider.Name()
					break
				}
			}
		}

		if providerName == "" {
			return fmt.Errorf("invalid selection")
		}
	}

	// Get the provider
	provider, err := registry.Get(providerName)
	if err != nil {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// Create configuration manager
	configManager := config.NewManager()

	// Check current provider
	currentProvider, err := configManager.GetCurrentProvider()
	if err != nil {
		if verbose {
			fmt.Printf("Warning: Could not determine current provider: %v\n", err)
		}
	}

	if currentProvider == provider.Name() {
		if !quiet {
			fmt.Printf("Already using %s provider\n", provider.DisplayName())
		}
		return nil
	}

	// Get API key
	apiKey, err := getAPIKey(provider)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	// Validate API key
	if err := provider.ValidateAPIKey(apiKey); err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	// Create backup before switching
	if !quiet {
		fmt.Print("Creating backup... ")
	}
	backup, err := configManager.CreateBackup()
	if err != nil {
		if verbose {
			fmt.Printf("\nWarning: Failed to create backup: %v\n", err)
		}
	} else if !quiet {
		fmt.Printf("done (ID: %s)\n", backup.ID)
	}

	// Merge provider configuration
	providerConfig := provider.GetConfig()
	settings := providerConfig.Merge(apiKey)

	// Save new settings
	if err := configManager.SaveSettings(settings); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	if !quiet {
		fmt.Printf("Successfully switched to %s\n", provider.DisplayName())
		fmt.Printf("\nModels:\n")
		for modelType, modelName := range provider.GetModels() {
			fmt.Printf("  %s: %s\n", modelType, modelName)
		}

		// Show setup instructions if needed
		if provider.RequiresSetup() {
			fmt.Printf("\nSetup Instructions:\n%s\n", provider.SetupInstructions())
		}
	}

	return nil
}

func getAPIKey(provider providers.Provider) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	// Check if API key is already stored
	configManager := config.NewManager()
	settings, err := configManager.LoadSettings()
	if err == nil {
		if existingKey, exists := settings.Env["ANTHROPIC_AUTH_TOKEN"]; exists {
			// Validate existing key
			if err := provider.ValidateAPIKey(existingKey); err == nil {
				fmt.Printf("Using existing API key for %s\n", provider.DisplayName())
				fmt.Print("Press Enter to use existing key, or type a new one: ")
				input, _ := reader.ReadString('\n')
				input = strings.TrimSpace(input)
				if input == "" {
					return existingKey, nil
				}
			}
		}
	}

	// Prompt for new API key
	fmt.Printf("Enter %s API key: ", provider.DisplayName())
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read API key: %w", err)
	}
	fmt.Println() // New line after password input

	apiKey := strings.TrimSpace(string(bytePassword))
	if apiKey == "" {
		return "", fmt.Errorf("API key cannot be empty")
	}

	return apiKey, nil
}