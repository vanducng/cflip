package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles Claude settings file operations
type Manager struct {
	config *Config
}

// NewManager creates a new configuration manager
func NewManager() *Manager {
	return &Manager{
		config: NewConfig(),
	}
}

// GetSettingsPath returns the path to the Claude settings file
func (m *Manager) GetSettingsPath() string {
	return m.config.SettingsPath
}

// LoadSettings reads the current Claude settings
func (m *Manager) LoadSettings() (*ClaudeSettings, error) {
	// Check if file exists
	if _, err := os.Stat(m.config.SettingsPath); os.IsNotExist(err) {
		// Return empty settings if file doesn't exist
		return &ClaudeSettings{Env: make(map[string]string)}, nil
	}

	// Open and read the file
	file, err := os.Open(m.config.SettingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open settings file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: failed to close settings file: %v\n", err)
		}
	}()

	// Decode JSON
	var settings ClaudeSettings
	if err := json.NewDecoder(file).Decode(&settings); err != nil {
		return nil, fmt.Errorf("failed to decode settings: %w", err)
	}

	// Initialize env map if nil
	if settings.Env == nil {
		settings.Env = make(map[string]string)
	}

	return &settings, nil
}

// SaveSettings writes Claude settings to file
func (m *Manager) SaveSettings(settings *ClaudeSettings) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.config.SettingsPath), 0750); err != nil {
		return fmt.Errorf("failed to create settings directory: %w", err)
	}

	// Create temporary file
	tempFile := m.config.SettingsPath + ".tmp"
	// Validate temp file path is within expected directory
	if !filepath.IsAbs(tempFile) || !strings.HasPrefix(filepath.Clean(tempFile), filepath.Clean(filepath.Dir(m.config.SettingsPath))) {
		return fmt.Errorf("invalid temp file path: %s", tempFile)
	}
	file, err := os.Create(tempFile) // #nosec G304 - path validated above
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	// Write to temp file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(settings); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close temp file after encode error: %v\n", closeErr)
		}
		if removeErr := os.Remove(tempFile); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file after encode error: %v\n", removeErr)
		}
		return fmt.Errorf("failed to encode settings: %w", err)
	}

	// Close file
	if err := file.Close(); err != nil {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file after close error: %v\n", removeErr)
		}
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Rename temp file to actual file (atomic operation)
	if err := os.Rename(tempFile, m.config.SettingsPath); err != nil {
		if removeErr := os.Remove(tempFile); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file after rename error: %v\n", removeErr)
		}
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// GetCurrentProvider detects the current provider from settings
func (m *Manager) GetCurrentProvider() (string, error) {
	settings, err := m.LoadSettings()
	if err != nil {
		return "", err
	}

	// Check base URL to determine provider
	if baseURL, exists := settings.Env["ANTHROPIC_BASE_URL"]; exists {
		switch baseURL {
		case "https://api.z.ai/api/anthropic":
			return "glm", nil
		case "", "https://api.anthropic.com":
			return "anthropic", nil
		default:
			return "custom", nil
		}
	}

	// Default to anthropic if no base URL is set
	return "anthropic", nil
}

// CreateBackup creates a backup of the current settings
func (m *Manager) CreateBackup() (*BackupInfo, error) {
	// Ensure backup directory exists
	if err := os.MkdirAll(m.config.BackupDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup ID with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupID := fmt.Sprintf("backup-%s", timestamp)
	backupPath := filepath.Join(m.config.BackupDir, backupID+".json")

	// Check if source file exists
	if _, err := os.Stat(m.config.SettingsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("settings file does not exist, cannot create backup")
	}

	// Copy file to backup location
	if err := copyFile(m.config.SettingsPath, backupPath); err != nil {
		return nil, fmt.Errorf("failed to create backup: %w", err)
	}

	// Get current provider
	currentProvider, err := m.GetCurrentProvider()
	if err != nil {
		currentProvider = "unknown"
	}

	// Get file size
	info, err := os.Stat(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get backup file info: %w", err)
	}

	backupInfo := &BackupInfo{
		ID:        backupID,
		Timestamp: timestamp,
		Provider:  currentProvider,
		Path:      backupPath,
		Size:      info.Size(),
	}

	// Clean old backups
	m.cleanOldBackups()

	return backupInfo, nil
}

// ListBackups returns all available backups
func (m *Manager) ListBackups() ([]*BackupInfo, error) {
	var backups []*BackupInfo

	// Check if backup directory exists
	if _, err := os.Stat(m.config.BackupDir); os.IsNotExist(err) {
		return backups, nil
	}

	// Read backup directory
	entries, err := os.ReadDir(m.config.BackupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Process each backup file
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "backup-") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupID := entry.Name()[:len(entry.Name())-5] // Remove .json
		timestamp := backupID[7:] // Remove "backup-" prefix
		backupPath := filepath.Join(m.config.BackupDir, entry.Name())

		backups = append(backups, &BackupInfo{
			ID:        backupID,
			Timestamp: timestamp,
			Provider:  "unknown", // We'd need to load the backup to determine this
			Path:      backupPath,
			Size:      info.Size(),
		})
	}

	return backups, nil
}

// RestoreBackup restores settings from a backup
func (m *Manager) RestoreBackup(backupID string) error {
	backupPath := filepath.Join(m.config.BackupDir, backupID+".json")

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup not found: %s", backupID)
	}

	// Copy backup to settings file
	if err := copyFile(backupPath, m.config.SettingsPath); err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

// cleanOldBackups removes old backups if we exceed the maximum
func (m *Manager) cleanOldBackups() {
	backups, err := m.ListBackups()
	if err != nil {
		return
	}

	if len(backups) <= m.config.MaxBackups {
		return
	}

	// Sort backups by timestamp (oldest first)
	// For now, just remove the oldest files
	for i := 0; i < len(backups)-m.config.MaxBackups; i++ {
		if err := os.Remove(backups[i].Path); err != nil {
			// Log error but continue cleaning up other backups
			fmt.Fprintf(os.Stderr, "Warning: failed to remove old backup %s: %v\n", backups[i].Path, err)
		}
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Validate paths
	if !filepath.IsAbs(src) {
		return fmt.Errorf("source path must be absolute: %s", src)
	}
	if !filepath.IsAbs(dst) {
		return fmt.Errorf("destination path must be absolute: %s", dst)
	}

	source, err := os.Open(src) // #nosec G304 - path validated above
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := source.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close source file %s: %v\n", src, closeErr)
		}
	}()

	destination, err := os.Create(dst) // #nosec G304 - path validated above
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := destination.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close destination file %s: %v\n", dst, closeErr)
		}
	}()

	_, err = io.Copy(destination, source)
	if err != nil {
		return err
	}

	// Sync the file to ensure data is written to disk
	if err := destination.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	return nil
}