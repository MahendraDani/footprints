// /*
// Copyright © 2026 NAME HERE <EMAIL ADDRESS>
// */
// package cmd

// import (
// 	"fmt"
// 	"os"

// 	"github.com/MahendraDani/footprints/internal/update"
// 	"github.com/spf13/cobra"
// )

// // rootCmd represents the base command when called without any subcommands
// var rootCmd = &cobra.Command{
// 	Use:   "footprints",
// 	Short: "A brief description of your application",
// 	Long: `A longer description that spans multiple lines and likely contains
// examples and usage of using your application. For example:

// Cobra is a CLI library for Go that empowers applications.
// This application is a tool to generate the needed files
// to quickly create a Cobra application.`,
// 	// Uncomment the following line if your bare application
// 	// has an action associated with it:
// 	// Run: func(cmd *cobra.Command, args []string) { },
// }

// // Execute adds all child commands to the root command and sets flags appropriately.
// // This is called by main.main(). It only needs to happen once to the rootCmd.
// func Execute() {
// 	notice := update.Start(version, os.Args[1:])
// 	err := rootCmd.Execute()
// 	if msg := notice(); msg != "" {
// 		fmt.Fprint(os.Stderr, msg)
// 	}
// 	if err != nil {
// 		os.Exit(1)
// 	}
// }

// func init() {
// 	// Here you will define your flags and configuration settings.
// 	// Cobra supports persistent flags, which, if defined here,
// 	// will be global for your application.

// 	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.footprints.yaml)")

// 	// Cobra also supports local flags, which will only run
// 	// when this action is called directly.
// 	rootCmd.Version = version
// 	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
// 	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
// }

package cmd

import (
	"fmt"
	"os"

	"github.com/MahendraDani/footprints/internal/update"
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
