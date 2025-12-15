package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available providers",
	Long: `List all available Claude providers that you can switch to.
Shows each provider's name, description, authentication method, and available models.`,
	RunE: runList,
}

func newListCmd() *cobra.Command {
	return listCmd
}

func runList(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Load configuration
	tomlManager := config.NewTOMLManagerV2()
	cfg, err := tomlManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get current provider
	currentProvider := cfg.Active.Provider

	// Create tabwriter for nice formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "PROVIDER\tNAME\tDESCRIPTION\tAUTH\tMODELS"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// List all providers
	for name, provider := range cfg.Providers {
		// Get model names for this provider
		var modelNames []string
		for _, modelID := range provider.Models {
			if model, exists := cfg.Models[modelID]; exists {
				modelNames = append(modelNames, model.Name)
			}
		}

		modelList := fmt.Sprintf("%d models", len(modelNames))
		if verbose && len(modelNames) > 0 {
			modelList = fmt.Sprintf("%d models (%s)", len(modelNames), modelNames[0])
			if len(modelNames) > 1 {
				modelList += fmt.Sprintf(" +%d", len(modelNames)-1)
			}
		}

		// Determine authentication method
		authMethod := "API Key"
		if !provider.IsAPIKeyRequired() {
			authMethod = "Subscription"
		}

		// Add indicator for current provider
		indicator := " "
		if currentProvider == name {
			indicator = "*"
		}

		// Check if API key is configured
		status := ""
		if provider.IsAPIKeyRequired() {
			if provider.HasAPIKey() {
				if verbose {
					status = " ✓"
				}
			} else {
				status = " ✗"
			}
		}

		if _, err := fmt.Fprintf(w, "%s %s\t%s\t%s\t%s\t%s%s\n",
			indicator,
			name,
			provider.DisplayName,
			provider.Description,
			authMethod,
			modelList,
			status,
		); err != nil {
			return fmt.Errorf("failed to write provider row: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	if currentProvider != "" {
		fmt.Printf("\n* Currently active provider\n")
	}

	if verbose {
		fmt.Printf("\nConfiguration file: %s\n", config.GetConfigPath())
		fmt.Printf("Total providers: %d\n", len(cfg.Providers))

		// Show models for each provider
		fmt.Printf("\nAvailable Models:\n")
		for _, provider := range cfg.Providers {
			fmt.Printf("\n%s (%s):\n", provider.DisplayName, provider.Name)
			for _, modelID := range provider.Models {
				if model, exists := cfg.Models[modelID]; exists {
					fmt.Printf("  • %s (%s) - %s\n", model.Name, model.Category, model.Description)
				}
			}
		}
	}

	return nil
}