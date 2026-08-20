package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	defaults := DefaultSettings()
	if defaults.EnginesDir == "" {
		t.Errorf("expected non-empty default engines dir")
	}
	if defaults.Theme != "dark" {
		t.Errorf("expected dark theme, got %s", defaults.Theme)
	}
	if defaults.DefaultChannel != "stable" {
		t.Errorf("expected stable channel, got %s", defaults.DefaultChannel)
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override config dir using environment across platforms
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("APPDATA", tmpDir)

	s := &Settings{
		EnginesDir:     filepath.Join(tmpDir, "custom-engines"),
		Theme:          "light",
		RecentProjects: []string{"/path/to/project1", "/path/to/project2"},
		DefaultChannel: "nightly",
	}

	if err := SaveSettings(s); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	loaded, err := LoadSettings()
	if err != nil {
		t.Fatalf("failed to load settings: %v", err)
	}

	if loaded.EnginesDir != s.EnginesDir {
		t.Errorf("expected engines dir %s, got %s", s.EnginesDir, loaded.EnginesDir)
	}
	if loaded.Theme != "light" {
		t.Errorf("expected theme light, got %s", loaded.Theme)
	}
	if loaded.DefaultChannel != "nightly" {
		t.Errorf("expected nightly channel, got %s", loaded.DefaultChannel)
	}
	if len(loaded.RecentProjects) != 2 {
		t.Errorf("expected 2 recent projects, got %d", len(loaded.RecentProjects))
	}

	// Test RemoveRecentProject
	if err := RemoveRecentProject("/path/to/project1"); err != nil {
		t.Fatalf("failed to remove recent project: %v", err)
	}

	afterRemove, err := LoadSettings()
	if err != nil {
		t.Fatalf("failed to load settings after removal: %v", err)
	}
	if len(afterRemove.RecentProjects) != 1 || afterRemove.RecentProjects[0] != "/path/to/project2" {
		t.Errorf("expected 1 recent project /path/to/project2, got %+v", afterRemove.RecentProjects)
	}
}

