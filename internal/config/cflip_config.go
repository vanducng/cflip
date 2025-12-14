package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuthMethod represents the authentication method
type AuthMethod string

const (
	AuthMethodAPIKey      AuthMethod = "api_key"
	AuthMethodSubscription AuthMethod = "subscription"
)

// ModelConfig represents a model configuration
type ModelConfig struct {
	ID           string            `toml:"id"`
	Name         string            `toml:"name"`
	Provider     string            `toml:"provider"`
	Category     string            `toml:"category"` // haiku, sonnet, opus, custom
	Description  string            `toml:"description"`
	MaxTokens    int               `toml:"max_tokens,omitempty"`
	ContextWindow int               `toml:"context_window,omitempty"`
	Capabilities []string          `toml:"capabilities,omitempty"`
	CustomParams map[string]string `toml:"custom_params,omitempty"`
}

// ProviderAuthConfig represents authentication configuration for a provider
type ProviderAuthConfig struct {
	Method           AuthMethod `toml:"method"`
	APIKey           string     `toml:"api_key,omitempty"`
	BaseURL          string     `toml:"base_url,omitempty"`
	AuthHeader       string     `toml:"auth_header,omitempty"`
	TimeoutSeconds   int        `toml:"timeout_seconds"`
	RateLimitRPM     int        `toml:"rate_limit_rpm,omitempty"`
	RateLimitTPM     int        `toml:"rate_limit_tpm,omitempty"`
	RequiresSetup    bool       `toml:"requires_setup"`
	SetupInstructions string    `toml:"setup_instructions,omitempty"`
	LastValidated    time.Time  `toml:"last_validated,omitempty"`
}

// ProviderInfo represents provider information
type ProviderInfo struct {
	Name        string            `toml:"name"`
	DisplayName string            `toml:"display_name"`
	Description string            `toml:"description"`
	Website     string            `toml:"website,omitempty"`
	Auth        ProviderAuthConfig `toml:"auth"`
	Models      []string          `toml:"models"` // List of model IDs
	EnvVars     map[string]string `toml:"env_vars,omitempty"`
	Tags        []string          `toml:"tags,omitempty"`
}

// ActiveConfig represents the current active configuration
type ActiveConfig struct {
	Provider     string            `toml:"provider"`
	ModelMapping map[string]string `toml:"model_mapping"` // category -> model_id
	EnvVars      map[string]string `toml:"env_vars,omitempty"`
	LastSwitched time.Time         `toml:"last_switched"`
}

// SettingsConfig represents global settings
type SettingsConfig struct {
	BackupDirectory string        `toml:"backup_directory"`
	MaxBackups      int           `toml:"max_backups"`
	AutoBackup      bool          `toml:"auto_backup"`
	SecureStorage   bool          `toml:"secure_storage"`
	DefaultTimeout  int           `toml:"default_timeout"`
	AutoValidate    bool          `toml:"auto_validate"`
	LogLevel        string        `toml:"log_level"`
	Telemetry       bool          `toml:"telemetry"`
	LastUpdateCheck time.Time     `toml:"last_update_check,omitempty"`
}

// CFLIPConfig represents the main configuration file structure
type CFLIPConfig struct {
	Version        string                    `toml:"version"`
	CreatedAt      time.Time                 `toml:"created_at"`
	UpdatedAt      time.Time                 `toml:"updated_at"`
	Models         map[string]ModelConfig    `toml:"models"`
	Providers      map[string]ProviderInfo   `toml:"providers"`
	Active         ActiveConfig              `toml:"active"`
	Settings       SettingsConfig            `toml:"settings"`
	UserPreferences UserPreferences           `toml:"user_preferences"`
}

// UserPreferences represents user-specific preferences
type UserPreferences struct {
	DefaultModelCategories []string `toml:"default_model_categories"`
	FavoriteProviders     []string `toml:"favorite_providers"`
	AutoSwitchInterval    int      `toml:"auto_switch_interval_hours,omitempty"`
	PromptOnSwitch        bool     `toml:"prompt_on_switch"`
	ShowModelInfo         bool     `toml:"show_model_info"`
	ColorOutput           bool     `toml:"color_output"`
}

