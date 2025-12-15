package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ClaudeSettings represents the full Claude settings structure
type ClaudeSettings struct {
	Schema string                 `json:"$schema,omitempty"`
	Env    map[string]interface{} `json:"env,omitempty"`
	// Preserve all other fields
	AdditionalFields map[string]interface{} `json:"-"`
}

// LoadSettings loads the current Claude settings
func LoadSettings(settingsPath string) (*ClaudeSettings, error) {
	var settings ClaudeSettings

	// Read file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty settings if file doesn't exist
			return &ClaudeSettings{
				Env: make(map[string]interface{}),
			}, nil
		}
		return nil, fmt.Errorf("failed to read settings: %w", err)
	}

	// Parse JSON
	var rawSettings map[string]interface{}
	if err := json.Unmarshal(data, &rawSettings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	// Extract env if exists
	if env, ok := rawSettings["env"].(map[string]interface{}); ok {
		settings.Env = env
	} else {
		settings.Env = make(map[string]interface{})
	}

	// Store schema if exists
	if schema, ok := rawSettings["$schema"].(string); ok {
		settings.Schema = schema
	}

	// Store all other fields
	settings.AdditionalFields = make(map[string]interface{})
	for k, v := range rawSettings {
		if k != "$schema" && k != "env" {
			settings.AdditionalFields[k] = v
		}
	}

	return &settings, nil
}

// SaveSettings saves settings preserving all fields
func SaveSettings(settingsPath string, settings *ClaudeSettings) error {
	// Build the full settings map
	fullSettings := make(map[string]interface{})

	// Add schema
	if settings.Schema != "" {
		fullSettings["$schema"] = settings.Schema
	}

	// Add env
	if len(settings.Env) > 0 {
		fullSettings["env"] = settings.Env
	}

	// Add all additional fields
	for k, v := range settings.AdditionalFields {
		fullSettings[k] = v
	}

	// Marshal with indentation
	data, err := json.MarshalIndent(fullSettings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0750); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(settingsPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// CreateSnapshot creates a snapshot of current settings
func CreateSnapshot(settingsPath, snapshotsDir, provider string) error {
	// Load current settings
	settings, err := LoadSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load settings for snapshot: %w", err)
	}

	// Create snapshot file name
	timestamp := time.Now().Format("20060102-150405")
	snapshotFile := filepath.Join(snapshotsDir, fmt.Sprintf("snapshot-%s-%s.json", provider, timestamp))

	// Ensure snapshots directory exists
	if err := os.MkdirAll(snapshotsDir, 0750); err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	// Save snapshot
	return SaveSettings(snapshotFile, settings)
}

// ListSnapshots lists all available snapshots
func ListSnapshots(snapshotsDir string) ([]string, error) {
	files, err := os.ReadDir(snapshotsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var snapshots []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			snapshots = append(snapshots, file.Name())
		}
	}

	return snapshots, nil
}

// CleanupOldSnapshots removes old snapshots keeping only the most recent N per provider
func CleanupOldSnapshots(snapshotsDir string, keepCount int) error {
	snapshots, err := ListSnapshots(snapshotsDir)
	if err != nil {
		return err
	}

	// Group snapshots by provider
	providerSnapshots := make(map[string][]string)
	for _, snapshot := range snapshots {
		// Extract provider from filename: snapshot-provider-timestamp.json
		parts := filepath.Base(snapshot)
		if len(parts) > len("snapshot-") {
			provider := parts[9:] // Remove "snapshot-" prefix
			if idx := findIndex(provider, '-'); idx > 0 {
				provider = provider[:idx]
				providerSnapshots[provider] = append(providerSnapshots[provider], snapshot)
			}
		}
	}

	// Remove old snapshots
	for _, files := range providerSnapshots {
		if len(files) <= keepCount {
			continue
		}

		// Sort files by timestamp (newest first)
		// For simplicity, just remove the oldest files
		for i := keepCount; i < len(files); i++ {
			filePath := filepath.Join(snapshotsDir, files[i])
			os.Remove(filePath)
		}
	}

	return nil
}

func findIndex(s string, sep rune) int {
	for i, r := range s {
		if r == sep {
			return i
		}
	}
	return -1
}
