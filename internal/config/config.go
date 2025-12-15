package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/BurntSushi/toml"
)

// Config represents the configuration structure
type Config struct {
	Provider  string                    `toml:"provider"` // "anthropic" or external name
	Providers map[string]ProviderConfig `toml:"providers"`
}

// ProviderConfig represents a provider configuration
type ProviderConfig struct {
	// For external providers only
	Token   string `toml:"token,omitempty"`
	BaseURL string `toml:"base_url,omitempty"`

	// Optional model mapping (external -> anthropic)
	ModelMap map[string]string `toml:"model_map,omitempty"`
}

// NewConfig creates a new default configuration
func NewConfig() *Config {
	return &Config{
		Provider: "anthropic",
		Providers: map[string]ProviderConfig{
			"anthropic": {},
		},
	}
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cflip", "config.toml")
}

// LoadConfig loads the configuration from file
func LoadConfig() (*Config, error) {
	configPath := GetConfigPath()

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return NewConfig(), nil
	}

	// Load and parse TOML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := NewConfig()
	if err := toml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return config, nil
}

// SaveConfig saves the configuration to file
func SaveConfig(config *Config) error {
	configPath := GetConfigPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to TOML
	var buf strings.Builder
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	data := []byte(buf.String())

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetActiveProvider returns the active provider configuration
func (c *Config) GetActiveProvider() (*ProviderConfig, error) {
	provider, exists := c.Providers[c.Provider]
	if !exists {
		return nil, fmt.Errorf("active provider '%s' not found", c.Provider)
	}
	return &provider, nil
}

// SetActiveProvider sets the active provider
func (c *Config) SetActiveProvider(providerName string) error {
	if _, exists := c.Providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}
	c.Provider = providerName
	return nil
}

// SetProviderConfig adds or updates a provider configuration
func (c *Config) SetProviderConfig(name string, config ProviderConfig) {
	if c.Providers == nil {
		c.Providers = make(map[string]ProviderConfig)
	}
	c.Providers[name] = config
}

// IsExternal returns true if the provider is an external provider (not Anthropic)
func (c *Config) IsExternal(providerName string) bool {
	return providerName != "anthropic"
}
