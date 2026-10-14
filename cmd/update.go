package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/MahendraDani/footprints/internal/update"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update footprints to the latest release",
	Long: `Downloads and runs the latest install script, replacing the installed binary.

footprints also checks for new releases once a day and prints a notice when one
is available. Set FOOTPRINTS_NO_UPDATE_CHECK=1 to turn that check off.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("self-update isn't supported on Windows yet; download the latest build from https://github.com/%s/releases/latest", update.Repo)
		}

		fmt.Println("Updating footprints...")
		resp, err := http.Get(update.InstallURL)
		if err != nil {
			return fmt.Errorf("could not download installer: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("could not download installer: %s", resp.Status)
		}

		// Save to a file first (instead of piping into sh) so a failed or partial download never executes.
		f, err := os.CreateTemp("", "footprints-install-*.sh")
		if err != nil {
			return err
		}
		defer os.Remove(f.Name())
		if _, err := io.Copy(f, resp.Body); err != nil {
			return err
		}
		f.Close()

		c := exec.Command("sh", f.Name())
		c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(updateCmd) }
