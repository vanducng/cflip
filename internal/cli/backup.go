package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/vanducng/cflip/internal/config"
)

var (
	backupDescription string
	backupOlderThan   string
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup [subcommand]",
	Short: "Manage configuration backups",
	Long: `Manage backups of your Claude configuration settings.
Backups are automatically created before switching providers.`,
}

func newBackupCmd() *cobra.Command {
	backupCmd.AddCommand(newBackupCreateCmd())
	backupCmd.AddCommand(newBackupListCmd())
	backupCmd.AddCommand(newBackupRestoreCmd())
	backupCmd.AddCommand(newBackupDeleteCmd())
	backupCmd.AddCommand(newBackupPruneCmd())
	return backupCmd
}

// backupCreateCmd represents the backup create command
var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup of current settings",
	Long: `Create a backup of the current Claude configuration settings.
The backup will be stored in ~/.claude/backups/`,
	RunE: runBackupCreate,
}

func newBackupCreateCmd() *cobra.Command {
	backupCreateCmd.Flags().StringVarP(&backupDescription, "description", "d", "", "Add a description to the backup")
	return backupCreateCmd
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	quiet, _ := cmd.Flags().GetBool("quiet")

	configManager := config.NewManager()
	backupManager := config.NewBackupManager(configManager)

	if !quiet {
		fmt.Print("Creating backup... ")
	}

	var backup *config.BackupInfo
	var err error

	if backupDescription != "" {
		backup, err = backupManager.CreateWithDescription(backupDescription)
	} else {
		backup, err = configManager.CreateBackup()
	}

	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if !quiet {
		fmt.Printf("done\n")
		fmt.Printf("Backup ID: %s\n", backup.ID)
		fmt.Printf("Size: %d bytes\n", backup.Size)
		fmt.Printf("Provider: %s\n", backup.Provider)
	}

	return nil
}

// backupListCmd represents the backup list command
var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available backups",
	Long: `List all available configuration backups with their details
including creation time, provider, and size.`,
	RunE: runBackupList,
}

func newBackupListCmd() *cobra.Command {
	return backupListCmd
}

func runBackupList(cmd *cobra.Command, args []string) error {
	configManager := config.NewManager()

	backups, err := configManager.ListBackups()
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	if len(backups) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	// Create tabwriter for nice formatting
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTIMESTAMP\tPROVIDER\tSIZE")

	for _, backup := range backups {
		// Parse timestamp for better display
		timestamp, _ := time.Parse("20060102-150405", backup.Timestamp)
		fmt.Fprintf(w, "%s\t%s\t%s\t%d bytes\n",
			backup.ID,
			timestamp.Format("2006-01-02 15:04:05"),
			backup.Provider,
			backup.Size,
		)
	}

	w.Flush()

	return nil
}

// backupRestoreCmd represents the backup restore command
var backupRestoreCmd = &cobra.Command{
	Use:   "restore [backup-id]",
	Short: "Restore settings from a backup",
	Long: `Restore Claude configuration settings from a backup.
You must specify the backup ID from the 'backup list' command.`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupRestore,
}

func newBackupRestoreCmd() *cobra.Command {
	return backupRestoreCmd
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	quiet, _ := cmd.Flags().GetBool("quiet")

	configManager := config.NewManager()

	if !quiet {
		fmt.Printf("Restoring backup %s... ", backupID)
	}

	err := configManager.RestoreBackup(backupID)
	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	if !quiet {
		fmt.Printf("done\n")
		fmt.Println("Settings restored successfully")
	}

	return nil
}

// backupDeleteCmd represents the backup delete command
var backupDeleteCmd = &cobra.Command{
	Use:   "delete [backup-id]",
	Short: "Delete a backup",
	Long: `Delete a specific backup. Use the backup ID from 'backup list' command.`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupDelete,
}

func newBackupDeleteCmd() *cobra.Command {
	return backupDeleteCmd
}

func runBackupDelete(cmd *cobra.Command, args []string) error {
	backupID := args[0]
	quiet, _ := cmd.Flags().GetBool("quiet")

	configManager := config.NewManager()
	backupManager := config.NewBackupManager(configManager)

	if !quiet {
		fmt.Printf("Deleting backup %s... ", backupID)
	}

	err := backupManager.DeleteBackup(backupID)
	if err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	if !quiet {
		fmt.Printf("done\n")
	}

	return nil
}

// backupPruneCmd represents the backup prune command
var backupPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Delete old backups",
	Long: `Delete backups older than the specified duration.
Examples:
  cflip backup prune --older-than 7d   # Delete backups older than 7 days
  cflip backup prune --older-than 24h  # Delete backups older than 24 hours`,
	RunE: runBackupPrune,
}

func newBackupPruneCmd() *cobra.Command {
	backupPruneCmd.Flags().StringVarP(&backupOlderThan, "older-than", "o", "7d", "Delete backups older than this duration (e.g., 7d, 24h, 30m)")
	return backupPruneCmd
}

func runBackupPrune(cmd *cobra.Command, args []string) error {
	quiet, _ := cmd.Flags().GetBool("verbose")

	// Parse duration
	duration, err := time.ParseDuration(backupOlderThan)
	if err != nil {
		// Try common formats
		if strings.HasSuffix(backupOlderThan, "d") {
			days := strings.TrimSuffix(backupOlderThan, "d")
			if daysInt, err := strconv.Atoi(days); err == nil {
				duration = time.Duration(daysInt) * 24 * time.Hour
			}
		} else if strings.HasSuffix(backupOlderThan, "h") {
			hours := strings.TrimSuffix(backupOlderThan, "h")
			if hoursInt, err := strconv.Atoi(hours); err == nil {
				duration = time.Duration(hoursInt) * time.Hour
			}
		} else if strings.HasSuffix(backupOlderThan, "m") {
			minutes := strings.TrimSuffix(backupOlderThan, "m")
			if minutesInt, err := strconv.Atoi(minutes); err == nil {
				duration = time.Duration(minutesInt) * time.Minute
			}
		}

		if duration == 0 {
			return fmt.Errorf("invalid duration format. Use formats like: 7d, 24h, 30m")
		}
	}

	configManager := config.NewManager()
	backupManager := config.NewBackupManager(configManager)

	err = backupManager.PruneBackups(duration)
	if err != nil {
		return fmt.Errorf("failed to prune backups: %w", err)
	}

	if !quiet {
		fmt.Printf("Pruned backups older than %s\n", backupOlderThan)
	}

	return nil
}