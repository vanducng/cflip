package config

import (
	"os"
	"path/filepath"
)

// ClaudeSettings represents the structure of ~/.claude/settings.json
type ClaudeSettings struct {
	Env map[string]string `json:"env"`
}

// Provider represents a Claude Code provider configuration
type Provider struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"displayName"`
	BaseURL     string            `json:"baseUrl"`
	Models      map[string]string `json:"models"`      // haiku, sonnet, opus
	AuthHeader  string            `json:"authHeader"`  // e.g., "x-api-key" or "authorization"
	EnvVars     map[string]string `json:"envVars"`     // Additional environment variables
}

// Config represents the application configuration
type Config struct {
	SettingsPath    string   `json:"settingsPath"`
	BackupDir       string   `json:"backupDir"`
	MaxBackups      int      `json:"maxBackups"`
	CurrentProvider string   `json:"currentProvider"`
	Providers       []string `json:"providers"`
}

// BackupInfo represents information about a configuration backup
type BackupInfo struct {
	ID        string    `json:"id"`
	Timestamp string    `json:"timestamp"`
	Provider  string    `json:"provider"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		SettingsPath: filepath.Join(homeDir, ".claude", "settings.json"),
		BackupDir:    filepath.Join(homeDir, ".claude", "backups"),
		MaxBackups:   10,
		Providers:    []string{"anthropic", "glm"},
	}
}

// Validate checks if the provider configuration is valid
func (p *Provider) Validate() error {
	if p.Name == "" {
		return &ValidationError{Field: "name", Message: "provider name cannot be empty"}
	}
	if p.BaseURL == "" {
		return &ValidationError{Field: "baseUrl", Message: "base URL cannot be empty"}
	}
	if p.Models == nil || len(p.Models) == 0 {
		return &ValidationError{Field: "models", Message: "at least one model must be specified"}
	}
	// Check for required models
	requiredModels := []string{"haiku", "sonnet", "opus"}
	for _, model := range requiredModels {
		if _, exists := p.Models[model]; !exists {
			return &ValidationError{
				Field:   "models",
				Message: "missing required model: " + model,
			}
		}
	}
	return nil
}

// Merge combines provider configuration into Claude settings
func (p *Provider) Merge(apiKey string) *ClaudeSettings {
	settings := &ClaudeSettings{
		Env: make(map[string]string),
	}

	// Set authentication token
	settings.Env["ANTHROPIC_AUTH_TOKEN"] = apiKey

	// Set base URL
	settings.Env["ANTHROPIC_BASE_URL"] = p.BaseURL

	// Set default models
	if p.Models != nil {
		if haiku, exists := p.Models["haiku"]; exists {
			settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] = haiku
		}
		if sonnet, exists := p.Models["sonnet"]; exists {
			settings.Env["ANTHROPIC_DEFAULT_SONNET_MODEL"] = sonnet
		}
		if opus, exists := p.Models["opus"]; exists {
			settings.Env["ANTHROPIC_DEFAULT_OPUS_MODEL"] = opus
		}
	}

	// Set additional environment variables
	if p.EnvVars != nil {
		for key, value := range p.EnvVars {
			settings.Env[key] = value
		}
	}

	// Set default timeout if not specified
	if _, exists := settings.Env["API_TIMEOUT_MS"]; !exists {
		settings.Env["API_TIMEOUT_MS"] = "3000000"
	}

	return settings
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}