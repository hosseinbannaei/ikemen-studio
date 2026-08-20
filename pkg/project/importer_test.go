package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectExistingGame(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-detect-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock MUGEN/Ikemen structure
	_ = os.MkdirAll(filepath.Join(tmpDir, "chars", "kfm"), 0755)
	_ = os.MkdirAll(filepath.Join(tmpDir, "stages"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "stages", "stage0.def"), []byte("[StageInfo]"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "data"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "data", "select.def"), []byte("[Characters]"), 0644)

	inspection, err := DetectExistingGame(tmpDir)
	if err != nil {
		t.Fatalf("DetectExistingGame failed: %v", err)
	}

	if !inspection.IsValid {
		t.Errorf("expected game to be detected as valid")
	}
	if inspection.CharacterCount != 1 {
		t.Errorf("expected 1 character, got %d", inspection.CharacterCount)
	}
	if inspection.StageCount != 1 {
		t.Errorf("expected 1 stage, got %d", inspection.StageCount)
	}
	if !inspection.HasSelectDef {
		t.Errorf("expected HasSelectDef to be true")
	}
}

func TestImportExistingGame(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-import-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "legacy-game")
	_ = os.MkdirAll(filepath.Join(srcDir, "chars", "kfm"), 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "chars", "kfm", "kfm.def"), []byte("[Info]"), 0644)
	_ = os.MkdirAll(filepath.Join(srcDir, "data"), 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "data", "select.def"), []byte("kfm, stages/stage0.def"), 0644)

	engineDir := filepath.Join(tmpDir, "engine")
	_ = os.MkdirAll(filepath.Join(engineDir, "external", "script"), 0755)
	_ = os.WriteFile(filepath.Join(engineDir, "external", "script", "main.lua"), []byte("print('engine')"), 0644)

	targetDir := filepath.Join(tmpDir, "imported-studio-game")

	manifest, err := ImportExistingGame(ImportOptions{
		SourceDir:     srcDir,
		TargetDir:     targetDir,
		ProjectName:   "Imported Game",
		EngineVersion: "nightly",
		EngineChannel: "nightly",
		EnginePath:    engineDir,
		Author:        "Importer",
	})

	if err != nil {
		t.Fatalf("ImportExistingGame failed: %v", err)
	}

	if manifest.Name != "Imported Game" {
		t.Errorf("expected name 'Imported Game', got %s", manifest.Name)
	}

	// Verify original source is untouched and target has files
	if _, err := os.Stat(filepath.Join(srcDir, "chars", "kfm", "kfm.def")); err != nil {
		t.Errorf("original source file was deleted or moved!")
	}

	if _, err := os.Stat(filepath.Join(targetDir, "chars", "kfm", "kfm.def")); err != nil {
		t.Errorf("expected target character file to exist")
	}

	if _, err := os.Stat(filepath.Join(targetDir, "external", "script", "main.lua")); err != nil {
		t.Errorf("expected engine runtime files to be synced")
	}

	if _, err := os.Stat(filepath.Join(targetDir, "ikemen-project.json")); err != nil {
		t.Errorf("expected ikemen-project.json to be created")
	}
}
