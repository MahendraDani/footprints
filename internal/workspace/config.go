package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds the user's saved state (~/.footprints/config.json).
type Config struct {
	Schema        int    `json:"schema"`                   // layout version, see SchemaVersion
	ChromeProfile string `json:"chrome_profile,omitempty"` // the current profile: `graph` reads its cached copy
}

// ConfigPath returns ~/.footprints/config.json
func ConfigPath() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "config.json"), nil
}

// LoadConfig returns the saved config, or an empty one if there is none.
func LoadConfig() Config {
	var c Config
	if p, err := ConfigPath(); err == nil {
		if b, err := os.ReadFile(p); err == nil {
			_ = json.Unmarshal(b, &c)
		}
	}
	return c
}

// SaveConfig writes the config, stamped with the current schema version.
func SaveConfig(c Config) error {
	c.Schema = SchemaVersion
	return writeConfig(c)
}

// writeConfig writes the config exactly as given (migrations use it to stamp intermediate versions).
func writeConfig(c Config) error {
	if _, err := EnsureRoot(); err != nil {
		return err
	}
	p, err := ConfigPath()
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
