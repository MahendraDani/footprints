package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	WorkspaceDirName = ".footprints"
	versionFileName  = "version.json"
)

// ErrNoProfile means no Chrome profile has been chosen yet.
var ErrNoProfile = errors.New("no Chrome profile chosen yet; run `footprints init`")

// RootDir returns the root workspace path (~/.footprints)
func RootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, WorkspaceDirName), nil
}

// EnsureRoot creates ~/.footprints (if missing) and returns its path.
func EnsureRoot() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("failed to create workspace root: %w", err)
	}
	return root, nil
}

// VersionPath returns the update-check cache path (~/.footprints/version.json)
func VersionPath() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, versionFileName), nil
}

// BinDir returns where older installs placed the executable (~/.footprints/bin)
func BinDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "bin"), nil
}

// TelemetryPath returns ~/.footprints/telemetry.json
func TelemetryPath() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "telemetry.json"), nil
}

// ChromeDir returns ~/.footprints/chrome, which holds one folder per cached Chrome profile.
func ChromeDir() (string, error) {
	root, err := RootDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "chrome"), nil
}

// validProfile rejects names that could escape ~/.footprints/chrome (e.g. from a hand-edited config.json).
func validProfile(name string) error {
	if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `/\`) {
		return fmt.Errorf("invalid Chrome profile name %q", name)
	}
	return nil
}

// ProfileDir returns ~/.footprints/chrome/<profile>
func ProfileDir(profile string) (string, error) {
	if err := validProfile(profile); err != nil {
		return "", err
	}
	base, err := ChromeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, profile), nil
}

// EnsureProfileDir creates ~/.footprints/chrome/<profile> and returns its path.
func EnsureProfileDir(profile string) (string, error) {
	dir, err := ProfileDir(profile)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create profile directory: %w", err)
	}
	return dir, nil
}

// ProfileHistoryPath returns ~/.footprints/chrome/<profile>/History
func ProfileHistoryPath(profile string) (string, error) {
	dir, err := ProfileDir(profile)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "History"), nil
}

// CurrentProfile returns the profile saved in config.json (the one `init` chose last).
func CurrentProfile() (string, error) {
	p := LoadConfig().Chrome.Current
	if p == "" {
		return "", ErrNoProfile
	}
	if err := validProfile(p); err != nil {
		return "", fmt.Errorf("config.json: %w", err)
	}
	return p, nil
}

// HistoryPath returns the cached History of the current profile.
func HistoryPath() (string, error) {
	p, err := CurrentProfile()
	if err != nil {
		return "", err
	}
	return ProfileHistoryPath(p)
}

// GraphPath returns ~/.footprints/chrome/<current profile>/templates/graph.html
func GraphPath() (string, error) {
	p, err := CurrentProfile()
	if err != nil {
		return "", err
	}
	dir, err := ProfileDir(p)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "templates", "graph.html"), nil
}

// IsInitialized reports whether the current profile has a cached History file.
func IsInitialized() bool {
	hp, err := HistoryPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(hp)
	return err == nil
}
