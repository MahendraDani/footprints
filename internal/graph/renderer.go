package graph

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/MahendraDani/footprints/internal/history"
	"github.com/MahendraDani/footprints/internal/workspace"
)

//go:embed index.html
var htmlTemplate string

// Options controls graph generation.
type Options struct {
	Limit      int
	OutputPath string
	Directed   bool
	From       *time.Time // inclusive
	To         *time.Time // exclusive
}

// pageConfig is injected into the template as JSON (read by the page as CFG).
type pageConfig struct {
	Directed  bool   `json:"directed"`
	From      *int64 `json:"from"` // unix ms
	To        *int64 `json:"to"`   // unix ms, exclusive
	Limit     int    `json:"limit"`
	Truncated bool   `json:"truncated"`
}

// ParseDateBound parses a --from-date / --to-date value in local time.
// Accepts YYYY-MM-DD, "YYYY-MM-DD HH:MM" and YYYY-MM-DDTHH:MM.
// For an end bound the result is exclusive: a plain date covers that whole day.
func ParseDateBound(s string, end bool) (time.Time, error) {
	s = strings.TrimSpace(s)
	layouts := []struct {
		layout   string
		dateOnly bool
	}{
		{"2006-01-02", true},
		{"2006-01-02 15:04", false},
		{"2006-01-02T15:04", false},
	}
	for _, l := range layouts {
		t, err := time.ParseInLocation(l.layout, s, time.Local)
		if err != nil {
			continue
		}
		if end {
			if l.dateOnly {
				t = t.AddDate(0, 0, 1)
			} else {
				t = t.Add(time.Minute)
			}
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date %q: use YYYY-MM-DD or \"YYYY-MM-DD HH:MM\"", s)
}

// GenerateAndOpen fetches visits, injects them and the config into the template, and opens the browser.
func GenerateAndOpen(opts Options) error {
	visits, truncated, err := history.GetGraphVisits(opts.Limit, opts.From, opts.To)
	if err != nil {
		return err
	}
	if len(visits) == 0 && (opts.From != nil || opts.To != nil) {
		return fmt.Errorf("no visits found in that date range")
	}
	fmt.Printf("Loaded %d visits\n", len(visits))
	if truncated {
		fmt.Printf("Note: reached the limit of %d, so older visits were not loaded. Use --limit to load more.\n", opts.Limit)
	}
	if opts.Limit <= 0 && len(visits) > 30000 {
		fmt.Println("Note: that's a lot of visits, so the page may be slow to open. Use --from-date or --limit to load less.")
	}

	cfg := pageConfig{Directed: opts.Directed, Limit: opts.Limit, Truncated: truncated}
	if opts.From != nil {
		ms := opts.From.UnixMilli()
		cfg.From = &ms
	} else {
		//  if --from-date is not provided as options, show the oldest visit's time
		ms := visits[len(visits)-1].VisitTS
		cfg.From = &ms
	}
	if opts.To != nil {
		ms := opts.To.UnixMilli()
		cfg.To = &ms
	} else {
		// if --to-date is not provided as options, show the latest visits's time
		ms := visits[0].VisitTS
		cfg.To = &ms
	}

	jsonData, err := json.Marshal(visits)
	if err != nil {
		return fmt.Errorf("failed to marshal visits to json: %w", err)
	}
	cfgData, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config to json: %w", err)
	}

	safeJSON := strings.ReplaceAll(string(jsonData), "</", "<\\/")

	// One pass over the template, so text inside your history that happens
	// to look like a placeholder can never be substituted a second time.
	htmlContent := strings.NewReplacer(
		"__DATA__", safeJSON,
		"__CONFIG__", string(cfgData),
	).Replace(htmlTemplate)

	// Determine output path: default to ~/.footprints/templates/graph.html
	outputPath := opts.OutputPath
	if outputPath == "" || outputPath == "graph.html" {
		p, err := workspace.GraphPath() // ~/.footprints/chrome/<profile>/templates/graph.html
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return fmt.Errorf("failed to create templates directory: %w", err)
		}
		outputPath = p
	}

	absPath, err := filepath.Abs(outputPath)
	if err != nil {
		absPath = outputPath
	}

	if err := os.WriteFile(absPath, []byte(htmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write graph html: %w", err)
	}

	fmt.Printf("Saving graph to: %s\n", absPath)
	return openBrowser(absPath)
}

func openBrowser(filePath string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", filePath)
	default:
		cmd = exec.Command("xdg-open", filePath)
	}
	return cmd.Start()
}
