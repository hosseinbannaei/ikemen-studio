package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGameConfigIO(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-config-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test default load
	cfg, err := LoadGameConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadGameConfig failed: %v", err)
	}
	if cfg["Width"] != "1280" {
		t.Errorf("expected default Width 1280, got %s", cfg["Width"])
	}

	// Test saving updates
	updates := map[string]string{
		"Width":        "1920",
		"Height":       "1080",
		"Fullscreen":   "1",
		"VolumeMaster": "100",
	}
	if err := SaveGameConfig(tmpDir, updates); err != nil {
		t.Fatalf("SaveGameConfig failed: %v", err)
	}

	// Reload and verify
	loaded, err := LoadGameConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadGameConfig after save failed: %v", err)
	}

	if loaded["Width"] != "1920" {
		t.Errorf("expected Width 1920, got %s", loaded["Width"])
	}
	if loaded["Height"] != "1080" {
		t.Errorf("expected Height 1080, got %s", loaded["Height"])
	}
	if loaded["Fullscreen"] != "1" {
		t.Errorf("expected Fullscreen 1, got %s", loaded["Fullscreen"])
	}

	// Check file existence
	if _, err := os.Stat(filepath.Join(tmpDir, "save", "config.ini")); err != nil {
		t.Errorf("save/config.ini was not created")
	}
}
