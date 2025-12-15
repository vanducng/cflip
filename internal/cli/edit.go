package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	darwinOS      = "darwin"
	windowsOS     = "windows"
	editorDarwin  = "open"
	editorWindows = "notepad"
	editorDefault = "nano"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit the current Claude settings file",
	Long: `Edit the current Claude settings file in your default editor.
This opens the ~/.claude/settings.json file with the system's default text editor.`,
	RunE: runEdit,
}

func init() {
	editCmd.Flags().BoolP("settings", "s", false, "Edit settings file (default)")
	editCmd.Flags().BoolP("cflip", "c", false, "Edit cflip config file")
	editCmd.Flags().BoolP("snapshot", "p", false, "List and manage snapshots")
}

func runEdit(cmd *cobra.Command, args []string) error {
	editCflip, _ := cmd.Flags().GetBool("cflip")
	editSnapshot, _ := cmd.Flags().GetBool("snapshot")

	if editSnapshot {
		return manageSnapshots()
	}

	if editCflip {
		return editCflipConfig()
	}

	// Default: edit Claude settings
	homeDir, _ := os.UserHomeDir()
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")

	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return fmt.Errorf("settings file not found at %s", settingsPath)
	}

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try common editors based on OS
		switch runtime.GOOS {
		case darwinOS:
			editor = editorDarwin
		case windowsOS:
			editor = editorWindows
		default:
			editor = editorDefault
		}
	}

	// Launch editor with context
	ctx := context.Background()
	var execCmd *exec.Cmd
	if runtime.GOOS == darwinOS && editor == editorDarwin {
		// On macOS, use 'open' with text editor mode
		execCmd = exec.CommandContext(ctx, editor, "-t", settingsPath)
	} else {
		execCmd = exec.CommandContext(ctx, editor, settingsPath)
	}

	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	fmt.Printf("Settings file opened: %s\n", settingsPath)
	return nil
}

func editCflipConfig() error {
	configPath := "internal/config/config.go"

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = editorDefault
	}

	// Launch editor with context
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	fmt.Printf("Config file opened: %s\n", configPath)
	return nil
}

func manageSnapshots() error {
	homeDir, _ := os.UserHomeDir()
	snapshotsDir := filepath.Join(homeDir, ".claude", "snapshots")

	// List snapshots
	snapshots, err := ListSnapshots(snapshotsDir)
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		fmt.Println("No snapshots found")
		return nil
	}

	fmt.Println("Available snapshots:")
	for i, snapshot := range snapshots {
		// Extract provider and timestamp
		parts := snapshot
		if len(parts) > len("snapshot-") {
			info := parts[9:] // Remove "snapshot-" prefix
			if idx := findIndex(info, '-'); idx > 0 {
				provider := info[:idx]
				timestamp := info[idx+1 : len(info)-5] // Remove .json
				fmt.Printf("  %d) %s - %s\n", i+1, provider, timestamp)
			}
		}
	}

	fmt.Printf("\nSnapshots directory: %s\n", snapshotsDir)
	fmt.Println("Note: To restore a snapshot, manually copy the contents to ~/.claude/settings.json")

	return nil
}
