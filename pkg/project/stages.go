package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"ikemen-studio/pkg/vault"
)

// ProjectStageInfo represents detailed metadata about an installed stage in a project.
type ProjectStageInfo struct {
	RelativePath       string   `json:"relative_path"`
	DisplayName        string   `json:"display_name"`
	Author             string   `json:"author"`
	Version            string   `json:"version"`
	BgmPath            string   `json:"bgm_path"`
	PreviewBase64      string   `json:"preview_base64"`
	IsExtraStage       bool     `json:"is_extra_stage"`
	AssignedCharacters []string `json:"assigned_characters"`
	XScale             float64  `json:"xscale"`
	YScale             float64  `json:"yscale"`
	ZOffset            float64  `json:"zoffset"`
	IsLinkedFromVault  bool     `json:"is_linked_from_vault"`
}

// GetProjectStages scans the project's stages/ folder and select.def to return all installed stages.
func GetProjectStages(projectDir string) ([]ProjectStageInfo, error) {
	stagesDir := filepath.Join(projectDir, "stages")
	if _, err := os.Stat(stagesDir); os.IsNotExist(err) {
		return []ProjectStageInfo{}, nil
	}

	// 1. Parse select.def to discover extra stages and character assignments
	selectDefPath := filepath.Join(projectDir, "data", "select.def")
	extraStagesMap, charStageMap := parseSelectDefStageMappings(selectDefPath)

	var stages []ProjectStageInfo

	// 2. Discover all .def files in stages/
	err := filepath.Walk(stagesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) == ".def" {
			rel, _ := filepath.Rel(projectDir, path)
			rel = filepath.ToSlash(rel)

			stageMeta := parseStageDef(path)
			if stageMeta == nil {
				return nil
			}

			// Check symlink
			isSymlink := false
			if fi, lErr := os.Lstat(path); lErr == nil && (fi.Mode()&os.ModeSymlink != 0) {
				isSymlink = true
			}
			if !isSymlink {
				parentDir := filepath.Dir(path)
				if fi, lErr := os.Lstat(parentDir); lErr == nil && (fi.Mode()&os.ModeSymlink != 0) {
					isSymlink = true
				}
			}

			// Extract sprite / portrait if possible
			previewBase64 := ""
			sffPath := stageMeta.SpriteFile
			if sffPath != "" {
				fullSff := filepath.Join(filepath.Dir(path), sffPath)
				if _, statErr := os.Stat(fullSff); statErr == nil {
					_, b64, _ := vault.ExtractAndCachePortrait(filepath.Dir(path), filepath.Base(path), fullSff)
					previewBase64 = b64
				}
			}

			assignedChars := charStageMap[normalizeStageKey(rel)]
			isExtra := extraStagesMap[normalizeStageKey(rel)]

			stage := ProjectStageInfo{
				RelativePath:       rel,
				DisplayName:        stageMeta.DisplayName,
				Author:             stageMeta.Author,
				Version:            stageMeta.MugenVersion,
				BgmPath:            stageMeta.SoundFile,
				PreviewBase64:      previewBase64,
				IsExtraStage:       isExtra,
				AssignedCharacters: assignedChars,
				XScale:             stageMeta.XScale,
				YScale:             stageMeta.YScale,
				ZOffset:            stageMeta.ZOffset,
				IsLinkedFromVault:  isSymlink,
			}

			stages = append(stages, stage)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan stages directory: %w", err)
	}

	return stages, nil
}

