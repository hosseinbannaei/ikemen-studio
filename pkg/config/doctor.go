package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ConfigIssue represents an invalid or unsupported configuration parameter
type ConfigIssue struct {
	Key            string `json:"key"`
	CurrentValue   string `json:"currentValue"`
	SuggestedValue string `json:"suggestedValue"`
	Severity       string `json:"severity"` // "error", "warning", "info"
	Description    string `json:"description"`
}

// ConfigInspectionResult details findings from analyzing project save/config.ini
type ConfigInspectionResult struct {
	IsValid             bool          `json:"isValid"`
	HasLegacyRenderMode bool          `json:"hasLegacyRenderMode"`
	Issues              []ConfigIssue `json:"issues"`
	TotalKeys           int           `json:"totalKeys"`
	ConfigPath          string        `json:"configPath"`
}

// InspectGameConfig analyzes save/config.ini and identifies obsolete, invalid, or crashing parameters
func InspectGameConfig(projectDir, engineDir string) (*ConfigInspectionResult, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("project directory cannot be empty")
	}

	iniPath := filepath.Join(projectDir, "save", "config.ini")
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		iniPath = filepath.Join(projectDir, "config.ini")
	}

	result := &ConfigInspectionResult{
		IsValid:    true,
		Issues:     make([]ConfigIssue, 0),
		ConfigPath: iniPath,
	}

	cfg, err := LoadGameConfig(projectDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read game config: %w", err)
	}

	result.TotalKeys = len(cfg)

	// 1. Validate RenderMode (Valid desktop values: "OpenGL 3.3" or "Vulkan 1.3")
	renderModeFound := false
	for _, rKey := range []string{"RenderMode", "rendermode"} {
		if val, ok := cfg[rKey]; ok {
			renderModeFound = true
			valTrimmed := strings.Trim(val, `"' `)
			if !strings.EqualFold(valTrimmed, "OpenGL 3.3") && !strings.EqualFold(valTrimmed, "Vulkan 1.3") {
				result.IsValid = false
				result.HasLegacyRenderMode = true
				result.Issues = append(result.Issues, ConfigIssue{
					Key:            rKey,
					CurrentValue:   val,
					SuggestedValue: "OpenGL 3.3",
					Severity:       "error",
					Description:    "Invalid RenderMode causes graphics initialization warning or crash. Recommended: OpenGL 3.3.",
				})
			}
		}
	}

	if !renderModeFound {
		result.Issues = append(result.Issues, ConfigIssue{
			Key:            "RenderMode",
			CurrentValue:   "(missing)",
			SuggestedValue: "OpenGL 3.3",
			Severity:       "warning",
			Description:    "RenderMode is not explicitly configured in save/config.ini.",
		})
	}

	// 2. Validate Width & Height
	if w, ok := cfg["Width"]; ok {
		if n, err := strconv.Atoi(w); err != nil || n < 320 || n > 7680 {
			result.IsValid = false
			result.Issues = append(result.Issues, ConfigIssue{
				Key:            "Width",
				CurrentValue:   w,
				SuggestedValue: "1280",
				Severity:       "error",
				Description:    "Invalid display width resolution value.",
			})
		}
	}

	if h, ok := cfg["Height"]; ok {
		if n, err := strconv.Atoi(h); err != nil || n < 240 || n > 4320 {
			result.IsValid = false
			result.Issues = append(result.Issues, ConfigIssue{
				Key:            "Height",
				CurrentValue:   h,
				SuggestedValue: "720",
				Severity:       "error",
				Description:    "Invalid display height resolution value.",
			})
		}
	}

	// 3. Validate Audio Volumes
	for _, vKey := range []string{"VolumeMaster", "VolumeBgm", "VolumeSfx"} {
		if v, ok := cfg[vKey]; ok {
			if n, err := strconv.Atoi(v); err != nil || n < 0 || n > 100 {
				result.Issues = append(result.Issues, ConfigIssue{
					Key:            vKey,
					CurrentValue:   v,
					SuggestedValue: "80",
					Severity:       "warning",
					Description:    "Volume should be between 0 and 100.",
				})
			}
		}
	}

	// 4. Validate Vsync & Fullscreen flags
	for _, bKey := range []string{"Fullscreen", "Vsync"} {
		if b, ok := cfg[bKey]; ok {
			if b != "0" && b != "1" {
				result.Issues = append(result.Issues, ConfigIssue{
					Key:            bKey,
					CurrentValue:   b,
					SuggestedValue: "0",
					Severity:       "warning",
					Description:    "Boolean flag should be 0 (off) or 1 (on).",
				})
			}
		}
	}

	if len(result.Issues) > 0 {
		result.IsValid = false
	}

	return result, nil
}

// RepairGameConfig automatically fixes invalid configuration keys and normalizes to engine standards
func RepairGameConfig(projectDir, engineDir string) error {
	inspection, err := InspectGameConfig(projectDir, engineDir)
	if err != nil {
		return err
	}

	updates := make(map[string]string)
	for _, issue := range inspection.Issues {
		updates[issue.Key] = issue.SuggestedValue
	}

	// Always ensure valid desktop RenderMode
	updates["RenderMode"] = "OpenGL 3.3"

	return SaveGameConfig(projectDir, updates)
}

// ResetGameConfig restores save/config.ini to clean official engine defaults
func ResetGameConfig(projectDir, engineDir string) error {
	if projectDir == "" {
		return fmt.Errorf("project directory cannot be empty")
	}

	saveDir := filepath.Join(projectDir, "save")
	_ = os.MkdirAll(saveDir, 0755)
	destPath := filepath.Join(saveDir, "config.ini")

	// If engine has a clean config.ini template, copy it
	if engineDir != "" {
		srcPath := filepath.Join(engineDir, "save", "config.ini")
		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			srcPath = filepath.Join(engineDir, "config.ini")
		}

		if data, err := os.ReadFile(srcPath); err == nil && len(data) > 0 {
			return os.WriteFile(destPath, data, 0644)
		}
	}

	// Standard clean default configuration
	defaultConfigContent := `; Ikemen GO Studio Default Configuration
[Video]
RenderMode = OpenGL 3.3
Width = 1280
Height = 720
Fullscreen = 0
Vsync = 1

[Audio]
VolumeMaster = 80
VolumeBgm = 80
VolumeSfx = 80
`
	return os.WriteFile(destPath, []byte(defaultConfigContent), 0644)
}