// NewCFLIPConfig creates a new default configuration
func NewCFLIPConfig() *CFLIPConfig {
	now := time.Now()
	homeDir, _ := os.UserHomeDir()

	return &CFLIPConfig{
		Version:   "1.0.0",
		CreatedAt: now,
		UpdatedAt: now,

		// Centralized model catalog
		Models: map[string]ModelConfig{
			// Anthropic models
			"claude-3-5-haiku-20241022": {
				ID:            "claude-3-5-haiku-20241022",
				Name:          "Claude 3.5 Haiku",
				Provider:      "anthropic",
				Category:      "haiku",
				Description:   "Fast and efficient model for quick tasks",
				MaxTokens:     200000,
				ContextWindow: 200000,
				Capabilities:  []string{"text", "code", "analysis"},
			},
			"claude-3-5-sonnet-20241022": {
				ID:            "claude-3-5-sonnet-20241022",
				Name:          "Claude 3.5 Sonnet",
				Provider:      "anthropic",
				Category:      "sonnet",
				Description:   "Balanced model for most tasks",
				MaxTokens:     200000,
				ContextWindow: 200000,
				Capabilities:  []string{"text", "code", "analysis", "reasoning"},
			},
			"claude-3-opus-20240229": {
				ID:            "claude-3-opus-20240229",
				Name:          "Claude 3 Opus",
				Provider:      "anthropic",
				Category:      "opus",
				Description:   "Most capable model for complex tasks",
				MaxTokens:     4096,
				ContextWindow: 200000,
				Capabilities:  []string{"text", "code", "analysis", "reasoning", "creative"},
			},

			// GLM models
			"glm-4.5-air": {
				ID:            "glm-4.5-air",
				Name:          "GLM-4.5 Air",
				Provider:      "glm",
				Category:      "haiku",
				Description:   "Lightweight model for fast responses",
				MaxTokens:     8192,
				ContextWindow: 128000,
				Capabilities:  []string{"text", "code"},
			},
			"glm-4.6": {
				ID:            "glm-4.6",
				Name:          "GLM-4.6",
				Provider:      "glm",
				Category:      "sonnet",
				Description:   "Advanced model for complex tasks",
				MaxTokens:     8192,
				ContextWindow: 128000,
				Capabilities:  []string{"text", "code", "reasoning"},
			},

			// Claude Code subscription (uses default models)
			"claude-code-default": {
				ID:           "claude-code-default",
				Name:         "Claude Code Default",
				Provider:     "claude-code",
				Category:     "sonnet",
				Description:  "Default model from Claude Code subscription",
				Capabilities: []string{"text", "code", "analysis", "reasoning"},
			},
		},

		// Provider configurations
		Providers: map[string]ProviderInfo{
			"anthropic": {
				Name:        "anthropic",
				DisplayName: "Anthropic",
				Description: "Official Anthropic Claude API",
				Website:     "https://anthropic.com",
				Auth: ProviderAuthConfig{
					Method:         AuthMethodAPIKey,
					BaseURL:        "https://api.anthropic.com",
					AuthHeader:     "x-api-key",
					TimeoutSeconds: 300,
					RateLimitRPM:   1000,
					RateLimitTPM:   100000,
					RequiresSetup:  false,
				},
				Models:  []string{"claude-3-5-haiku-20241022", "claude-3-5-sonnet-20241022", "claude-3-opus-20240229"},
				EnvVars: map[string]string{"API_TIMEOUT_MS": "3000000"},
				Tags:    []string{"official", "api", "paid"},
			},
			"claude-code": {
				Name:        "claude-code",
				DisplayName: "Claude Code",
				Description: "Claude Code CLI with subscription authentication",
				Website:     "https://docs.anthropic.com/claude/docs/claude-code",
				Auth: ProviderAuthConfig{
					Method:            AuthMethodSubscription,
					TimeoutSeconds:    300,
					RequiresSetup:     true,
					SetupInstructions: "Run 'claude /login' to authenticate with your Claude subscription",
				},
				Models: []string{"claude-code-default"},
				Tags:   []string{"official", "subscription", "cli"},
			},
			"glm": {
				Name:        "glm",
				DisplayName: "GLM by z.ai",
				Description: "GLM models compatible with Claude API",
				Website:     "https://z.ai",
				Auth: ProviderAuthConfig{
					Method:         AuthMethodAPIKey,
					BaseURL:        "https://api.z.ai/api/anthropic",
					AuthHeader:     "authorization",
					TimeoutSeconds: 300,
					RateLimitRPM:   500,
					RateLimitTPM:   50000,
					RequiresSetup:  false,
				},
				Models:  []string{"glm-4.5-air", "glm-4.6"},
				EnvVars: map[string]string{"API_TIMEOUT_MS": "3000000"},
				Tags:    []string{"third-party", "api", "paid"},
			},
		},

		// Active configuration
		Active: ActiveConfig{
			Provider: "anthropic",
			ModelMapping: map[string]string{
				"haiku":  "claude-3-5-haiku-20241022",
				"sonnet": "claude-3-5-sonnet-20241022",
				"opus":   "claude-3-opus-20240229",
			},
			LastSwitched: now,
		},

		// Global settings
		Settings: SettingsConfig{
			BackupDirectory: filepath.Join(homeDir, ".cflip", "backups"),
			MaxBackups:      10,
			AutoBackup:      true,
			SecureStorage:   true,
			DefaultTimeout:  300,
			AutoValidate:    true,
			LogLevel:        "info",
			Telemetry:       false,
		},

		// User preferences
		UserPreferences: UserPreferences{
			DefaultModelCategories: []string{"sonnet", "haiku"},
			FavoriteProviders:     []string{"anthropic", "claude-code"},
			PromptOnSwitch:        true,
			ShowModelInfo:         true,
			ColorOutput:           true,
		},
	}
}

