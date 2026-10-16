package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func fakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows
	return filepath.Join(home, WorkspaceDirName)
}

func TestMigrateRemovesOnlyLegacyLayout(t *testing.T) {
	root := fakeHome(t)
	for _, d := range []string{"user", "templates", "bin", "chrome/Default"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	note, err := Migrate()
	if err != nil {
		t.Fatal(err)
	}
	if note == "" {
		t.Error("expected a note about the removed cache")
	}
	for _, d := range []string{"user", "templates"} {
		if _, err := os.Stat(filepath.Join(root, d)); !os.IsNotExist(err) {
			t.Errorf("%s should have been removed", d)
		}
	}
	for _, d := range []string{"bin", "chrome/Default"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(d))); err != nil {
			t.Errorf("%s must survive: %v", d, err)
		}
	}
	if got := LoadConfig().Schema; got != SchemaVersion {
		t.Errorf("schema = %d, want %d", got, SchemaVersion)
	}
	if note, _ := Migrate(); note != "" {
		t.Error("second run should do nothing")
	}
}

func TestMigrateFreshMachineCreatesNothing(t *testing.T) {
	root := fakeHome(t)
	if note, err := Migrate(); err != nil || note != "" {
		t.Fatalf("note=%q err=%v", note, err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Error("Migrate must not create ~/.footprints on a fresh machine")
	}
}