// ToggleStageExtraStage adds or removes a stage from the [ExtraStages] section of select.def.
func ToggleStageExtraStage(projectDir, stageRelativePath string, enable bool) error {
	selectDefPath := filepath.Join(projectDir, "data", "select.def")
	if _, err := os.Stat(selectDefPath); os.IsNotExist(err) {
		return fmt.Errorf("select.def not found in project")
	}

	data, err := os.ReadFile(selectDefPath)
	if err != nil {
		return err
	}

	cleanTarget := filepath.ToSlash(strings.TrimSpace(stageRelativePath))
	normTarget := normalizeStageKey(cleanTarget)

	lines := strings.Split(string(data), "\n")
	var newLines []string
	inExtraStages := false
	stageFound := false
	hasExtraStagesHeader := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "[") {
			if strings.HasPrefix(lower, "[extrastages]") {
				inExtraStages = true
				hasExtraStagesHeader = true
				newLines = append(newLines, line)
				continue
			} else {
				// Exiting extra stages section
				if inExtraStages && enable && !stageFound {
					newLines = append(newLines, cleanTarget)
					stageFound = true
				}
				inExtraStages = false
			}
		}

		if inExtraStages {
			if trimmed == "" || strings.HasPrefix(trimmed, ";") {
				newLines = append(newLines, line)
				continue
			}
			entryPath := normalizeStageKey(trimmed)
			if entryPath == normTarget {
				stageFound = true
				if enable {
					newLines = append(newLines, cleanTarget)
				}
				// If disable, we omit this line
				continue
			}
		}

		newLines = append(newLines, line)
	}

	if enable && !stageFound {
		if hasExtraStagesHeader {
			// Append before EOF
			newLines = append(newLines, cleanTarget)
		} else {
			newLines = append(newLines, "\n[ExtraStages]", cleanTarget)
		}
	}

	return os.WriteFile(selectDefPath, []byte(strings.Join(newLines, "\n")), 0644)
}

// SetFighterHomeStage assigns or clears a home stage for a given character in select.def.
func SetFighterHomeStage(projectDir, fighterName, stageRelativePath string) error {
	selectDefPath := filepath.Join(projectDir, "data", "select.def")
	if _, err := os.Stat(selectDefPath); os.IsNotExist(err) {
		return fmt.Errorf("select.def not found in project")
	}

	data, err := os.ReadFile(selectDefPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	inCharacters := false

	fighterClean := strings.ToLower(strings.TrimSpace(fighterName))
	stageClean := filepath.ToSlash(strings.TrimSpace(stageRelativePath))

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "[") {
			inCharacters = strings.HasPrefix(lower, "[characters]")
			newLines = append(newLines, line)
			continue
		}

		if inCharacters && trimmed != "" && !strings.HasPrefix(trimmed, ";") {
			parts := strings.Split(trimmed, ",")
			charName := strings.ToLower(strings.TrimSpace(parts[0]))
			if charName == fighterClean {
				// Found the character line, update stage parameter
				if stageClean != "" {
					// Format: charname, stages/stage.def
					var opts []string
					opts = append(opts, parts[0])
					opts = append(opts, stageClean)
					// Preserve any trailing params (music, order, includestage)
					for i := 2; i < len(parts); i++ {
						opts = append(opts, strings.TrimSpace(parts[i]))
					}
					newLines = append(newLines, strings.Join(opts, ", "))
				} else {
					// Clear stage assignment
					var opts []string
					opts = append(opts, parts[0])
					for i := 2; i < len(parts); i++ {
						opts = append(opts, strings.TrimSpace(parts[i]))
					}
					newLines = append(newLines, strings.Join(opts, ", "))
				}
				continue
			}
		}

		newLines = append(newLines, line)
	}

	return os.WriteFile(selectDefPath, []byte(strings.Join(newLines, "\n")), 0644)
}

// DeleteProjectStage removes a stage from disk and unregisters it from select.def.
func DeleteProjectStage(projectDir, stageRelativePath string) error {
	fullPath := filepath.Join(projectDir, filepath.FromSlash(stageRelativePath))
	_ = ToggleStageExtraStage(projectDir, stageRelativePath, false)

	// If stage has its own folder, remove folder if it only contains stage files, or remove .def & .sff
	parent := filepath.Dir(fullPath)
	if filepath.Base(parent) != "stages" {
		_ = os.RemoveAll(parent)
	} else {
		_ = os.Remove(fullPath)
		sffPath := strings.TrimSuffix(fullPath, filepath.Ext(fullPath)) + ".sff"
		_ = os.Remove(sffPath)
	}
	return nil
}

