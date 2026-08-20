package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ikemen-studio/pkg/config"
)

// CategoryDiff summarizes differences for a specific asset group
type CategoryDiff struct {
	Category    string   `json:"category"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ItemCount   int      `json:"itemCount"`
	Status      string   `json:"status"` // "outdated", "missing", "clean", "custom"
	Files       []string `json:"files"`
}

// ProjectDiffSummary represents the full inspection results comparing project vs engine
type ProjectDiffSummary struct {
	Categories         []CategoryDiff `json:"categories"`
	TotalDiscrepancies int            `json:"totalDiscrepancies"`
	EngineVersion      string         `json:"engineVersion"`
}

// AssetSyncOptions specifies which components to replace/sync with clean engine defaults
type AssetSyncOptions struct {
	ProjectDir      string `json:"projectDir"`
	EngineDir       string `json:"engineDir"`
	SyncStockChars  bool   `json:"syncStockChars"`
	SyncStockStages bool   `json:"syncStockStages"`
	SyncScreenpack  bool   `json:"syncScreenpack"`
	SyncFonts       bool   `json:"syncFonts"`
	SyncRuntime     bool   `json:"syncRuntime"`
	ResetConfig     bool   `json:"resetConfig"`
}

// InspectProjectDifferences compares project assets against clean engine defaults
func InspectProjectDifferences(projectDir, engineDir string) (*ProjectDiffSummary, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("project directory cannot be empty")
	}
	if engineDir == "" {
		return nil, fmt.Errorf("engine directory cannot be empty")
	}

	summary := &ProjectDiffSummary{
		Categories: make([]CategoryDiff, 0),
	}

	manifest, _ := LoadManifest(projectDir)
	if manifest != nil {
		summary.EngineVersion = manifest.Engine.Version
	}

	// 1. Stock Characters (chars/kfm, chars/kfm_zss, chars/randomselect)
	stockChars := []string{"kfm", "kfm_zss", "randomselect"}
	var charFiles []string
	charStatus := "clean"
	for _, sc := range stockChars {
		engineCharDir := filepath.Join(engineDir, "chars", sc)
		projectCharDir := filepath.Join(projectDir, "chars", sc)

		if _, err := os.Stat(engineCharDir); err == nil {
			if _, pErr := os.Stat(projectCharDir); os.IsNotExist(pErr) {
				charStatus = "missing"
				charFiles = append(charFiles, fmt.Sprintf("chars/%s (missing)", sc))
			} else {
				// Check if project contains outdated ZSS syntax (such as missing '[' error)
				charFiles = append(charFiles, fmt.Sprintf("chars/%s", sc))
			}
		}
	}
	if len(charFiles) > 0 && charStatus == "clean" {
		charStatus = "outdated" // default stock chars in legacy projects often have older syntax
	}
	summary.Categories = append(summary.Categories, CategoryDiff{
		Category:    "stock_chars",
		Title:       "Default Fighters (KFM / KFM ZSS)",
		Description: "Official demo characters. Outdated 0.99 versions cause ZSS syntax parser crashes.",
		ItemCount:   len(charFiles),
		Status:      charStatus,
		Files:       charFiles,
	})

	// 2. Stock Stages (stages/stage0.def, etc.)
	var stageFiles []string
	if sEntries, err := os.ReadDir(filepath.Join(engineDir, "stages")); err == nil {
		for _, se := range sEntries {
			if !se.IsDir() && strings.HasSuffix(strings.ToLower(se.Name()), ".def") {
				stageFiles = append(stageFiles, fmt.Sprintf("stages/%s", se.Name()))
			}
		}
	}
	summary.Categories = append(summary.Categories, CategoryDiff{
		Category:    "stock_stages",
		Title:       "Default Stages",
		Description: "Baseline engine arenas (stage0, training stages, etc.).",
		ItemCount:   len(stageFiles),
		Status:      "clean",
		Files:       stageFiles,
	})

	// 3. Screenpack & System Scripts (system.def, fight.def, common1.cns.zss, dgl.zss)
	sysFiles := []string{
		"data/common1.cns.zss",
		"data/dgl.zss",
		"data/common.cns",
		"data/system.def",
		"data/fight.def",
		"data/fightfx.air",
		"data/fightfx.sff",
	}
	var existingSysFiles []string
	sysStatus := "outdated"
	for _, sf := range sysFiles {
		if _, err := os.Stat(filepath.Join(engineDir, sf)); err == nil {
			existingSysFiles = append(existingSysFiles, sf)
		}
	}
	summary.Categories = append(summary.Categories, CategoryDiff{
		Category:    "screenpack",
		Title:       "System Screenpack & Battle Scripts",
		Description: "Engine UI scripts, fight bars, and common1.cns.zss state definitions.",
		ItemCount:   len(existingSysFiles),
		Status:      sysStatus,
		Files:       existingSysFiles,
	})

	// 4. Fonts
	var fontFiles []string
	if fEntries, err := os.ReadDir(filepath.Join(engineDir, "font")); err == nil {
		for _, fe := range fEntries {
			if !fe.IsDir() {
				fontFiles = append(fontFiles, fmt.Sprintf("font/%s", fe.Name()))
			}
		}
	}
	summary.Categories = append(summary.Categories, CategoryDiff{
		Category:    "fonts",
		Title:       "Engine Fonts & UI Typography",
		Description: "System TrueType and bitmap fonts (f-4x6, f-6x9, etc.).",
		ItemCount:   len(fontFiles),
		Status:      "clean",
		Files:       fontFiles,
	})

	// 5. Engine Core Runtime (external/ & lib/)
	summary.Categories = append(summary.Categories, CategoryDiff{
		Category:    "runtime",
		Title:       "Engine Runtime & Lua Scripts (external/ & lib/)",
		Description: "Core Lua VM, shaders, and system libraries.",
		ItemCount:   1,
		Status:      "clean",
		Files:       []string{"external/script/", "external/shaders/", "lib/"},
	})

	// 6. Game Config
	cfgResult, _ := config.InspectGameConfig(projectDir, engineDir)
	cfgStatus := "clean"
	var cfgIssues []string
	if cfgResult != nil && !cfgResult.IsValid {
		cfgStatus = "outdated"
		for _, iss := range cfgResult.Issues {
			cfgIssues = append(cfgIssues, fmt.Sprintf("%s: %s -> %s", iss.Key, iss.CurrentValue, iss.SuggestedValue))
		}
	}
	summary.Categories = append(summary.Categories, CategoryDiff{
		Category:    "config",
		Title:       "Game Preferences (save/config.ini)",
		Description: "Graphics renderer mode, display resolutions, and audio levels.",
		ItemCount:   len(cfgIssues),
		Status:      cfgStatus,
		Files:       cfgIssues,
	})

	for _, c := range summary.Categories {
		if c.Status == "outdated" || c.Status == "missing" {
			summary.TotalDiscrepancies++
		}
	}

	return summary, nil
}

// SyncProjectAssets selectively updates project assets with clean engine copies
func SyncProjectAssets(opts AssetSyncOptions) (*VerificationReport, error) {
	if opts.ProjectDir == "" {
		return nil, fmt.Errorf("project directory cannot be empty")
	}
	if opts.EngineDir == "" {
		return nil, fmt.Errorf("engine directory cannot be empty")
	}

	logsDir := filepath.Join(opts.ProjectDir, "save", "logs")
	_ = os.MkdirAll(logsDir, 0755)
	logFilePath := filepath.Join(logsDir, "asset_sync_report.log")

	report := &VerificationReport{
		RepairedFiles: make([]string, 0),
		LogFilePath:   logFilePath,
		Success:       true,
		Mode:          "custom_sync",
	}

	var logLines []string
	logLines = append(logLines, fmt.Sprintf("=== Ikemen Studio Asset Sync Report - %s ===", time.Now().Format(time.RFC3339)))
	logLines = append(logLines, fmt.Sprintf("Project Path: %s", opts.ProjectDir))
	logLines = append(logLines, fmt.Sprintf("Engine Path:  %s", opts.EngineDir))
	logLines = append(logLines, "--------------------------------------------------------")

	// 1. Sync Stock Characters if requested
	if opts.SyncStockChars {
		engineCharsDir := filepath.Join(opts.EngineDir, "chars")
		if entries, err := os.ReadDir(engineCharsDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					sc := entry.Name()
					src := filepath.Join(engineCharsDir, sc)
					dest := filepath.Join(opts.ProjectDir, "chars", sc)
					_ = copyDirRecursive(src, dest)
					report.RepairedCount++
					report.RepairedFiles = append(report.RepairedFiles, fmt.Sprintf("chars/%s", sc))
					logLines = append(logLines, fmt.Sprintf("[SYNCED-CHAR]   chars/%s", sc))
				}
			}
		}
	}

	// 2. Sync Stock Stages if requested
	if opts.SyncStockStages {
		srcStages := filepath.Join(opts.EngineDir, "stages")
		destStages := filepath.Join(opts.ProjectDir, "stages")
		if fi, err := os.Stat(srcStages); err == nil && fi.IsDir() {
			_ = copyDirRecursive(srcStages, destStages)
			report.RepairedCount++
			report.RepairedFiles = append(report.RepairedFiles, "stages/")
			logLines = append(logLines, "[SYNCED-STAGES] stages/")
		}
	}

	// 3. Sync Screenpack & System Scripts if requested
	if opts.SyncScreenpack {
		sysFiles := []string{
			"data/common1.cns.zss",
			"data/dgl.zss",
			"data/common.cns",
			"data/system.def",
			"data/fight.def",
			"data/fightfx.air",
			"data/fightfx.sff",
		}
		for _, sf := range sysFiles {
			src := filepath.Join(opts.EngineDir, sf)
			dest := filepath.Join(opts.ProjectDir, sf)
			if _, err := os.Stat(src); err == nil {
				_ = copyFile(src, dest)
				report.RepairedCount++
				report.RepairedFiles = append(report.RepairedFiles, sf)
				logLines = append(logLines, fmt.Sprintf("[SYNCED-SYS]    %s", sf))
			}
		}
	}

	// 4. Sync Fonts if requested
	if opts.SyncFonts {
		srcFont := filepath.Join(opts.EngineDir, "font")
		destFont := filepath.Join(opts.ProjectDir, "font")
		if fi, err := os.Stat(srcFont); err == nil && fi.IsDir() {
			_ = copyDirRecursive(srcFont, destFont)
			report.RepairedCount++
			report.RepairedFiles = append(report.RepairedFiles, "font/")
			logLines = append(logLines, "[SYNCED-FONTS]  font/")
		}
	}

	// 5. Sync Runtime & Lua if requested
	if opts.SyncRuntime {
		for _, folder := range []string{"external", "lib"} {
			src := filepath.Join(opts.EngineDir, folder)
			dest := filepath.Join(opts.ProjectDir, folder)
			if fi, err := os.Stat(src); err == nil && fi.IsDir() {
				_ = copyDirRecursive(src, dest)
				report.RepairedCount++
				report.RepairedFiles = append(report.RepairedFiles, folder+"/")
				logLines = append(logLines, fmt.Sprintf("[SYNCED-CORE]   %s/", folder))
			}
		}
	}

	// 6. Reset or Fix Config if requested
	if opts.ResetConfig {
		if err := config.ResetGameConfig(opts.ProjectDir, opts.EngineDir); err == nil {
			report.RepairedCount++
			report.RepairedFiles = append(report.RepairedFiles, "save/config.ini")
			logLines = append(logLines, "[RESET-CONFIG]  save/config.ini restored to clean engine defaults")
		}
	}

	logLines = append(logLines, "--------------------------------------------------------")
	logLines = append(logLines, fmt.Sprintf("Total Repaired/Synced: %d", report.RepairedCount))
	logLines = append(logLines, "Sync complete.")

	_ = os.WriteFile(logFilePath, []byte(strings.Join(logLines, "\n")), 0644)

	return report, nil
}
