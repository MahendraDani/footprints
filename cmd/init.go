package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/MahendraDani/footprints/internal/workspace"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize footprints workspace and cache Chrome history",
	Long:  `Sets up the local ~/.footprints directory and copies your Chrome history`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing footprints...")

		// 1. Create workspace directories
		dir, err := workspace.EnsureInitialized()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Workspace ready at:", dir)

		// 2. Get live Chrome history source path
		srcPath, err := workspace.ChromeSourcePath()
		if err != nil {
			fmt.Println("Error locating Chrome history:", err)
			os.Exit(1)
		}

		// 3. Get local destination path
		destPath, err := workspace.HistoryPath()
		if err != nil {
			fmt.Println("Error determining destination path:", err)
			os.Exit(1)
		}

		// 4. Copy the file
		err = copyFile(srcPath, destPath)
		if err != nil {
			fmt.Println("Error copying history file:", err)
			fmt.Println("Tip: Make sure Google Chrome is fully closed before running 'fp init'.")
			os.Exit(1)
		}

		fmt.Println("Successfully cached Chrome history to:", destPath)
	},
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

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

func init() {
	rootCmd.AddCommand(initCmd)
}
