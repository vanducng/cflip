package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"golang.org/x/crypto/sha3"
)

// TOMLManagerV2 handles CFLIP configuration file operations with the new structure
type TOMLManagerV2 struct {
	configPath string
}

// NewTOMLManagerV2 creates a new TOML configuration manager
func NewTOMLManagerV2() *TOMLManagerV2 {
	return &TOMLManagerV2{
		configPath: GetConfigPath(),
	}
}

// LoadConfig loads the CFLIP configuration from file
func (m *TOMLManagerV2) LoadConfig() (*CFLIPConfig, error) {
	// Check if file exists
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		config := NewCFLIPConfig()
		// Save default config for future use
		if err := m.SaveConfig(config); err != nil {
			return config, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Read file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse TOML
	var config CFLIPConfig
	if _, err := toml.Decode(string(data), &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Initialize maps if nil
	if config.Models == nil {
		config.Models = make(map[string]ModelConfig)
	}
	if config.Providers == nil {
		config.Providers = make(map[string]ProviderInfo)
	}
	if config.Active.ModelMapping == nil {
		config.Active.ModelMapping = make(map[string]string)
	}
	if config.Active.EnvVars == nil {
		config.Active.EnvVars = make(map[string]string)
	}
	if config.UserPreferences.FavoriteProviders == nil {
		config.UserPreferences.FavoriteProviders = []string{}
	}
	if config.UserPreferences.DefaultModelCategories == nil {
		config.UserPreferences.DefaultModelCategories = []string{}
	}

	// Decrypt API keys if needed
	if config.Settings.SecureStorage {
		m.decryptAPIKeys(&config)
	}

	return &config, nil
}

// SaveConfig saves the CFLIP configuration to file
func (m *TOMLManagerV2) SaveConfig(config *CFLIPConfig) error {
	// Validate configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Update timestamp
	config.UpdateTimestamp()

	// For security, encrypt API keys before saving
	if config.Settings.SecureStorage {
		m.encryptAPIKeys(config)
	}

	// Marshal to TOML
	var builder strings.Builder
	encoder := toml.NewEncoder(&builder)
	if err := encoder.Encode(config); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Write to file with atomic operation
	tempFile := m.configPath + ".tmp"
	if err := os.WriteFile(tempFile, []byte(builder.String()), 0600); err != nil {
		return fmt.Errorf("failed to write temp config: %w", err)
	}

	// Rename atomically
	if err := os.Rename(tempFile, m.configPath); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename config: %w", err)
	}

	return nil
}

// GetProvider returns a provider configuration by name
func (m *TOMLManagerV2) GetProvider(name string) (*ProviderInfo, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	provider, exists := config.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}

	return &provider, nil
}

// SaveProvider saves a provider configuration
func (m *TOMLManagerV2) SaveProvider(name string, provider *ProviderInfo) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	// Save provider
	config.Providers[name] = *provider

	return m.SaveConfig(config)
}

// SetActiveProvider sets the active provider
func (m *TOMLManagerV2) SetActiveProvider(name string) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	if err := config.SetActiveProvider(name); err != nil {
		return err
	}

	return m.SaveConfig(config)
}

// GetActiveProvider returns the current active provider
func (m *TOMLManagerV2) GetActiveProvider() (*ProviderInfo, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.GetActiveProvider()
}

// SetActiveModel sets the active model for a category
func (m *TOMLManagerV2) SetActiveModel(category, modelID string) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	if err := config.SetActiveModel(category, modelID); err != nil {
		return err
	}

	return m.SaveConfig(config)
}

// GetActiveModel returns the active model for a category
func (m *TOMLManagerV2) GetActiveModel(category string) (*ModelConfig, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.GetActiveModel(category)
}

// ListProviders returns all available providers
func (m *TOMLManagerV2) ListProviders() (map[string]ProviderInfo, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.Providers, nil
}

// ListModels returns all available models
func (m *TOMLManagerV2) ListModels() (map[string]ModelConfig, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.Models, nil
}

// GetModelsByProvider returns all models for a specific provider
func (m *TOMLManagerV2) GetModelsByProvider(providerName string) ([]ModelConfig, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.GetModelsByProvider(providerName), nil
}

// GetModelsByCategory returns all models in a specific category
func (m *TOMLManagerV2) GetModelsByCategory(category string) ([]ModelConfig, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return config.GetModelsByCategory(category), nil
}

// UpdateSettings updates global settings
func (m *TOMLManagerV2) UpdateSettings(settings SettingsConfig) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	config.Settings = settings
	return m.SaveConfig(config)
}

// GetSettings returns current global settings
func (m *TOMLManagerV2) GetSettings() (*SettingsConfig, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return &config.Settings, nil
}

// UpdatePreferences updates user preferences
func (m *TOMLManagerV2) UpdatePreferences(prefs UserPreferences) error {
	config, err := m.LoadConfig()
	if err != nil {
		return err
	}

	config.UserPreferences = prefs
	return m.SaveConfig(config)
}

// GetPreferences returns current user preferences
func (m *TOMLManagerV2) GetPreferences() (*UserPreferences, error) {
	config, err := m.LoadConfig()
	if err != nil {
		return nil, err
	}

	return &config.UserPreferences, nil
}

// encryptAPIKeys encrypts API keys in the configuration
func (m *TOMLManagerV2) encryptAPIKeys(config *CFLIPConfig) {
	for name, provider := range config.Providers {
		if provider.Auth.Method == AuthMethodAPIKey && provider.Auth.APIKey != "" {
			provider := provider // Create a copy to modify
			provider.Auth.APIKey = m.obfuscateAPIKey(provider.Auth.APIKey)
			config.Providers[name] = provider
		}
	}
}

// decryptAPIKeys decrypts API keys in the configuration
func (m *TOMLManagerV2) decryptAPIKeys(config *CFLIPConfig) {
	for name, provider := range config.Providers {
		if provider.Auth.Method == AuthMethodAPIKey && provider.Auth.APIKey != "" {
			provider := provider // Create a copy to modify
			provider.Auth.APIKey = m.deobfuscateAPIKey(provider.Auth.APIKey)
			config.Providers[name] = provider
		}
	}
}

// obfuscateAPIKey simple obfuscation for API keys
func (m *TOMLManagerV2) obfuscateAPIKey(key string) string {
	// This is a simple obfuscation, not true encryption
	// In production, use proper encryption with a secure key
	hash := sha3.New256()
	hash.Write([]byte(key))
	hash.Write([]byte("cflip-salt-v2")) // Salt the hash
	result := hash.Sum(nil)

	// Store prefix + hash for verification
	if len(key) > 8 {
		return "encrypted:" + string(result) + ":" + key[:8]
	}
	return "encrypted:" + string(result) + ":"
}

// deobfuscateAPIKey reverses the obfuscation
func (m *TOMLManagerV2) deobfuscateAPIKey(obfuscated string) string {
	// In a real implementation, you would properly decrypt
	// For now, check if it's encrypted and return a placeholder if we can't decrypt
	if strings.HasPrefix(obfuscated, "encrypted:") {
		// In production, implement proper decryption
		// For now, return empty to trigger re-authentication
		return ""
	}
	return obfuscated
}