// GetConfigPath returns the path to the cflip configuration file
func GetConfigPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".cflip", "config.toml")
}

// GetLegacySettingsPath returns the path to the legacy Claude settings file
func GetLegacySettingsPath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude", "settings.json")
}


// Helper methods for CFLIPConfig

// GetActiveProvider returns the active provider configuration
func (c *CFLIPConfig) GetActiveProvider() (*ProviderInfo, error) {
	provider, exists := c.Providers[c.Active.Provider]
	if !exists {
		return nil, fmt.Errorf("active provider '%s' not found", c.Active.Provider)
	}
	return &provider, nil
}

// SetActiveProvider sets the active provider and initializes model mappings
func (c *CFLIPConfig) SetActiveProvider(providerName string) error {
	provider, exists := c.Providers[providerName]
	if !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	c.Active.Provider = providerName
	c.Active.LastSwitched = time.Now()

	// Initialize model mappings based on provider's models
	if c.Active.ModelMapping == nil {
		c.Active.ModelMapping = make(map[string]string)
	}

	// Map models by category
	for _, modelID := range provider.Models {
		if model, exists := c.Models[modelID]; exists {
			c.Active.ModelMapping[model.Category] = modelID
		}
	}

	return nil
}

// GetModelConfig returns a model configuration by ID
func (c *CFLIPConfig) GetModelConfig(modelID string) (*ModelConfig, error) {
	model, exists := c.Models[modelID]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found", modelID)
	}
	return &model, nil
}

// GetModelsByCategory returns all models in a specific category
func (c *CFLIPConfig) GetModelsByCategory(category string) []ModelConfig {
	var models []ModelConfig
	for _, model := range c.Models {
		if model.Category == category {
			models = append(models, model)
		}
	}
	return models
}

// GetModelsByProvider returns all models for a specific provider
func (c *CFLIPConfig) GetModelsByProvider(providerName string) []ModelConfig {
	var models []ModelConfig
	for _, model := range c.Models {
		if model.Provider == providerName {
			models = append(models, model)
		}
	}
	return models
}

// GetActiveModel returns the active model for a category
func (c *CFLIPConfig) GetActiveModel(category string) (*ModelConfig, error) {
	modelID, exists := c.Active.ModelMapping[category]
	if !exists {
		return nil, fmt.Errorf("no active model for category '%s'", category)
	}
	return c.GetModelConfig(modelID)
}

