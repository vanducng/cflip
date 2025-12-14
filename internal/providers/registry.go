package providers

import (
	"fmt"
	"sync"
)

// DefaultRegistry is the default implementation of the Registry interface
type DefaultRegistry struct {
	mu        sync.RWMutex
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() Registry {
	r := &DefaultRegistry{
		providers: make(map[string]Provider),
	}

	// Register default providers
	r.Register(NewAnthropicProvider())
	r.Register(NewGLMProvider())

	return r
}

// Register adds a provider to the registry
func (r *DefaultRegistry) Register(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if provider already exists
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider '%s' is already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Get returns a provider by name
func (r *DefaultRegistry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	return provider, nil
}

// List returns all registered providers
func (r *DefaultRegistry) List() []Provider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]Provider, 0, len(r.providers))
	for _, provider := range r.providers {
		list = append(list, provider)
	}

	return list
}

// Exists checks if a provider is registered
func (r *DefaultRegistry) Exists(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.providers[name]
	return exists
}

// GetNames returns a list of all registered provider names
func (r *DefaultRegistry) GetNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}

	return names
}

// Global registry instance
var globalRegistry Registry

// Initialize the global registry
func init() {
	globalRegistry = NewRegistry()
}

// GetGlobalRegistry returns the global provider registry
func GetGlobalRegistry() Registry {
	return globalRegistry
}

// GetProvider is a convenience function to get a provider from the global registry
func GetProvider(name string) (Provider, error) {
	return globalRegistry.Get(name)
}

// ListProviders is a convenience function to list all providers from the global registry
func ListProviders() []Provider {
	return globalRegistry.List()
}

// ProviderExists is a convenience function to check if a provider exists in the global registry
func ProviderExists(name string) bool {
	return globalRegistry.Exists(name)
}