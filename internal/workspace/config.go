package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ChromeProfileInfo is what footprints remembers about one cached Chrome profile.
type ChromeProfileInfo struct {
	Name       string    `json:"name,omitempty"`     // display name from Chrome, e.g. "Work"
	Username   string    `json:"username,omitempty"` // the profile's Google account, if signed in
	LastSynced time.Time `json:"last_synced,omitzero"`
}

// ChromeConfig tracks which profile is in use and every profile that has been cached.
type ChromeConfig struct {
	Current  string                       `json:"current,omitempty"`  // profile folder name, e.g. "Profile 1"
	Profiles map[string]ChromeProfileInfo `json:"profiles,omitempty"` // keyed by profile folder name
}

// Config holds the user's saved state (~/.footprints/config.json).
type Config struct {
	Schema  int          `json:"schema"`            // layout version, see SchemaVersion
	Version string       `json:"version,omitempty"` // last release that used this workspace (informational only)
	Chrome  ChromeConfig `json:"chrome"`
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

// RecordSync marks a profile as the current one and records when its history was copied.
// Call it after the History file has been copied successfully.
func RecordSync(dir, name, username string, at time.Time) error {
	if err := validProfile(dir); err != nil {
		return err
	}
	cfg := LoadConfig()
	cfg.Chrome.Current = dir
	if cfg.Chrome.Profiles == nil {
		cfg.Chrome.Profiles = map[string]ChromeProfileInfo{}
	}
	cfg.Chrome.Profiles[dir] = ChromeProfileInfo{
		Name:       name,
		Username:   username,
		LastSynced: at.UTC().Truncate(time.Second),
	}
	return SaveConfig(cfg)
}