type stageDefMeta struct {
	DisplayName  string
	Author       string
	MugenVersion string
	SpriteFile   string
	SoundFile    string
	XScale       float64
	YScale       float64
	ZOffset      float64
}

func parseStageDef(defPath string) *stageDefMeta {
	file, err := os.Open(defPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	meta := &stageDefMeta{
		DisplayName:  strings.TrimSuffix(filepath.Base(defPath), filepath.Ext(defPath)),
		Author:       "Unknown",
		MugenVersion: "1.0",
		XScale:       1.0,
		YScale:       1.0,
	}

	scanner := bufio.NewScanner(file)
	currentSec := ""
	sectionRe := regexp.MustCompile(`^\s*\[\s*([^\]]+)\s*\]`)
	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if m := sectionRe.FindStringSubmatch(trimmed); len(m) > 1 {
			currentSec = strings.ToLower(strings.TrimSpace(m[1]))
			continue
		}

		if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
			k := strings.ToLower(strings.TrimSpace(m[1]))
			v := cleanVal(m[2])

			switch currentSec {
			case "info":
				switch k {
				case "name":
					if meta.DisplayName == "" || meta.DisplayName == strings.TrimSuffix(filepath.Base(defPath), filepath.Ext(defPath)) {
						meta.DisplayName = v
					}
				case "displayname":
					meta.DisplayName = v
				case "author":
					meta.Author = v
				case "mugenversion":
					meta.MugenVersion = v
				}
			case "stageinfo":
				switch k {
				case "zoffset":
					if val, pErr := strconv.ParseFloat(v, 64); pErr == nil {
						meta.ZOffset = val
					}
				case "xscale":
					if val, pErr := strconv.ParseFloat(v, 64); pErr == nil {
						meta.XScale = val
					}
				case "yscale":
					if val, pErr := strconv.ParseFloat(v, 64); pErr == nil {
						meta.YScale = val
					}
				}
			case "bgdef":
				switch k {
				case "spr", "sprite", "sff":
					meta.SpriteFile = filepath.ToSlash(v)
				case "bgmusic", "music":
					meta.SoundFile = filepath.ToSlash(v)
				}
			}
		}
	}

	return meta
}

func parseSelectDefStageMappings(selectDefPath string) (map[string]bool, map[string][]string) {
	extraMap := make(map[string]bool)
	charStageMap := make(map[string][]string)

	file, err := os.Open(selectDefPath)
	if err != nil {
		return extraMap, charStageMap
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inExtraStages := false
	inCharacters := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "[") {
			inExtraStages = strings.HasPrefix(lower, "[extrastages]")
			inCharacters = strings.HasPrefix(lower, "[characters]")
			continue
		}

		if inExtraStages {
			extraMap[normalizeStageKey(line)] = true
		} else if inCharacters {
			parts := strings.Split(line, ",")
			if len(parts) >= 2 {
				charName := strings.TrimSpace(parts[0])
				stageName := normalizeStageKey(parts[1])
				if stageName != "" && !strings.Contains(stageName, "=") {
					charStageMap[stageName] = append(charStageMap[stageName], charName)
				}
			}
		}
	}

	return extraMap, charStageMap
}

func normalizeStageKey(p string) string {
	cleaned := filepath.ToSlash(strings.TrimSpace(p))
	cleaned = strings.TrimPrefix(cleaned, "stages/")
	cleaned = strings.TrimSuffix(cleaned, ".def")
	return strings.ToLower(cleaned)
}

func cleanVal(v string) string {
	if idx := strings.Index(v, ";"); idx != -1 {
		v = v[:idx]
	}
	if idx := strings.Index(v, "#"); idx != -1 {
		v = v[:idx]
	}
	return strings.Trim(strings.TrimSpace(v), "\"'")
}
