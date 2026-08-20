package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectManifestValidation(t *testing.T) {
	manifest := &ProjectManifest{
		Name:    "",
		Version: "0.1.0",
		Engine: EngineConfig{
			Version: "v0.99.0",
			Channel: "stable",
		},
	}

	if err := manifest.Validate(); err == nil {
		t.Errorf("expected error for empty name, got nil")
	}

	manifest.Name = "My Fighter"
	if err := manifest.Validate(); err != nil {
		t.Errorf("expected valid manifest, got: %v", err)
	}
}

func TestProjectScaffoldingAndManifestIO(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-project-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake engine directory with mock data/system.def
	engineDir := filepath.Join(tmpDir, "mock-engine")
	if err := os.MkdirAll(filepath.Join(engineDir, "data"), 0755); err != nil {
		t.Fatal(err)
	}
	systemDefPath := filepath.Join(engineDir, "data", "system.def")
	if err := os.WriteFile(systemDefPath, []byte("[Info]\nname = Test System"), 0644); err != nil {
		t.Fatal(err)
	}

	projectDir := filepath.Join(tmpDir, "my-game")
	manifest, err := Scaffold(ScaffoldOptions{
		Name:          "My Game",
		TargetDir:     projectDir,
		EngineVersion: "v0.99.0",
		EngineChannel: "stable",
		EnginePath:    engineDir,
		Author:        "Test Author",
	})

	if err != nil {
		t.Fatalf("Scaffold failed: %v", err)
	}

	if manifest.Name != "My Game" {
		t.Errorf("expected name 'My Game', got %s", manifest.Name)
	}
	if manifest.Engine.Version != "v0.99.0" {
		t.Errorf("expected engine version 'v0.99.0', got %s", manifest.Engine.Version)
	}

	// Verify directories were created
	expectedDirs := []string{"chars", "stages", "data", "font", "sound"}
	for _, dir := range expectedDirs {
		p := filepath.Join(projectDir, dir)
		fi, err := os.Stat(p)
		if err != nil || !fi.IsDir() {
			t.Errorf("expected directory %s to exist", p)
		}
	}

	// Verify select.def was created
	selectDef := filepath.Join(projectDir, "data", "select.def")
	content, err := os.ReadFile(selectDef)
	if err != nil {
		t.Fatalf("expected select.def to exist: %v", err)
	}
	if !strings.Contains(string(content), "[Characters]") {
		t.Errorf("expected select.def to contain [Characters], got: %s", string(content))
	}

	// Verify system.def was copied from engine
	copiedSystemDef := filepath.Join(projectDir, "data", "system.def")
	sysContent, err := os.ReadFile(copiedSystemDef)
	if err != nil {
		t.Fatalf("expected copied system.def to exist: %v", err)
	}
	if !strings.Contains(string(sysContent), "Test System") {
		t.Errorf("expected system.def content 'Test System', got: %s", string(sysContent))
	}

	// Test LoadManifest
	loaded, err := LoadManifest(projectDir)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}
	if loaded.Name != "My Game" {
		t.Errorf("expected loaded name 'My Game', got %s", loaded.Name)
	}
	if loaded.Author != "Test Author" {
		t.Errorf("expected loaded author 'Test Author', got %s", loaded.Author)
	}
}

func TestVerifyAndRepairProject(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-verify-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	engineDir := filepath.Join(tmpDir, "engine")
	projectDir := filepath.Join(tmpDir, "project")

	// Setup engine with mock external/script/main.lua
	scriptDir := filepath.Join(engineDir, "external", "script")
	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptDir, "main.lua"), []byte("print('hello')"), 0644); err != nil {
		t.Fatal(err)
	}

	// Setup project without main.lua
	if err := os.MkdirAll(filepath.Join(projectDir, "chars"), 0755); err != nil {
		t.Fatal(err)
	}

	report, err := VerifyAndRepairProject(engineDir, projectDir)
	if err != nil {
		t.Fatalf("VerifyAndRepairProject failed: %v", err)
	}

	if report.MissingCount != 1 {
		t.Errorf("expected 1 missing file, got %d", report.MissingCount)
	}
	if report.RepairedCount != 1 {
		t.Errorf("expected 1 repaired file, got %d", report.RepairedCount)
	}

	// Verify main.lua exists in project after repair
	repairedFile := filepath.Join(projectDir, "external", "script", "main.lua")
	if _, err := os.Stat(repairedFile); os.IsNotExist(err) {
		t.Errorf("expected %s to be restored", repairedFile)
	}

	// Verify log file was written
	if _, err := os.Stat(report.LogFilePath); os.IsNotExist(err) {
		t.Errorf("expected log file %s to exist", report.LogFilePath)
	}
}

func TestVerifyAndRepairCoreSystemMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-verify-core-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	engineDir := filepath.Join(tmpDir, "engine")
	projectDir := filepath.Join(tmpDir, "project")

	// Setup clean engine common1.cns.zss
	_ = os.MkdirAll(filepath.Join(engineDir, "data"), 0755)
	_ = os.WriteFile(filepath.Join(engineDir, "data", "common1.cns.zss"), []byte("modern_zss_content"), 0644)

	// Setup project with outdated common1.cns.zss and custom select.def
	_ = os.MkdirAll(filepath.Join(projectDir, "data"), 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "data", "common1.cns.zss"), []byte("legacy_outdated_content"), 0644)
	_ = os.WriteFile(filepath.Join(projectDir, "data", "select.def"), []byte("my_custom_character"), 0644)

	// Run Core System Data Update mode
	report, err := VerifyAndRepairProjectWithMode(engineDir, projectDir, true)
	if err != nil {
		t.Fatalf("VerifyAndRepairProjectWithMode failed: %v", err)
	}

	if report.Mode != "core_update" {
		t.Errorf("expected mode 'core_update', got %s", report.Mode)
	}

	// Verify common1.cns.zss was replaced with clean engine version
	updatedContent, _ := os.ReadFile(filepath.Join(projectDir, "data", "common1.cns.zss"))
	if string(updatedContent) != "modern_zss_content" {
		t.Errorf("expected common1.cns.zss to be updated to modern content, got %s", string(updatedContent))
	}

	// Verify custom select.def was NEVER touched
	selectContent, _ := os.ReadFile(filepath.Join(projectDir, "data", "select.def"))
	if string(selectContent) != "my_custom_character" {
		t.Errorf("custom select.def was modified! content: %s", string(selectContent))
	}
}
