package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all providers and show current selection",
	Long: `List all configured providers and indicate which one is currently active.
Shows provider names, plan types, and configuration status.`,
	Aliases: []string{"ls"},
	RunE:    runList,
}

func init() {
	listCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}

// NewListCmd exports the list command
func NewListCmd() *cobra.Command {
	return listCmd
}

func runList(cmd *cobra.Command, args []string) error {
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if jsonOutput {
		return outputProvidersJSON(cfg)
	}

	return outputProvidersText(cfg)
}

func outputProvidersText(cfg *config.Config) error {
	fmt.Println("Providers:")
	fmt.Println()

	// Always include anthropic as first option
	providerNames := []string{anthropicProvider}

	// Add configured external providers in sorted order
	var externalProviders []string
	for name := range cfg.Providers {
		if name != anthropicProvider {
			externalProviders = append(externalProviders, name)
		}
	}
	sort.Strings(externalProviders)
	providerNames = append(providerNames, externalProviders...)

	// Find current provider index
	var currentIndex = -1
	for i, name := range providerNames {
		if cfg.Provider == name {
			currentIndex = i + 1
			break
		}
	}

	// Display each provider
	for i, name := range providerNames {
		isCurrent := cfg.Provider == name
		displayName, statusText := getProviderDisplayInfo(name, cfg.Providers[name])

		// Format the output
		prefix := "  "
		if isCurrent {
			prefix = "→ "
		}

		fmt.Printf("%s%d) %s", prefix, i+1, displayName)
		if statusText != "" {
			fmt.Printf(" (%s)", statusText)
		}
		if isCurrent {
			fmt.Printf(" [CURRENT]")
		}
		fmt.Printf("\n")
	}

	fmt.Println()
	if currentIndex > 0 {
		fmt.Printf("Current provider: %d) %s\n", currentIndex, cfg.Provider)
	} else {
		fmt.Println("No provider selected")
	}

	return nil
}

func outputProvidersJSON(cfg *config.Config) error {
	// Always include anthropic as first option
	providerNames := []string{anthropicProvider}

	// Add configured external providers in sorted order
	var externalProviders []string
	for name := range cfg.Providers {
		if name != anthropicProvider {
			externalProviders = append(externalProviders, name)
		}
	}
	sort.Strings(externalProviders)
	providerNames = append(providerNames, externalProviders...)

	fmt.Println("{")
	fmt.Printf(`  "current": "%s",`+"\n", cfg.Provider)
	fmt.Println(`  "providers": [`)

	for i, name := range providerNames {
		provider := cfg.Providers[name]
		displayName, statusText := getProviderDisplayInfo(name, provider)

		fmt.Printf("    {")
		fmt.Printf(`"index": %d, `, i+1)
		fmt.Printf(`"name": "%s", `, name)
		fmt.Printf(`"displayName": "%s", `, displayName)
		fmt.Printf(`"status": "%s", `, statusText)
		fmt.Printf(`"isCurrent": %t`, cfg.Provider == name)

		fmt.Printf("}")

		if i < len(providerNames)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}

	fmt.Println("  ]")
	fmt.Println("}")

	return nil
}
