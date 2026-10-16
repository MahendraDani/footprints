package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// mockHome isolates tests by pointing HOME/USERPROFILE to a temporary directory
func mockHome(t *testing.T) string {
	tmpDir := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", tmpDir)
	} else {
		t.Setenv("HOME", tmpDir)
	}
	return tmpDir
}

// writeHistory creates a fake cached History file for a profile.
func writeHistory(t *testing.T, profile string) string {
	t.Helper()
	if _, err := EnsureProfileDir(profile); err != nil {
		t.Fatalf("failed to create profile dir: %v", err)
	}
	p, err := ProfileHistoryPath(profile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := os.WriteFile(p, []byte("fake sqlite db bytes"), 0o644); err != nil {
		t.Fatalf("failed to write mock history file: %v", err)
	}
	return p
}

func TestRootDir(t *testing.T) {
	tmpDir := mockHome(t)

	root, err := RootDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmpDir, ".footprints")
	if root != expected {
		t.Errorf("expected root dir %s, got %s", expected, root)
	}
}

func TestProfilePaths(t *testing.T) {
	tmpDir := mockHome(t)
	base := filepath.Join(tmpDir, ".footprints", "chrome", "Profile 1")

	dir, err := ProfileDir("Profile 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dir != base {
		t.Errorf("ProfileDir = %s, want %s", dir, base)
	}

	hp, err := ProfileHistoryPath("Profile 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(base, "History"); hp != want {
		t.Errorf("ProfileHistoryPath = %s, want %s", hp, want)
	}
}

func TestProfileDirRejectsUnsafeNames(t *testing.T) {
	mockHome(t)
	for _, name := range []string{"", ".", "..", "../escape", "a/b", `a\b`} {
		if _, err := ProfileDir(name); err == nil {
			t.Errorf("ProfileDir(%q) should fail", name)
		}
		if _, err := EnsureProfileDir(name); err == nil {
			t.Errorf("EnsureProfileDir(%q) should fail", name)
		}
	}
}

func TestEnsureProfileDir(t *testing.T) {
	mockHome(t)

	dir, err := EnsureProfileDir("Default")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", dir)
	}
}

func TestCurrentProfile(t *testing.T) {
	mockHome(t)

	if _, err := CurrentProfile(); !errors.Is(err, ErrNoProfile) {
		t.Errorf("expected ErrNoProfile before init, got %v", err)
	}

	useProfile(t, "Profile 1")
	got, err := CurrentProfile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Profile 1" {
		t.Errorf("CurrentProfile = %q, want %q", got, "Profile 1")
	}
}

func TestCurrentProfileRejectsTamperedConfig(t *testing.T) {
	mockHome(t)

	// A hand-edited config must not be able to point outside ~/.footprints/chrome
	useProfile(t, "../../etc")
	_, err := CurrentProfile()
	if err == nil || errors.Is(err, ErrNoProfile) {
		t.Errorf("expected an invalid-profile error, got %v", err)
	}
	if IsInitialized() {
		t.Error("IsInitialized must be false for an invalid profile name")
	}
}

func TestSaveConfigStampsSchema(t *testing.T) {
	mockHome(t)

	useProfile(t, "Default")
	cfg := LoadConfig()
	if cfg.Schema != SchemaVersion {
		t.Errorf("Schema = %d, want %d", cfg.Schema, SchemaVersion)
	}
	if cfg.Chrome.Current != "Default" {
		t.Errorf("ChromeProfile = %q, want %q", cfg.Chrome.Current, "Default")
	}
}

