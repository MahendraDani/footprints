package cmd

import (
	"fmt"

	"github.com/MahendraDani/footprints/internal/graph"
	"github.com/spf13/cobra"
)

var (
	graphLimit int
	outputPath string
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate and open an interactive visit graph from Chrome history",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Analyzing history and generating visit graph (limit: %d)...\n", graphLimit)

		if err := graph.GenerateAndOpen(graphLimit, outputPath); err != nil {
			return err
		}

		fmt.Println("Graph successfully generated and opened in your browser!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().IntVarP(&graphLimit, "limit", "l", 3000, "Maximum number of recent visits to include")
	graphCmd.Flags().StringVarP(&outputPath, "output", "o", "graph.html", "Path to output HTML file")
}
