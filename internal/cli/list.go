package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
	"github.com/vanducng/cflip/internal/providers"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available providers",
	Long: `List all available Claude providers that you can switch to.
Shows each provider's name, description, and available models.`,
	RunE: runList,
}

func newListCmd() *cobra.Command {
	return listCmd
}

func runList(cmd *cobra.Command, args []string) error {
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Get the provider registry
	registry := providers.GetGlobalRegistry()

	// Get current provider
	configManager := config.NewManager()
	currentProvider, err := configManager.GetCurrentProvider()
	if err != nil && verbose {
		fmt.Printf("Warning: Could not determine current provider: %v\n", err)
	}

	// Create tabwriter for nice formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if _, err := fmt.Fprintln(w, "PROVIDER\tNAME\tDESCRIPTION\tMODELS"); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// List all providers
	for _, provider := range registry.List() {
		models := provider.GetModels()
		modelList := fmt.Sprintf("%s/%s/%s", models["haiku"], models["sonnet"], models["opus"])

		// Add indicator for current provider
		indicator := " "
		if currentProvider == provider.Name() {
			indicator = "*"
		}

		if _, err := fmt.Fprintf(w, "%s %s\t%s\t%s\t%s\n",
			indicator,
			provider.Name(),
			provider.DisplayName(),
			provider.Description(),
			modelList,
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
		fmt.Printf("\nTotal providers: %d\n", len(registry.List()))
	}

	return nil
}