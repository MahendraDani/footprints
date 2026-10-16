package cmd

import (
	"fmt"
	"os"

	"github.com/MahendraDani/footprints/internal/update"
	"github.com/MahendraDani/footprints/internal/workspace"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "footprints",
	Short: "Explore your Chrome browsing history as an interactive graph",
	Long: `footprints turns your Chrome history into an interactive graph in your browser.

Every page you visit becomes a node, linked to the page that led you there,
so you can see how you moved from site to site. Node size shows time spent
on a page, and colors show how you got there (link, typed, bookmark, ...).

Everything stays on your machine: footprints reads a copy of your history
file and generates a local HTML page. Nothing is uploaded anywhere.`,
	Example: `  footprints init                  # copy Chrome history into ~/.footprints
  footprints graph                 # open the visit graph in your browser
  footprints graph --directed      # draw arrows from earlier to later visits
  footprints update                # upgrade to the latest release`,
	SilenceUsage: true, // don't dump the help text when a command fails at runtime

	// Upgrades ~/.footprints to the current layout before any real command runs.
	// It's idempotent and costs one small file read once the workspace is current.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		switch cmd.Name() {
		case "help", "completion", "update", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
			return
		}
		note, err := workspace.Migrate()
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning:", err) // non-fatal: commands still work on the new layout
		}
		if note != "" {
			fmt.Fprintln(os.Stderr, note)
		}
	},
}

func init() {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	rootCmd.CompletionOptions.HiddenDefaultCmd = true // keeps `completion` out of the command list
}

func Execute() {
	notice := update.Start(version, os.Args[1:])
	err := rootCmd.Execute()
	if msg := notice(); msg != "" {
		fmt.Fprint(os.Stderr, msg)
	}
	if err != nil {
		os.Exit(1)
	}
}
