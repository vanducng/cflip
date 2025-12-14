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

	registry := providers.GetGlobalRegistry()

	providerName, err := getProviderName(args, registry)
	if err != nil {
		return err
	}

	provider, err := registry.Get(providerName)
	if err != nil {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	configManager := config.NewManager()

	if shouldSkipSwitch, err := checkCurrentProvider(configManager, provider, verbose, quiet); err != nil {
		return err
	} else if shouldSkipSwitch {
		return nil
	}

	apiKey, err := getAPIKey(provider)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}

	if err := provider.ValidateAPIKey(apiKey); err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}

	if err := createBackup(configManager, verbose, quiet); err != nil {
		if verbose {
			fmt.Printf("\nWarning: Failed to create backup: %v\n", err)
		}
	}

	if err := switchProvider(configManager, provider, apiKey); err != nil {
		return err
	}

	if !quiet {
		displaySwitchSuccess(provider, verbose)
	}

	return nil
}

func getProviderName(args []string, registry providers.Registry) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}

	return promptProviderSelection(registry)
}

func promptProviderSelection(registry providers.Registry) (string, error) {
	providerList := registry.List()
	if len(providerList) == 0 {
		return "", fmt.Errorf("no providers available")
	}

	fmt.Println("Available providers:")
	for i, provider := range providerList {
		fmt.Printf("  %d) %s - %s\n", i+1, provider.DisplayName(), provider.Description())
	}

	fmt.Print("\nSelect provider (1-?): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)

	// Try to convert number to provider name
	for i, provider := range providerList {
		if fmt.Sprintf("%d", i+1) == input {
			return provider.Name(), nil
		}
	}

	// Check if input matches provider name directly
	for _, provider := range providerList {
		if strings.EqualFold(provider.Name(), input) {
			return provider.Name(), nil
		}
	}

	return "", fmt.Errorf("invalid selection")
}

func checkCurrentProvider(configManager *config.Manager, provider providers.Provider, verbose, quiet bool) (bool, error) {
	currentProvider, err := configManager.GetCurrentProvider()
	if err != nil {
		if verbose {
			fmt.Printf("Warning: Could not determine current provider: %v\n", err)
		}
		return false, nil
	}

	if currentProvider == provider.Name() {
		if !quiet {
			fmt.Printf("Already using %s provider\n", provider.DisplayName())
		}
		return true, nil
	}

	return false, nil
}

func createBackup(configManager *config.Manager, verbose, quiet bool) error {
	if !quiet {
		fmt.Print("Creating backup... ")
	}

	backup, err := configManager.CreateBackup()
	if err != nil {
		return err
	}

	if !quiet {
		fmt.Printf("done (ID: %s)\n", backup.ID)
	}

	return nil
}

func switchProvider(configManager *config.Manager, provider providers.Provider, apiKey string) error {
	providerConfig := provider.GetConfig()
	settings := providerConfig.Merge(apiKey)

	return configManager.SaveSettings(settings)
}

func displaySwitchSuccess(provider providers.Provider, verbose bool) {
	fmt.Printf("Successfully switched to %s\n", provider.DisplayName())
	fmt.Printf("\nModels:\n")
	for modelType, modelName := range provider.GetModels() {
		fmt.Printf("  %s: %s\n", modelType, modelName)
	}

	if provider.RequiresSetup() {
		fmt.Printf("\nSetup Instructions:\n%s\n", provider.SetupInstructions())
	}
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