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
	IsValid                bool   `json:"isValid"`
	DetectedName           string `json:"detectedName"`
	CharacterCount         int    `json:"characterCount"`
	StageCount             int    `json:"stageCount"`
	HasSelectDef           bool   `json:"hasSelectDef"`
	HasSystemDef           bool   `json:"hasSystemDef"`
	HasConfigIni           bool   `json:"hasConfigIni"`
	SourcePath             string `json:"sourcePath"`
	DetectedEngineVersion  string `json:"detectedEngineVersion"`
}

// ImportOptions parameters for importing an existing game directory
type ImportOptions struct {
	SourceDir           string `json:"sourceDir"`
	TargetDir           string `json:"targetDir"`
	ProjectName         string `json:"projectName"`
	EngineVersion       string `json:"engineVersion"`
	EngineChannel       string `json:"engineChannel"`
	EnginePath          string `json:"enginePath"`
	BaselineEnginePath  string `json:"baselineEnginePath"`
	Author              string `json:"author"`
	Mode                string `json:"mode"` // "rebuild", "diff_upgrade", "legacy_match"
	IncludeChars        bool   `json:"includeChars"`
	IncludeStages       bool   `json:"includeStages"`
	IncludeSound        bool   `json:"includeSound"`
	IncludeFonts        bool   `json:"includeFonts"`
	IncludeRoster       bool   `json:"includeRoster"`
	IncludeLegacySystem bool   `json:"includeLegacySystem"`
}

// DetectExistingGame analyzes a folder to determine if it is an existing MUGEN or Ikemen game and detects engine version
func DetectExistingGame(dir string) (*ExistingGameInspection, error) {
	if dir == "" {
		return nil, fmt.Errorf("directory cannot be empty")
	}

	fi, err := os.Stat(dir)
	if err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("directory does not exist: %s", dir)
	}

	inspection := &ExistingGameInspection{
		SourcePath:            dir,
		DetectedName:          filepath.Base(dir),
		DetectedEngineVersion: DetectLegacyEngineVersion(dir),
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

	if inspection.HasSelectDef || inspection.HasSystemDef || inspection.CharacterCount > 0 || inspection.StageCount > 0 {
		inspection.IsValid = true
	}

	return inspection, nil
}

// ImportExistingGame creates a clean, non-destructive import of user game content into targetDir
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
		opts.EngineVersion = "nightly"
	}
	if opts.EngineChannel == "" {
		opts.EngineChannel = "stable"
	}
	if opts.Mode == "" {
		opts.Mode = "rebuild"
	}
	if opts.Mode == "rebuild" && !opts.IncludeChars && !opts.IncludeStages && !opts.IncludeSound && !opts.IncludeFonts && !opts.IncludeRoster {
		opts.IncludeChars = true
		opts.IncludeStages = true
		opts.IncludeSound = true
		opts.IncludeFonts = true
		opts.IncludeRoster = true
	}

	// 1. Create target directory
	if err := os.MkdirAll(opts.TargetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Ensure basic target folders exist
	for _, folder := range []string{"chars", "stages", "data", "font", "sound", "video", "save/logs"} {
		_ = os.MkdirAll(filepath.Join(opts.TargetDir, folder), 0755)
	}

	// 2. Sync clean engine runtime assets from selected official engine first
	if opts.EnginePath != "" {
		_ = EnsureProjectRuntimeAssets(opts.EnginePath, opts.TargetDir)
		// Copy clean engine data templates (like clean common1.cns.zss, etc.)
		engineDataDir := filepath.Join(opts.EnginePath, "data")
		if dfi, err := os.Stat(engineDataDir); err == nil && dfi.IsDir() {
			_ = copyDirRecursive(engineDataDir, filepath.Join(opts.TargetDir, "data"))
		}
	}

	// 3. Migrate user assets according to selected mode / checklist
	if opts.IncludeChars || opts.Mode == "legacy_match" {
		srcChars := filepath.Join(opts.SourceDir, "chars")
		if fi, err := os.Stat(srcChars); err == nil && fi.IsDir() {
			_ = copyDirRecursive(srcChars, filepath.Join(opts.TargetDir, "chars"))
		}
	}

	if opts.IncludeStages || opts.Mode == "legacy_match" {
		srcStages := filepath.Join(opts.SourceDir, "stages")
		if fi, err := os.Stat(srcStages); err == nil && fi.IsDir() {
			_ = copyDirRecursive(srcStages, filepath.Join(opts.TargetDir, "stages"))
		}
	}

	if opts.IncludeSound || opts.Mode == "legacy_match" {
		srcSound := filepath.Join(opts.SourceDir, "sound")
		if fi, err := os.Stat(srcSound); err == nil && fi.IsDir() {
			_ = copyDirRecursive(srcSound, filepath.Join(opts.TargetDir, "sound"))
		}
	}

	if opts.IncludeFonts || opts.Mode == "legacy_match" {
		srcFont := filepath.Join(opts.SourceDir, "font")
		if fi, err := os.Stat(srcFont); err == nil && fi.IsDir() {
			_ = copyDirRecursive(srcFont, filepath.Join(opts.TargetDir, "font"))
		}
	}

	// Migrate select.def (user's roster)
	if opts.IncludeRoster || opts.Mode == "legacy_match" {
		srcSelect := filepath.Join(opts.SourceDir, "data", "select.def")
		if _, err := os.Stat(srcSelect); err == nil {
			_ = copyFile(srcSelect, filepath.Join(opts.TargetDir, "data", "select.def"))
		}
	}

	// In legacy_match or if requested, copy custom system files
	if opts.IncludeLegacySystem || opts.Mode == "legacy_match" {
		srcData := filepath.Join(opts.SourceDir, "data")
		if fi, err := os.Stat(srcData); err == nil && fi.IsDir() {
			_ = copyDirRecursive(srcData, filepath.Join(opts.TargetDir, "data"))
		}
	}

	// Copy config.ini if present in source
	srcConfig := filepath.Join(opts.SourceDir, "save", "config.ini")
	if _, err := os.Stat(srcConfig); os.IsNotExist(err) {
		srcConfig = filepath.Join(opts.SourceDir, "config.ini")
	}
	if _, err := os.Stat(srcConfig); err == nil {
		_ = copyFile(srcConfig, filepath.Join(opts.TargetDir, "save", "config.ini"))
	}

	// 4. Ensure select.def exists if not copied
	selectDefPath := filepath.Join(opts.TargetDir, "data", "select.def")
	if _, err := os.Stat(selectDefPath); os.IsNotExist(err) {
		_ = os.WriteFile(selectDefPath, []byte(DefaultSelectDefContent), 0644)
	}

	// 5. Create and save project manifest
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
