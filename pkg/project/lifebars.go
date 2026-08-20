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

// ProjectLifebarInfo represents detailed metadata about an installed lifebar / fight HUD package.
type ProjectLifebarInfo struct {
	Key               string `json:"key"`
	DisplayName       string `json:"display_name"`
	Author            string `json:"author"`
	Version           string `json:"version"`
	IsActive          bool   `json:"is_active"`
	SpriteFile        string `json:"sprite_file"`
	SoundFile         string `json:"sound_file"`
	FontCount         int    `json:"font_count"`
	PreviewBase64     string `json:"preview_base64"`
	IsLinkedFromVault bool   `json:"is_linked_from_vault"`
}

// GetProjectLifebars scans data/ and subdirectories for fight.def files and detects the active lifebar.
func GetProjectLifebars(projectDir string) ([]ProjectLifebarInfo, error) {
	dataDir := filepath.Join(projectDir, "data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return []ProjectLifebarInfo{}, nil
	}

	// 1. Get currently active fight.def from the active system.def
	activeFightPath := getActiveLifebarFromSystemDef(projectDir)

	var lifebars []ProjectLifebarInfo

	// 2. Walk data/ to discover all fight.def or lifebar.def files
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		baseName := strings.ToLower(filepath.Base(path))
		if baseName == "fight.def" || baseName == "lifebar.def" {
			rel, _ := filepath.Rel(projectDir, path)
			rel = filepath.ToSlash(rel)

			lifebarMeta := parseLifebarDef(path)
			if lifebarMeta == nil {
				return nil
			}

			// Active check
			isActive := normalizeLifebarKey(rel) == normalizeLifebarKey(activeFightPath) ||
				normalizeLifebarKey(filepath.Base(rel)) == normalizeLifebarKey(activeFightPath)

			// Symlink check
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

			// Portrait / HUD sprite preview extraction
			previewBase64 := ""
			if lifebarMeta.SpriteFile != "" {
				fullSff := filepath.Join(filepath.Dir(path), lifebarMeta.SpriteFile)
				if _, sErr := os.Stat(fullSff); sErr == nil {
					_, b64, _ := vault.ExtractAndCachePortrait(filepath.Dir(path), filepath.Base(filepath.Dir(path)), fullSff)
					previewBase64 = b64
				}
			}

			lifebar := ProjectLifebarInfo{
				Key:               rel,
				DisplayName:       lifebarMeta.DisplayName,
				Author:            lifebarMeta.Author,
				Version:           lifebarMeta.MugenVersion,
				IsActive:          isActive,
				SpriteFile:        lifebarMeta.SpriteFile,
				SoundFile:         lifebarMeta.SoundFile,
				FontCount:         lifebarMeta.FontCount,
				PreviewBase64:     previewBase64,
				IsLinkedFromVault: isSymlink,
			}

			lifebars = append(lifebars, lifebar)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan lifebars: %w", err)
	}

	return lifebars, nil
}

// SetActiveLifebar updates [Files] fight = ... in the active system.def.
func SetActiveLifebar(projectDir, lifebarRelativePath string) error {
	cleanPath := filepath.ToSlash(strings.TrimSpace(lifebarRelativePath))
	fullPath := filepath.Join(projectDir, filepath.FromSlash(cleanPath))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("fight.def not found at %s", cleanPath)
	}

	systemDefPath := findSystemDef(projectDir)
	if _, err := os.Stat(systemDefPath); os.IsNotExist(err) {
		return fmt.Errorf("system.def not found at %s", systemDefPath)
	}

	data, err := os.ReadFile(systemDefPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	var newLines []string
	inFiles := false
	fightUpdated := false
	hasFilesBlock := false

	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "[") {
			if inFiles && !fightUpdated {
				newLines = append(newLines, fmt.Sprintf("fight = %s", cleanPath))
				fightUpdated = true
			}
			inFiles = strings.HasPrefix(lower, "[files]")
			if inFiles {
				hasFilesBlock = true
			}
			newLines = append(newLines, line)
			continue
		}

		if inFiles && !fightUpdated {
			if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
				k := strings.ToLower(strings.TrimSpace(m[1]))
				if k == "fight" {
					newLines = append(newLines, fmt.Sprintf("fight = %s", cleanPath))
					fightUpdated = true
					continue
				}
			}
		}

		newLines = append(newLines, line)
	}

	if !fightUpdated {
		if hasFilesBlock {
			newLines = append(newLines, fmt.Sprintf("fight = %s", cleanPath))
		} else {
			newLines = append(newLines, "\n[Files]", fmt.Sprintf("fight = %s", cleanPath))
		}
	}

	return os.WriteFile(systemDefPath, []byte(strings.Join(newLines, "\n")), 0644)
}

type lifebarDefMeta struct {
	DisplayName  string
	Author       string
	MugenVersion string
	SpriteFile   string
	SoundFile    string
	FontCount    int
}

func parseLifebarDef(defPath string) *lifebarDefMeta {
	file, err := os.Open(defPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	parentFolder := filepath.Base(filepath.Dir(defPath))
	if parentFolder == "data" {
		parentFolder = "Default Lifebars"
	}

	meta := &lifebarDefMeta{
		DisplayName:  parentFolder,
		Author:       "Elecbyte / Ikemen",
		MugenVersion: "1.0",
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
					if meta.DisplayName == "Default Lifebars" || meta.DisplayName == parentFolder {
						meta.DisplayName = v
					}
				case "displayname":
					meta.DisplayName = v
				case "author":
					meta.Author = v
				case "mugenversion", "ikemenversion":
					meta.MugenVersion = v
				}
			case "files":
				switch k {
				case "sff", "spr", "sprite":
					meta.SpriteFile = filepath.ToSlash(v)
				case "snd", "sound":
					meta.SoundFile = filepath.ToSlash(v)
				}
				if strings.HasPrefix(k, "font") {
					meta.FontCount++
				}
			}
		}
	}

	return meta
}

func getActiveLifebarFromSystemDef(projectDir string) string {
	systemDefPath := findSystemDef(projectDir)
	file, err := os.Open(systemDefPath)
	if err != nil {
		return "data/fight.def"
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
		if strings.HasPrefix(strings.ToLower(line), "[files]") {
			inFiles = true
			continue
		} else if strings.HasPrefix(line, "[") {
			inFiles = false
		}

		if inFiles {
			if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
				if strings.ToLower(strings.TrimSpace(m[1])) == "fight" {
					val := cleanVal(m[2])
					if val != "" {
						return filepath.ToSlash(val)
					}
				}
			}
		}
	}

	return "data/fight.def"
}

func normalizeLifebarKey(p string) string {
	cleaned := filepath.ToSlash(strings.TrimSpace(p))
	return strings.ToLower(cleaned)
}
