package utils

import (
	"os"
)

// RenameFile renames a file from oldPath to newPath
func RenameFile(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

// RemoveFile removes a file at the specified path
func RemoveFile(path string) error {
	return os.Remove(path)
}

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}