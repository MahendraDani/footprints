package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SchemaVersion is the current layout of ~/.footprints.
//
//	1 (no schema in config.json): user/History + templates/graph.html
//	2: chrome/<profile>/History + chrome/<profile>/templates/graph.html
const SchemaVersion = 2

type migration struct {
	to   int
	name string
	run  func(root string) (removed []string, err error)
}

// Append new migrations at the end. Never edit or reorder old ones: someone may still be upgrading from that version.
var migrations = []migration{
	{to: 2, name: "per-profile layout", run: toPerProfileLayout},
}

// toPerProfileLayout deletes the old single-profile cache. It's only a cache, so it's safe to
// rebuild with `footprints init`, and the old graph.html embedded browsing history anyway.
// It removes exactly these two names and nothing else (never ~/.footprints/*, which also holds bin/).
func toPerProfileLayout(root string) ([]string, error) {
	var removed []string
	for _, name := range []string{"user", "templates"} {
		p := filepath.Join(root, name)
		if _, err := os.Lstat(p); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return removed, err
		}
		if err := os.RemoveAll(p); err != nil { // on a symlink this removes the link, not its target
			return removed, err
		}
		removed = append(removed, "~/.footprints/"+name)
	}
	return removed, nil
}

// Migrate upgrades ~/.footprints to the current layout. It's cheap and idempotent, so it's safe to
// run before every command. It returns a note for the user when it removed something.
func Migrate() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return "", nil // fresh machine: nothing to migrate, and nothing to create yet
		}
		return "", err
	}

	cfg := LoadConfig()
	if cfg.Schema >= SchemaVersion {
		return "", nil
	}

	var removed []string
	for _, m := range migrations {
		if m.to <= cfg.Schema {
			continue
		}
		r, err := m.run(root)
		removed = append(removed, r...)
		if err != nil {
			// The schema isn't stamped, so the next run retries; finished steps are skipped
			return "", fmt.Errorf("workspace upgrade (%s) failed: %w", m.name, err)
		}
		cfg.Schema = m.to
		if err := writeConfig(cfg); err != nil {
			return "", fmt.Errorf("workspace upgrade (%s): couldn't save config: %w", m.name, err)
		}
	}
	if len(removed) == 0 {
		return "", nil
	}
	return fmt.Sprintf("footprints: upgraded the workspace layout and removed the old cache (%s).\n"+
		"Run `footprints init` to cache your history again.", strings.Join(removed, ", ")), nil
}

// RecordVersion notes which release last used this workspace. It's purely informational:
// migrations are driven by Schema and never by this. It does nothing for dev builds, and it
// waits until the workspace is on the current layout, so it can't mask a failed migration.
func RecordVersion(v string) error {
	if v == "" || v == "dev" {
		return nil
	}
	if !strings.HasPrefix(v, "v") {
		v = "v" + v // the build stamps "0.1.3"; tags look like "v0.1.3"
	}
	cfg := LoadConfig()
	if cfg.Schema != SchemaVersion || cfg.Version == v {
		return nil
	}
	cfg.Version = v
	return SaveConfig(cfg)
}
