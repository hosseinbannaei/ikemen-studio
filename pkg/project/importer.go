package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ExistingGameInspection details the scan results of a raw MUGEN / Ikemen folder
type ExistingGameInspection struct {
	IsValid        bool   `json:"isValid"`
	DetectedName   string `json:"detectedName"`
	CharacterCount int    `json:"characterCount"`
	StageCount     int    `json:"stageCount"`
	HasSelectDef   bool   `json:"hasSelectDef"`
	HasSystemDef   bool   `json:"hasSystemDef"`
	HasConfigIni   bool   `json:"hasConfigIni"`
	SourcePath     string `json:"sourcePath"`
}

// ImportOptions parameters for importing an existing game directory non-destructively
type ImportOptions struct {
	SourceDir     string
	TargetDir     string
	ProjectName   string
	EngineVersion string
	EngineChannel string
	EnginePath    string
	Author        string
}

// DetectExistingGame analyzes a folder to determine if it is an existing MUGEN or Ikemen game
func DetectExistingGame(dir string) (*ExistingGameInspection, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory cannot be empty")
	}

	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	inspection := &ExistingGameInspection{
		SourcePath:   dir,
		DetectedName: filepath.Base(dir),
	}

	// Check data/system.def or data/select.def
	selectDefPath := filepath.Join(dir, "data", "select.def")
	if sfi, err := os.Stat(selectDefPath); err == nil && !sfi.IsDir() {
		inspection.HasSelectDef = true
	}

	systemDefPath := filepath.Join(dir, "data", "system.def")
	if sfi, err := os.Stat(systemDefPath); err == nil && !sfi.IsDir() {
		inspection.HasSystemDef = true
	}

	// Check config.ini
	if _, err := os.Stat(filepath.Join(dir, "save", "config.ini")); err == nil {
		inspection.HasConfigIni = true
	} else if _, err := os.Stat(filepath.Join(dir, "config.ini")); err == nil {
		inspection.HasConfigIni = true
	}

	// Count characters in chars/
	charsDir := filepath.Join(dir, "chars")
	if cEntries, err := os.ReadDir(charsDir); err == nil {
		for _, e := range cEntries {
			if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
				inspection.CharacterCount++
			}
		}
	}

	// Count stages in stages/
	stagesDir := filepath.Join(dir, "stages")
	if sEntries, err := os.ReadDir(stagesDir); err == nil {
		for _, e := range sEntries {
			if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".def") {
				inspection.StageCount++
			}
		}
	}

	// Valid if it has select.def, system.def, or chars/stages content
	if inspection.HasSelectDef || inspection.HasSystemDef || inspection.CharacterCount > 0 || inspection.StageCount > 0 {
		inspection.IsValid = true
	}

	return inspection, nil
}

// ImportExistingGame creates a safe, non-destructive copy of user game content into targetDir with Ikemen Studio manifest
func ImportExistingGame(opts ImportOptions) (*ProjectManifest, error) {
	if opts.SourceDir == "" {
		return nil, fmt.Errorf("source directory cannot be empty")
	}
	if opts.TargetDir == "" {
		return nil, fmt.Errorf("target directory cannot be empty")
	}
	if strings.EqualFold(filepath.Clean(opts.SourceDir), filepath.Clean(opts.TargetDir)) {
		return nil, fmt.Errorf("target directory must be different from source directory to avoid overwriting original files")
	}
	if opts.ProjectName == "" {
		opts.ProjectName = filepath.Base(opts.TargetDir)
	}
	if opts.EngineVersion == "" {
		opts.EngineVersion = "latest"
	}
	if opts.EngineChannel == "" {
		opts.EngineChannel = "stable"
	}

	// 1. Create target directory
	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// 2. Copy user asset folders from source (chars, stages, data, font, sound, video)
	userFolders := []string{"chars", "stages", "data", "font", "sound", "video"}
	for _, folder := range userFolders {
		srcFolder := filepath.Join(opts.SourceDir, folder)
		if fi, err := os.Stat(srcFolder); err == nil && fi.IsDir() {
			destFolder := filepath.Join(opts.TargetDir, folder)
			_ = copyDirRecursive(srcFolder, destFolder)
		} else {
			// Ensure empty standard directory exists if not present in source
			_ = os.MkdirAll(filepath.Join(opts.TargetDir, folder), 0755)
		}
	}

	// 3. Ensure save/logs folder exists
	_ = os.MkdirAll(filepath.Join(opts.TargetDir, "save", "logs"), 0755)

	// Copy config.ini if present in source
	srcConfig := filepath.Join(opts.SourceDir, "save", "config.ini")
	if _, err := os.Stat(srcConfig); os.IsNotExist(err) {
		srcConfig = filepath.Join(opts.SourceDir, "config.ini")
	}
	if _, err := os.Stat(srcConfig); err == nil {
		_ = copyFile(srcConfig, filepath.Join(opts.TargetDir, "save", "config.ini"))
	}

	// 4. Ensure runtime assets (external scripts, lib DLLs) exist from selected engine
	if opts.EnginePath != "" {
		_ = EnsureProjectRuntimeAssets(opts.EnginePath, opts.TargetDir)
	}

	// 5. Ensure select.def exists
	selectDefPath := filepath.Join(opts.TargetDir, "data", "select.def")
	if _, err := os.Stat(selectDefPath); os.IsNotExist(err) {
		_ = os.WriteFile(selectDefPath, []byte(DefaultSelectDefContent), 0644)
	}

	// 6. Create and save project manifest
	manifest := &ProjectManifest{
		Name:    opts.ProjectName,
		Version: "1.0.0",
		Engine: EngineConfig{
			Version: opts.EngineVersion,
			Channel: opts.EngineChannel,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Author:    opts.Author,
		Path:      opts.TargetDir,
	}

	if err := SaveManifest(opts.TargetDir, manifest); err != nil {
		return nil, fmt.Errorf("failed to save manifest: %w", err)
	}

	return manifest, nil
}
