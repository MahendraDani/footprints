package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const WorkspaceDirName = ".footprints"
const versionFileName = "version.json"

// RootDir returns the root workspace path (~/.footprints)
func RootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, WorkspaceDirName), nil
}

// UserDir returns the user data workspace path (~/.footprints/user)
func UserDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "user"), nil
}

// HistoryPath returns the path of the copied SQLite history database (~/.footprints/user/History)
func HistoryPath() (string, error) {
	userDir, err := UserDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, "History"), nil
}

// EnsureInitialized creates ~/.footprints/user and returns its path
func EnsureInitialized() (string, error) {
	userDir, err := UserDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace directory: %w", err)
	}
	return userDir, nil
}

// IsInitialized checks if the workspace has already been initialized (History file exists)
func IsInitialized() bool {
	hp, err := HistoryPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(hp)
	return err == nil
}

// ChromeSourcePath returns the OS-specific default path to Chrome's live History file
func ChromeSourcePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "Default", "History"), nil
	case "windows":
		return filepath.Join(home, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "History"), nil
	case "linux":
		return filepath.Join(home, ".config", "google-chrome", "Default", "History"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// EnsureRoot creates ~/.footprints (if missing) and returns its path.
func EnsureRoot() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0755); err != nil {
		return "", fmt.Errorf("failed to create workspace root: %w", err)
	}
	return root, nil
}

// VersionPath returns the update-check cache path (~/.footprints/version.json)
func VersionPath() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, versionFileName), nil
}

// BinDir returns where install.sh places the executable (~/.footprints/bin)
func BinDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "bin"), nil
}
