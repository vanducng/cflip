package config

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/vanducng/cflip/pkg/utils"
)

// BackupManager handles backup-specific operations
type BackupManager struct {
	manager *Manager
}

// NewBackupManager creates a new backup manager
func NewBackupManager(manager *Manager) *BackupManager {
	return &BackupManager{
		manager: manager,
	}
}

// CreateWithDescription creates a backup with a description
func (bm *BackupManager) CreateWithDescription(description string) (*BackupInfo, error) {
	backup, err := bm.manager.CreateBackup()
	if err != nil {
		return nil, err
	}

	// Add description to backup ID
	if description != "" {
		// Sanitize description
		description = strings.ReplaceAll(description, " ", "_")
		description = strings.ReplaceAll(description, "/", "-")
		newID := fmt.Sprintf("%s-%s", backup.ID, description)
		newPath := bm.manager.config.BackupDir + "/" + newID + ".json"

		// Rename backup file
		if err := utils.RenameFile(backup.Path, newPath); err == nil {
			backup.ID = newID
			backup.Path = newPath
		}
	}

	return backup, nil
}

// GetLatestBackup returns the most recent backup
func (bm *BackupManager) GetLatestBackup() (*BackupInfo, error) {
	backups, err := bm.manager.ListBackups()
	if err != nil {
		return nil, err
	}

	if len(backups) == 0 {
		return nil, fmt.Errorf("no backups found")
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp > backups[j].Timestamp
	})

	return backups[0], nil
}

// DeleteBackup removes a backup
func (bm *BackupManager) DeleteBackup(backupID string) error {
	backupPath := bm.manager.config.BackupDir + "/" + backupID + ".json"

	if err := utils.RemoveFile(backupPath); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	return nil
}

// PruneBackups removes backups older than the specified duration
func (bm *BackupManager) PruneBackups(olderThan time.Duration) error {
	backups, err := bm.manager.ListBackups()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-olderThan)
	var deleted []string

	for _, backup := range backups {
		// Parse timestamp from backup ID
		timestamp, err := time.Parse("20060102-150405", backup.Timestamp)
		if err != nil {
			continue
		}

		if timestamp.Before(cutoff) {
			if err := bm.DeleteBackup(backup.ID); err == nil {
				deleted = append(deleted, backup.ID)
			}
		}
	}

	return nil
}

// BackupStats provides statistics about backups
type BackupStats struct {
	TotalCount   int           `json:"totalCount"`
	TotalSize    int64         `json:"totalSize"`
	OldestBackup time.Time     `json:"oldestBackup"`
	NewestBackup time.Time     `json:"newestBackup"`
	ByProvider   map[string]int `json:"byProvider"`
}

// GetStats returns backup statistics
func (bm *BackupManager) GetStats() (*BackupStats, error) {
	backups, err := bm.manager.ListBackups()
	if err != nil {
		return nil, err
	}

	stats := &BackupStats{
		TotalCount: len(backups),
		ByProvider: make(map[string]int),
	}

	if len(backups) == 0 {
		return stats, nil
	}

	var oldestTime, newestTime time.Time
	first := true

	for _, backup := range backups {
		stats.TotalSize += backup.Size
		stats.ByProvider[backup.Provider]++

		// Parse timestamp
		timestamp, err := time.Parse("20060102-150405", backup.Timestamp)
		if err != nil {
			continue
		}

		if first {
			oldestTime = timestamp
			newestTime = timestamp
			first = false
		} else {
			if timestamp.Before(oldestTime) {
				oldestTime = timestamp
			}
			if timestamp.After(newestTime) {
				newestTime = timestamp
			}
		}
	}

	stats.OldestBackup = oldestTime
	stats.NewestBackup = newestTime

	return stats, nil
}