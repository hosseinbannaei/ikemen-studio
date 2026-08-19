package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
}

// VerifyAndRepairProject compares project files against engine runtime assets and restores any missing files
func VerifyAndRepairProject(engineDir, projectDir string) (*VerificationReport, error) {
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
	}

	var logLines []string
	logLines = append(logLines, fmt.Sprintf("=== Ikemen Studio Project Verification - %s ===", time.Now().Format(time.RFC3339)))
	logLines = append(logLines, fmt.Sprintf("Project Path: %s", projectDir))
	logLines = append(logLines, fmt.Sprintf("Engine Path:  %s", engineDir))
	logLines = append(logLines, "--------------------------------------------------------")

	coreFolders := []string{"external", "lib", "data", "font", "sound", "stages", "video", "chars"}

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

			report.TotalChecked++
			targetFile := filepath.Join(projectDir, relPath)

			targetInfo, statErr := os.Stat(targetFile)
			if os.IsNotExist(statErr) || (statErr == nil && targetInfo.Size() == 0 && info.Size() > 0) {
				report.MissingCount++

				// Perform repair
				copyErr := copyFile(path, targetFile)
				if copyErr == nil {
					report.RepairedCount++
					report.RepairedFiles = append(report.RepairedFiles, relPath)
					logLines = append(logLines, fmt.Sprintf("[RESTORED] %s", relPath))
				} else {
					logLines = append(logLines, fmt.Sprintf("[FAILED]   %s (Error: %v)", relPath, copyErr))
				}
			} else {
				logLines = append(logLines, fmt.Sprintf("[OK]       %s", relPath))
			}

			return nil
		})
	}

	logLines = append(logLines, "--------------------------------------------------------")
	logLines = append(logLines, fmt.Sprintf("Total Files Checked: %d", report.TotalChecked))
	logLines = append(logLines, fmt.Sprintf("Missing / Damaged:   %d", report.MissingCount))
	logLines = append(logLines, fmt.Sprintf("Successfully Fixed:  %d", report.RepairedCount))
	logLines = append(logLines, "Verification complete.")

	_ = os.WriteFile(logFilePath, []byte(strings.Join(logLines, "\n")), 0644)

	return report, nil
}
