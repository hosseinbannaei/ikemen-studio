package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ikemen-studio/pkg/config"
)

// VerificationReport details the results of verifying and repairing project assets
type VerificationReport struct {
	TotalChecked  int      `json:"totalChecked"`
	MissingCount  int      `json:"missingCount"`
	RepairedCount int      `json:"repairedCount"`
	RepairedFiles []string `json:"repairedFiles"`
	LogFilePath   string   `json:"logFilePath"`
	Success       bool     `json:"success"`
	ErrorMessage  string   `json:"errorMessage,omitempty"`
	Mode          string   `json:"mode"`
}

// VerifyAndRepairProject performs standard file verification
func VerifyAndRepairProject(engineDir, projectDir string) (*VerificationReport, error) {
	return VerifyAndRepairProjectWithMode(engineDir, projectDir, false)
}

// VerifyAndRepairProjectWithMode compares project files against engine runtime assets
// If updateCoreSystem is true, it updates core system data scripts and stock characters
func VerifyAndRepairProjectWithMode(engineDir, projectDir string, updateCoreSystem bool) (*VerificationReport, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("project directory cannot be empty")
	}
	if engineDir == "" {
		return nil, fmt.Errorf("engine directory cannot be empty")
	}

	if fi, err := os.Stat(engineDir); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("configured engine directory not found: %s", engineDir)
	}

	// Prepare logs directory
	logsDir := filepath.Join(projectDir, "save", "logs")
	_ = os.MkdirAll(logsDir, 0755)
	logFilePath := filepath.Join(logsDir, "verify_report.log")

	report := &VerificationReport{
		RepairedFiles: make([]string, 0),
		LogFilePath:   logFilePath,
		Success:       true,
		Mode:          "standard",
	}
	if updateCoreSystem {
		report.Mode = "core_update"
	}

	var logLines []string
	logLines = append(logLines, fmt.Sprintf("=== Ikemen Studio Project Verification - %s ===", time.Now().Format(time.RFC3339)))
	logLines = append(logLines, fmt.Sprintf("Project Path: %s", projectDir))
	logLines = append(logLines, fmt.Sprintf("Engine Path:  %s", engineDir))
	logLines = append(logLines, fmt.Sprintf("Mode:         %s", report.Mode))
	logLines = append(logLines, "--------------------------------------------------------")

	coreFolders := []string{"external", "lib", "data", "font", "sound", "stages", "video", "chars"}

	// Core engine system files that should be updated if updateCoreSystem is true
	coreSystemFiles := map[string]bool{
		"data/common1.cns.zss": true,
		"data/dgl.zss":         true,
		"data/common.cns":      true,
		"data/fightfx.air":     true,
		"data/fightfx.sff":     true,
	}

	for _, folder := range coreFolders {
		srcFolder := filepath.Join(engineDir, folder)
		if fi, err := os.Stat(srcFolder); err != nil || !fi.IsDir() {
			continue
		}

		destFolder := filepath.Join(projectDir, folder)
		_ = os.MkdirAll(destFolder, 0755)

		_ = filepath.Walk(srcFolder, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(engineDir, path)
			if err != nil {
				return nil
			}

			normalizedRel := filepath.ToSlash(relPath)

			// Never overwrite select.def
			if normalizedRel == "data/select.def" {
				return nil
			}

			report.TotalChecked++
			targetFile := filepath.Join(projectDir, relPath)

			targetInfo, statErr := os.Stat(targetFile)
			isMissing := os.IsNotExist(statErr) || (statErr == nil && targetInfo.Size() == 0 && info.Size() > 0)
			
			// Stock characters that must be updated to avoid ZSS syntax panics
			isStockChar := strings.HasPrefix(normalizedRel, "chars/kfm") || strings.HasPrefix(normalizedRel, "chars/randomselect")
			isCoreSystemScript := updateCoreSystem && (strings.HasPrefix(normalizedRel, "external/") || coreSystemFiles[normalizedRel] || isStockChar)

			if isMissing || isCoreSystemScript {
				if isMissing {
					report.MissingCount++
				}

				// Perform repair / update
				copyErr := copyFile(path, targetFile)
				if copyErr == nil {
					report.RepairedCount++
					report.RepairedFiles = append(report.RepairedFiles, relPath)
					if isCoreSystemScript && !isMissing {
						logLines = append(logLines, fmt.Sprintf("[UPDATED-CORE] %s", relPath))
					} else {
						logLines = append(logLines, fmt.Sprintf("[RESTORED]     %s", relPath))
					}
				} else {
					logLines = append(logLines, fmt.Sprintf("[FAILED]       %s (Error: %v)", relPath, copyErr))
				}
			} else {
				logLines = append(logLines, fmt.Sprintf("[OK]           %s", relPath))
			}

			return nil
		})
	}

	// Always ensure config.ini has valid RenderMode
	if err := config.RepairGameConfig(projectDir, engineDir); err == nil {
		logLines = append(logLines, "[NORMALIZED-CFG] save/config.ini RenderMode set to OpenGL 3.3")
	}

	logLines = append(logLines, "--------------------------------------------------------")
	logLines = append(logLines, fmt.Sprintf("Total Files Checked: %d", report.TotalChecked))
	logLines = append(logLines, fmt.Sprintf("Missing Identified:  %d", report.MissingCount))
	logLines = append(logLines, fmt.Sprintf("Updated / Repaired:  %d", report.RepairedCount))
	logLines = append(logLines, "Verification complete.")

	_ = os.WriteFile(logFilePath, []byte(strings.Join(logLines, "\n")), 0644)

	return report, nil
}
