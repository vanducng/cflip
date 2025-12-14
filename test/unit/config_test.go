package unit

import (
	"os"
	"testing"

	"github.com/vanducng/cflip/internal/config"
)

func TestNewConfig(t *testing.T) {
	cfg := config.NewConfig()

	if cfg.SettingsPath == "" {
		t.Error("SettingsPath should not be empty")
	}

	if cfg.BackupDir == "" {
		t.Error("BackupDir should not be empty")
	}

	if cfg.MaxBackups <= 0 {
		t.Error("MaxBackups should be positive")
	}
}

func TestProviderValidation(t *testing.T) {
	// Valid provider
	validProvider := &config.Provider{
		Name:    "test",
		BaseURL: "https://api.example.com",
		Models: map[string]string{
			"haiku":  "test-haiku",
			"sonnet": "test-sonnet",
			"opus":   "test-opus",
		},
	}

	if err := validProvider.Validate(); err != nil {
		t.Errorf("Valid provider should not return error: %v", err)
	}

	// Invalid provider - missing name
	invalidProvider := &config.Provider{
		BaseURL: "https://api.example.com",
		Models: map[string]string{
			"haiku":  "test-haiku",
			"sonnet": "test-sonnet",
			"opus":   "test-opus",
		},
	}

	if err := invalidProvider.Validate(); err == nil {
		t.Error("Provider without name should return error")
	}

	// Invalid provider - missing models
	invalidProvider = &config.Provider{
		Name:    "test",
		BaseURL: "https://api.example.com",
		Models:  make(map[string]string),
	}

	if err := invalidProvider.Validate(); err == nil {
		t.Error("Provider without models should return error")
	}
}

func TestProviderMerge(t *testing.T) {
	provider := &config.Provider{
		BaseURL: "https://api.test.com",
		Models: map[string]string{
			"haiku":  "test-haiku",
			"sonnet": "test-sonnet",
			"opus":   "test-opus",
		},
		EnvVars: map[string]string{
			"CUSTOM_VAR": "custom_value",
		},
	}

	settings := provider.Merge("test-api-key")

	if settings.Env["ANTHROPIC_AUTH_TOKEN"] != "test-api-key" {
		t.Error("API key not set correctly")
	}

	if settings.Env["ANTHROPIC_BASE_URL"] != "https://api.test.com" {
		t.Error("Base URL not set correctly")
	}

	if settings.Env["ANTHROPIC_DEFAULT_HAIKU_MODEL"] != "test-haiku" {
		t.Error("Haiku model not set correctly")
	}

	if settings.Env["CUSTOM_VAR"] != "custom_value" {
		t.Error("Custom env var not set correctly")
	}

	if settings.Env["API_TIMEOUT_MS"] != "3000000" {
		t.Error("Default timeout not set")
	}
}

func TestConfigManager(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cflip-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	manager := config.NewManager()

	// Test loading non-existent file
	settings, err := manager.LoadSettings()
	if err != nil {
		t.Errorf("Loading non-existent file should not error: %v", err)
	}

	if settings.Env == nil {
		t.Error("Settings.Env should be initialized")
	}

	// Test saving and loading
	settings.Env["TEST"] = "value"

	// We can't easily test without modifying the package structure
	// For now, just verify the structure is correct
	if manager.GetSettingsPath() == "" {
		t.Error("SettingsPath should not be empty")
	}
}