func TestHistoryAndGraphPathFollowCurrentProfile(t *testing.T) {
	tmpDir := mockHome(t)

	if _, err := HistoryPath(); !errors.Is(err, ErrNoProfile) {
		t.Errorf("HistoryPath before init: expected ErrNoProfile, got %v", err)
	}
	if _, err := GraphPath(); !errors.Is(err, ErrNoProfile) {
		t.Errorf("GraphPath before init: expected ErrNoProfile, got %v", err)
	}

	useProfile(t, "Profile 2")
	base := filepath.Join(tmpDir, ".footprints", "chrome", "Profile 2")

	hp, err := HistoryPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(base, "History"); hp != want {
		t.Errorf("HistoryPath = %s, want %s", hp, want)
	}

	gp, err := GraphPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := filepath.Join(base, "templates", "graph.html"); gp != want {
		t.Errorf("GraphPath = %s, want %s", gp, want)
	}
}

func TestIsInitialized(t *testing.T) {
	mockHome(t)

	// No profile chosen yet
	if IsInitialized() {
		t.Errorf("expected IsInitialized() to be false before init")
	}

	// Profile chosen, but nothing cached yet
	useProfile(t, "Default")
	if IsInitialized() {
		t.Errorf("expected IsInitialized() to be false when History file is missing")
	}

	// History cached for the current profile
	writeHistory(t, "Default")
	if !IsInitialized() {
		t.Errorf("expected IsInitialized() to be true after History file is present")
	}
}

func TestSwitchingProfilesKeepsBothCopies(t *testing.T) {
	mockHome(t)

	a := writeHistory(t, "Default")
	b := writeHistory(t, "Profile 1")

	useProfile(t, "Default")
	if hp, _ := HistoryPath(); hp != a {
		t.Errorf("HistoryPath = %s, want %s", hp, a)
	}

	// Switching the dispatcher must not touch the previous profile's cached copy
	useProfile(t, "Profile 1")
	if hp, _ := HistoryPath(); hp != b {
		t.Errorf("HistoryPath = %s, want %s", hp, b)
	}
	if _, err := os.Stat(a); err != nil {
		t.Errorf("previous profile's copy should still exist: %v", err)
	}
}

// useProfile sets the current profile in config.json.
func useProfile(t *testing.T, name string) {
	t.Helper()
	if err := SaveConfig(Config{Chrome: ChromeConfig{Current: name}}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}
}

func TestRecordSync(t *testing.T) {
	mockHome(t)
	t1 := time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC)
	t2 := t1.Add(48 * time.Hour)

	if err := RecordSync("Default", "Person 1", "a@x.com", t1); err != nil {
		t.Fatal(err)
	}
	if err := RecordSync("Profile 1", "Work", "w@x.com", t2); err != nil {
		t.Fatal(err)
	}

	cfg := LoadConfig()
	if cfg.Chrome.Current != "Profile 1" {
		t.Errorf("Current = %q, want Profile 1", cfg.Chrome.Current)
	}
	d, ok := cfg.Chrome.Profiles["Default"]
	if !ok || d.Name != "Person 1" || d.Username != "a@x.com" || !d.LastSynced.Equal(t1) {
		t.Errorf("Default entry wrong: %+v", d)
	}

	// Re-syncing Default moves "current" back and updates only its own timestamp
	t3 := t2.Add(time.Hour)
	if err := RecordSync("Default", "Person 1", "a@x.com", t3); err != nil {
		t.Fatal(err)
	}
	cfg = LoadConfig()
	if cfg.Chrome.Current != "Default" {
		t.Errorf("Current = %q, want Default", cfg.Chrome.Current)
	}
	if !cfg.Chrome.Profiles["Default"].LastSynced.Equal(t3) {
		t.Error("Default's last_synced should have been refreshed")
	}
	if !cfg.Chrome.Profiles["Profile 1"].LastSynced.Equal(t2) {
		t.Error("Profile 1's last_synced must not change")
	}
}

func TestRecordSyncRejectsUnsafeName(t *testing.T) {
	mockHome(t)
	if err := RecordSync("../escape", "x", "", time.Now()); err == nil {
		t.Error("expected an error for an unsafe profile name")
	}
	if p := LoadConfig().Chrome.Current; p != "" {
		t.Errorf("config must not be modified on error, Current = %q", p)
	}
}
