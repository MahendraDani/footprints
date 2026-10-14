package cmd

import (
	"fmt"

	"github.com/MahendraDani/footprints/internal/graph"
	"github.com/spf13/cobra"
)

var (
	graphLimit    int
	outputPath    string
	graphDirected bool
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate and open an interactive visit graph from Chrome history",
	RunE: func(cmd *cobra.Command, args []string) error {
		mode := "undirected"
		if graphDirected {
			mode = "directed"
		}
		fmt.Printf("Analyzing history and generating %s visit graph (limit: %d)...\n", mode, graphLimit)

		if err := graph.GenerateAndOpen(graph.Options{
			Limit:      graphLimit,
			OutputPath: outputPath,
			Directed:   graphDirected,
		}); err != nil {
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
	graphCmd.Flags().BoolVarP(&graphDirected, "directed", "d", false, "Start with a directed graph (arrows show where you navigated from and to)")
}
