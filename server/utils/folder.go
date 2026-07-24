package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// WorksDirectory name for server data
	WorksDirectory = "Works"
)

// GetServerFolderPath returns the server folder path
// serverID: server database ID
// Returns: absolute path to the server folder
func GetServerFolderPath(serverID uint) string {
	// Get current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		// If failed, use current directory
		currentDir = "."
	}

	// Build path: currentDir/Works/serverID
	folderPath := filepath.Join(currentDir, WorksDirectory, fmt.Sprintf("%d", serverID))
	return folderPath
}

// CreateServerFolder creates the server folder
// serverID: server database ID
// Returns: created folder path and error
func CreateServerFolder(serverID uint) (string, error) {
	folderPath := GetServerFolderPath(serverID)

	// Create folder (including parent directories)
	err := os.MkdirAll(folderPath, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create server folder: %v", err)
	}

	return folderPath, nil
}

// FolderExists checks if a folder exists
// folderPath: path to check
// Returns: whether the folder exists
func FolderExists(folderPath string) bool {
	info, err := os.Stat(folderPath)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// RemoveServerFolder removes the server folder
// serverID: server database ID
// Returns: error
func RemoveServerFolder(serverID uint) error {
	folderPath := GetServerFolderPath(serverID)

	// Check if folder exists
	if !FolderExists(folderPath) {
		return nil // Folder doesn't exist, consider it removed
	}

	// Delete folder and its contents
	err := os.RemoveAll(folderPath)
	if err != nil {
		return fmt.Errorf("Failed to delete server folder: %v", err)
	}

	return nil
}

// GetFolderSize returns the folder size in bytes
// folderPath: path to the folder
// Returns: folder size and error
func GetFolderSize(folderPath string) (int64, error) {
	var size int64

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// EnsureWorksDirectory ensures the Works directory exists
// Returns: Works directory path and error
func EnsureWorksDirectory() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		currentDir = "."
	}

	worksPath := filepath.Join(currentDir, WorksDirectory)

	// Create Works directory
	err = os.MkdirAll(worksPath, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create Works directory: %v", err)
	}

	return worksPath, nil
}
