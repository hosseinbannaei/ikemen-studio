package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEngineSwitchAndRollback(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-switch-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	projectDir := filepath.Join(tmpDir, "game")
	_ = os.MkdirAll(filepath.Join(projectDir, "external"), 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "external", "v1.txt"), []byte("v1"), 0644)

	manifest := &ProjectManifest{
		Name:    "Test Game",
		Version: "1.0.0",
		Engine: EngineConfig{
			Version: "v0.99.0",
			Channel: "stable",
		},
		Path: projectDir,
	}
	if err := SaveManifest(projectDir, manifest); err != nil {
		t.Fatal(err)
	}

	// New engine v2
	engineV2 := filepath.Join(tmpDir, "engine_v2")
	_ = os.MkdirAll(filepath.Join(engineV2, "external"), 0755)
	_ = os.WriteFile(filepath.Join(engineV2, "external", "v2.txt"), []byte("v2"), 0644)

	// 1. Switch Engine to v2
	if err := SwitchEngine(projectDir, engineV2, "nightly"); err != nil {
		t.Fatalf("SwitchEngine failed: %v", err)
	}

	// Check that manifest updated
	updated, err := LoadManifest(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Engine.Version != "nightly" {
		t.Errorf("expected engine version 'nightly', got %s", updated.Engine.Version)
	}

	// Check that external has v2.txt and not v1.txt
	if _, err := os.Stat(filepath.Join(projectDir, "external", "v2.txt")); err != nil {
		t.Errorf("expected v2.txt to exist")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "external", "v1.txt")); err == nil {
		t.Errorf("expected old v1.txt to be removed")
	}

	// Check backup was created
	backups, err := GetEngineBackups(projectDir)
	if err != nil {
		t.Fatalf("GetEngineBackups failed: %v", err)
	}
	if len(backups) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(backups))
	}
	if backups[0].Version != "v0.99.0" {
		t.Errorf("expected backup version 'v0.99.0', got %s", backups[0].Version)
	}

	// 2. Rollback Engine to v1
	if err := RollbackEngine(projectDir, backups[0].ID); err != nil {
		t.Fatalf("RollbackEngine failed: %v", err)
	}

	// Check that manifest rolled back
	rolledBack, err := LoadManifest(projectDir)
	if err != nil {
		t.Fatal(err)
	}
	if rolledBack.Engine.Version != "v0.99.0" {
		t.Errorf("expected engine version 'v0.99.0' after rollback, got %s", rolledBack.Engine.Version)
	}

	// Check that v1.txt is restored
	if _, err := os.Stat(filepath.Join(projectDir, "external", "v1.txt")); err != nil {
		t.Errorf("expected v1.txt to be restored after rollback")
	}
}
