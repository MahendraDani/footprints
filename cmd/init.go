package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const footprintsDirName = ".footprints"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init footprints (fp)",
	Long:  `Initializes the footprints workspace and caches Chrome history locally under the user directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing footprints...")

		// 1. Create the .footprints/user directory
		dir, err := ensureUserDir()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("User directory created at:", dir)

		// 2. Locate Chrome History file
		historyPath, err := getChromeHistoryPath()
		if err != nil {
			fmt.Println("Error locating Chrome history:", err)
			os.Exit(1)
		}

		// 3. Copy History file to .footprints/user/History
		destPath := filepath.Join(dir, "History")
		err = copyFile(historyPath, destPath)
		if err != nil {
			fmt.Println("Error copying history file:", err)
			fmt.Println("Tip: Make sure Google Chrome is closed before running init.")
			os.Exit(1)
		}

		fmt.Println("Chrome history successfully copied to:", destPath)
	},
}

func ensureUserDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Creates .footprints/user nested path safely
	dirPath := filepath.Join(homeDir, footprintsDirName, "user")
	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create user directory: %w", err)
	}

	return dirPath, nil
}

func getChromeHistoryPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Google", "Chrome", "Default", "History"), nil
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "History"), nil
	case "linux":
		return filepath.Join(homeDir, ".config", "google-chrome", "Default", "History"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Stream the data from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Ensure contents are written to disk
	return destFile.Sync()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
