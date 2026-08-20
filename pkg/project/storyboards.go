package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"ikemen-studio/pkg/vault"
)

// ProjectStoryboardInfo represents a cinematic cutscene storyboard (.def) with its scenes and system assignments.
type ProjectStoryboardInfo struct {
	RelativePath      string   `json:"relative_path"`
	DisplayName       string   `json:"display_name"`
	SceneCount        int      `json:"scene_count"`
	BgmPath           string   `json:"bgm_path"`
	SpriteFile        string   `json:"sprite_file"`
	AssignedSlots     []string `json:"assigned_slots"`
	PreviewBase64     string   `json:"preview_base64"`
	IsLinkedFromVault bool     `json:"is_linked_from_vault"`
}

// GetProjectStoryboards scans data/ and subdirectories for storyboard .def files containing [SceneDef] or [Scene].
func GetProjectStoryboards(projectDir string) ([]ProjectStoryboardInfo, error) {
	dataDir := filepath.Join(projectDir, "data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return []ProjectStoryboardInfo{}, nil
	}

	storyboardSlotsMap := parseSystemDefStoryboards(projectDir)

	var storyboards []ProjectStoryboardInfo

	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) == ".def" {
			meta := parseStoryboardDef(path)
			if meta != nil && meta.IsStoryboard {
				rel, _ := filepath.Rel(projectDir, path)
				rel = filepath.ToSlash(rel)

				isSymlink := false
				if fi, lErr := os.Lstat(path); lErr == nil && (fi.Mode()&os.ModeSymlink != 0) {
					isSymlink = true
				}

				baseName := filepath.Base(path)
				normKey := normalizeStoryboardKey(rel)

				slots := storyboardSlotsMap[normKey]
				if len(slots) == 0 {
					slots = storyboardSlotsMap[normalizeStoryboardKey(baseName)]
				}

				previewBase64 := ""
				if meta.SpriteFile != "" {
					fullSff := filepath.Join(filepath.Dir(path), meta.SpriteFile)
					if _, sErr := os.Stat(fullSff); sErr == nil {
						_, b64, _ := vault.ExtractAndCachePortrait(filepath.Dir(path), filepath.Base(filepath.Dir(path)), fullSff)
						previewBase64 = b64
					}
				}

				sb := ProjectStoryboardInfo{
					RelativePath:      rel,
					DisplayName:       meta.DisplayName,
					SceneCount:        meta.SceneCount,
					BgmPath:           meta.BgmPath,
					SpriteFile:        meta.SpriteFile,
					AssignedSlots:     slots,
					PreviewBase64:     previewBase64,
					IsLinkedFromVault: isSymlink,
				}

				storyboards = append(storyboards, sb)
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan storyboards: %w", err)
	}

	return storyboards, nil
}

// SetSystemStoryboard maps a storyboard to intro, ending, or credits in system.def.
// storyType can be "intro", "ending", "credits", "gameover".
func SetSystemStoryboard(projectDir, storyType, storyboardRelativePath string) error {
	systemDefPath := findSystemDef(projectDir)
	if _, err := os.Stat(systemDefPath); os.IsNotExist(err) {
		return fmt.Errorf("system.def not found at %s", systemDefPath)
	}

	data, err := os.ReadFile(systemDefPath)
	if err != nil {
		return err
	}

	cleanType := strings.ToLower(strings.TrimSpace(storyType))
	slotKey := cleanType + ".storyboard"
	cleanPath := filepath.ToSlash(strings.TrimSpace(storyboardRelativePath))

	lines := strings.Split(string(data), "\n")
	var newLines []string
	inFiles := false
	slotUpdated := false
	hasFilesBlock := false

	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "[") {
			if inFiles && !slotUpdated {
				newLines = append(newLines, fmt.Sprintf("%s = %s", slotKey, cleanPath))
				slotUpdated = true
			}
			inFiles = strings.HasPrefix(lower, "[files]")
			if inFiles {
				hasFilesBlock = true
			}
			newLines = append(newLines, line)
			continue
		}

		if inFiles && !slotUpdated {
			if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
				k := strings.ToLower(strings.TrimSpace(m[1]))
				if k == slotKey {
					newLines = append(newLines, fmt.Sprintf("%s = %s", slotKey, cleanPath))
					slotUpdated = true
					continue
				}
			}
		}

		newLines = append(newLines, line)
	}

	if !slotUpdated {
		if hasFilesBlock {
			newLines = append(newLines, fmt.Sprintf("%s = %s", slotKey, cleanPath))
		} else {
			newLines = append(newLines, "\n[Files]", fmt.Sprintf("%s = %s", slotKey, cleanPath))
		}
	}

	return os.WriteFile(systemDefPath, []byte(strings.Join(newLines, "\n")), 0644)
}

type storyboardDefMeta struct {
	DisplayName  string
	SceneCount   int
	BgmPath      string
	SpriteFile   string
	IsStoryboard bool
}

func parseStoryboardDef(defPath string) *storyboardDefMeta {
	file, err := os.Open(defPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	meta := &storyboardDefMeta{
		DisplayName: strings.TrimSuffix(filepath.Base(defPath), filepath.Ext(defPath)),
	}

	scanner := bufio.NewScanner(file)
	currentSec := ""
	sectionRe := regexp.MustCompile(`^\s*\[\s*([^\]]+)\s*\]`)
	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	hasSceneDef := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if m := sectionRe.FindStringSubmatch(trimmed); len(m) > 1 {
			currentSec = strings.ToLower(strings.TrimSpace(m[1]))
			if strings.HasPrefix(currentSec, "scenedef") {
				hasSceneDef = true
				meta.IsStoryboard = true
			}
			if strings.HasPrefix(currentSec, "scene ") || currentSec == "scene" {
				meta.SceneCount++
				meta.IsStoryboard = true
			}
			continue
		}

		if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
			k := strings.ToLower(strings.TrimSpace(m[1]))
			v := cleanVal(m[2])

			switch currentSec {
			case "scenedef":
				switch k {
				case "spr", "sprite", "sff":
					meta.SpriteFile = filepath.ToSlash(v)
				}
			}
			if strings.HasPrefix(currentSec, "scene") {
				if k == "bgm" || k == "music" {
					if meta.BgmPath == "" {
						meta.BgmPath = filepath.ToSlash(v)
					}
				}
			}
		}
	}

	if hasSceneDef || meta.SceneCount > 0 {
		meta.IsStoryboard = true
	}

	return meta
}

func parseSystemDefStoryboards(projectDir string) map[string][]string {
	res := make(map[string][]string)
	systemDefPath := findSystemDef(projectDir)

	file, err := os.Open(systemDefPath)
	if err != nil {
		return res
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFiles := false
	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "[") {
			inFiles = strings.HasPrefix(lower, "[files]")
			continue
		}

		if inFiles {
			if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
				k := strings.ToLower(strings.TrimSpace(m[1]))
				v := cleanVal(m[2])
				if strings.Contains(k, "storyboard") && v != "" {
					label := k
					switch k {
					case "intro.storyboard":
						label = "Opening Intro"
					case "ending.storyboard":
						label = "Story Ending"
					case "credits.storyboard":
						label = "Credits Roll"
					case "gameover.storyboard":
						label = "Game Over"
					}
					norm := normalizeStoryboardKey(v)
					res[norm] = append(res[norm], label)
				}
			}
		}
	}

	return res
}

func normalizeStoryboardKey(p string) string {
	cleaned := filepath.ToSlash(strings.TrimSpace(p))
	cleaned = strings.TrimPrefix(cleaned, "data/")
	return strings.ToLower(cleaned)
}