// SetActiveModel sets the active model for a category
func (c *CFLIPConfig) SetActiveModel(category, modelID string) error {
	model, exists := c.Models[modelID]
	if !exists {
		return fmt.Errorf("model '%s' not found", modelID)
	}

	// Ensure model category matches
	if model.Category != category {
		return fmt.Errorf("model '%s' is not in category '%s' (it's in '%s')",
			modelID, category, model.Category)
	}

	if c.Active.ModelMapping == nil {
		c.Active.ModelMapping = make(map[string]string)
	}
	c.Active.ModelMapping[category] = modelID
	c.UpdatedAt = time.Now()

	return nil
}

// Validate validates the entire configuration
func (c *CFLIPConfig) Validate() error {
	// Check if active provider exists
	if _, err := c.GetActiveProvider(); err != nil {
		return fmt.Errorf("invalid active provider: %w", err)
	}

	// Validate providers
	for name, provider := range c.Providers {
		if err := c.validateProvider(&provider); err != nil {
			return fmt.Errorf("provider '%s': %w", name, err)
		}
	}

	// Validate models
	for id, model := range c.Models {
		if err := c.validateModel(&model); err != nil {
			return fmt.Errorf("model '%s': %w", id, err)
		}
	}

	// Validate active model mappings
	for category, modelID := range c.Active.ModelMapping {
		if _, exists := c.Models[modelID]; !exists {
			return fmt.Errorf("active model mapping for '%s': model '%s' not found",
				category, modelID)
		}
	}

	return nil
}

func (c *CFLIPConfig) validateProvider(provider *ProviderInfo) error {
	if provider.Name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	// Validate auth configuration
	if provider.Auth.Method == AuthMethodAPIKey {
		if provider.Auth.BaseURL == "" {
			return fmt.Errorf("base URL is required for API key authentication")
		}
		if provider.Auth.AuthHeader == "" {
			return fmt.Errorf("auth header is required for API key authentication")
		}
	}

	// Validate models exist
	for _, modelID := range provider.Models {
		if _, exists := c.Models[modelID]; !exists {
			return fmt.Errorf("references non-existent model '%s'", modelID)
		}
	}

	return nil
}

func (c *CFLIPConfig) validateModel(model *ModelConfig) error {
	if model.ID == "" {
		return fmt.Errorf("model ID cannot be empty")
	}
	if model.Name == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if model.Provider == "" {
		return fmt.Errorf("model provider cannot be empty")
	}
	if model.Category == "" {
		return fmt.Errorf("model category cannot be empty")
	}

	// Validate provider exists
	if _, exists := c.Providers[model.Provider]; !exists {
		return fmt.Errorf("provider '%s' does not exist", model.Provider)
	}

	return nil
}

// UpdateTimestamp updates the configuration timestamp
func (c *CFLIPConfig) UpdateTimestamp() {
	c.UpdatedAt = time.Now()
}

// GetProviderByName returns a provider by name
func (c *CFLIPConfig) GetProviderByName(name string) (*ProviderInfo, error) {
	provider, exists := c.Providers[name]
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	return &provider, nil
}

// ListProviders returns all provider names
func (c *CFLIPConfig) ListProviders() []string {
	var names []string
	for name := range c.Providers {
		names = append(names, name)
	}
	return names
}

// IsAPIKeyRequired returns true if the provider requires an API key
func (p *ProviderInfo) IsAPIKeyRequired() bool {
	return p.Auth.Method == AuthMethodAPIKey
}

// HasAPIKey returns true if the provider has an API key configured
func (p *ProviderInfo) HasAPIKey() bool {
	return p.Auth.APIKey != ""
}

// SetAPIKey sets the API key for a provider
func (p *ProviderInfo) SetAPIKey(apiKey string) {
	p.Auth.APIKey = apiKey
}

// GetAPIKey returns the API key for a provider
func (p *ProviderInfo) GetAPIKey() string {
	return p.Auth.APIKey
}

// ClearAPIKey removes the API key from the provider
func (p *ProviderInfo) ClearAPIKey() {
	p.Auth.APIKey = ""
}