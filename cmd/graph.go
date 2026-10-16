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
	graphFrom     string
	graphTo       string
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Generate and open an interactive visit graph from Chrome history",
	Long: `Builds an interactive graph from your cached Chrome history and opens it in your browser.

Each node is one page visit, and each link connects a visit to the one that led to it.
In the page you can filter by how you arrived (link, typed, bookmark, ...), pick a time
range, switch between directed and undirected views, search, and click a node for details.

Use --from-date and --to-date to load only a period. Dates are in your local time:
YYYY-MM-DD or "YYYY-MM-DD HH:MM". A plain --to-date includes that whole day.

Run "footprints init" first to cache your Chrome history.`,
	Example: `  footprints graph
  footprints graph --directed
  footprints graph --from-date 2026-10-01 --to-date 2026-10-15
  footprints graph --from-date "2026-10-14 09:00"
  footprints graph --limit 5000`,
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := graph.Options{
			Limit:      graphLimit,
			OutputPath: outputPath,
			Directed:   graphDirected,
		}
		if graphFrom != "" {
			t, err := graph.ParseDateBound(graphFrom, false)
			if err != nil {
				return fmt.Errorf("--from-date: %w", err)
			}
			opts.From = &t
		}
		if graphTo != "" {
			t, err := graph.ParseDateBound(graphTo, true)
			if err != nil {
				return fmt.Errorf("--to-date: %w", err)
			}
			opts.To = &t
		}
		if opts.From != nil && opts.To != nil && !opts.From.Before(*opts.To) {
			return fmt.Errorf("--from-date must be earlier than --to-date")
		}

		if graphLimit > 0 {
			fmt.Printf("Analyzing history and generating visit graph (latest %d visits)...\n", graphLimit)
		} else {
			fmt.Println("Analyzing history and generating visit graph...")
		}
		if err := graph.GenerateAndOpen(opts); err != nil {
			return err
		}
		fmt.Println("Graph successfully generated and opened in your browser!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(graphCmd)
	graphCmd.Flags().IntVarP(&graphLimit, "limit", "l", 3000, "Maximum number of recent visits to load (0 = all)")
	graphCmd.Flags().StringVarP(&outputPath, "output", "o", "graph.html", "Path to output HTML file")
	graphCmd.Flags().BoolVarP(&graphDirected, "directed", "d", false, "Start with a directed graph (arrows show where you navigated from and to)")
	graphCmd.Flags().StringVar(&graphFrom, "from-date", "", "Only include visits on or after this date (YYYY-MM-DD)")
	graphCmd.Flags().StringVar(&graphTo, "to-date", "", "Only include visits on or before this date (YYYY-MM-DD)")
}
