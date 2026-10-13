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

// GenerateAndOpen fetches history graph visits, injects them into the template, and opens the browser.
func GenerateAndOpen(limit int, customPath string) error {
	visits, err := history.GetGraphVisits(limit)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(visits)
	if err != nil {
		return fmt.Errorf("failed to marshal visits to json: %w", err)
	}
	safeJSON := strings.ReplaceAll(string(jsonData), "</", "<\\/")
	htmlContent := strings.ReplaceAll(htmlTemplate, "__DATA__", safeJSON)

	// Determine output path: default to ~/.footprints/templates/graph.html
	outputPath := customPath
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
