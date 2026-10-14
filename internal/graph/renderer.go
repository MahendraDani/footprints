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

	"github.com/MahendraDani/footprints/internal/history"
)

//go:embed index.html
var htmlTemplate string

// Options controls graph generation.
type Options struct {
	Limit      int
	OutputPath string
	Directed   bool
}

// pageConfig is injected into the template as JSON (read by the page as CFG).
type pageConfig struct {
	Directed bool `json:"directed"`
}

// GenerateAndOpen fetches visits, injects them and the config into the template, and opens the browser.
func GenerateAndOpen(opts Options) error {
	visits, err := history.GetGraphVisits(opts.Limit)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(visits)
	if err != nil {
		return fmt.Errorf("failed to marshal visits to json: %w", err)
	}
	cfgData, err := json.Marshal(pageConfig{Directed: opts.Directed})
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
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home dir: %w", err)
		}
		templatesDir := filepath.Join(home, ".footprints", "templates")
		if err := os.MkdirAll(templatesDir, 0755); err != nil {
			return fmt.Errorf("failed to create templates directory: %w", err)
		}
		outputPath = filepath.Join(templatesDir, "graph.html")
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
