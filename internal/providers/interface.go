package providers

import "github.com/vanducng/cflip/internal/config"

// Provider defines the interface for Claude Code providers
type Provider interface {
	// Name returns the unique identifier for the provider
	Name() string

	// DisplayName returns a human-readable name for the provider
	DisplayName() string

	// Description returns a brief description of the provider
	Description() string

	// GetConfig returns the provider configuration
	GetConfig() *config.Provider

	// ValidateAPIKey validates if the provided API key is valid for this provider
	ValidateAPIKey(apiKey string) error

	// GetModels returns the available models for this provider
	GetModels() map[string]string

	// GetBaseURL returns the base URL for the provider's API
	GetBaseURL() string

	// RequiresSetup returns true if the provider requires additional setup
	RequiresSetup() bool

	// SetupInstructions returns setup instructions for the provider
	SetupInstructions() string
}

// Registry manages available providers
type Registry interface {
	// Register adds a provider to the registry
	Register(provider Provider) error

	// Get returns a provider by name
	Get(name string) (Provider, error)

	// List returns all registered providers
	List() []Provider

	// Exists checks if a provider is registered
	Exists(name string) bool
}