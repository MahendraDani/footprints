package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
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

func TestUserDir(t *testing.T) {
	tmpDir := mockHome(t)

	userDir, err := UserDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(tmpDir, ".footprints", "user")
	if userDir != expected {
		t.Errorf("expected user dir %s, got %s", expected, userDir)
	}
}

func TestEnsureInitialized(t *testing.T) {
	mockHome(t)

	// Initially, workspace should not be initialized
	if IsInitialized() {
		t.Errorf("expected IsInitialized() to be false before initialization")
	}

	// Run EnsureInitialized
	dir, err := EnsureInitialized()
	if err != nil {
		t.Fatalf("unexpected error during initialization: %v", err)
	}

	// Verify directory was actually created on disk
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", dir)
	}
}

func TestIsInitializedWithHistoryFile(t *testing.T) {
	mockHome(t)

	// Create the workspace structure first
	userDir, err := EnsureInitialized()
	if err != nil {
		t.Fatalf("failed to init workspace: %v", err)
	}

	// Still false because 'History' file doesn't exist yet
	if IsInitialized() {
		t.Errorf("expected IsInitialized() to be false when History file is missing")
	}

	// Simulate copying the History file
	historyPath := filepath.Join(userDir, "History")
	err = os.WriteFile(historyPath, []byte("fake sqlite db bytes"), 0644)
	if err != nil {
		t.Fatalf("failed to write mock history file: %v", err)
	}

	// Now should return true
	if !IsInitialized() {
		t.Errorf("expected IsInitialized() to be true after History file is present")
	}
}
