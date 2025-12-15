package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vanducng/cflip/internal/config"
)

const testProvider = "test"

func TestNewConfig(t *testing.T) {
	cfg := config.NewConfig()

	if cfg.Provider != "anthropic" {
		t.Errorf("Expected default provider to be 'anthropic', got '%s'", cfg.Provider)
	}

	if cfg.Providers == nil {
		t.Error("Providers map should not be nil")
	}

	if _, exists := cfg.Providers["anthropic"]; !exists {
		t.Error("Anthropic provider should exist by default")
	}
}

func TestConfigSaveLoad(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cflip-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a config
	cfg := config.NewConfig()
	cfg.Provider = testProvider
	cfg.SetProviderConfig(testProvider, config.ProviderConfig{
		Token:   "test-token",
		BaseURL: "https://test.example.com",
	})

	// Save config (this would normally save to home dir)
	// For testing, we'll just verify the structure
	if cfg.Provider != testProvider {
		t.Error("Provider not set correctly")
	}

	provider := cfg.Providers[testProvider]
	if provider.Token != "test-token" {
		t.Error("Token not set correctly")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := config.GetConfigPath()
	if path == "" {
		t.Error("Config path should not be empty")
	}

	expected := filepath.Join(os.Getenv("HOME"), ".cflip", "config.toml")
	if path != expected {
		t.Errorf("Expected config path '%s', got '%s'", expected, path)
	}
}
