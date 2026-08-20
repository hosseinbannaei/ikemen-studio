package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectFontInfo represents a font asset (.fnt, .def, .ttf) with its current slot mappings in system.def / fight.def.
type ProjectFontInfo struct {
	RelativePath       string   `json:"relative_path"`
	FileName           string   `json:"file_name"`
	FontType           string   `json:"font_type"`
	SizeBytes          int64    `json:"size_bytes"`
	SystemSlotMappings []string `json:"system_slot_mappings"`
	IsLinkedFromVault  bool     `json:"is_linked_from_vault"`
}

// GetProjectFonts scans font/ and subfolders and inspects font slot assignments in system.def and fight.def.
func GetProjectFonts(projectDir string) ([]ProjectFontInfo, error) {
	fontDir := filepath.Join(projectDir, "font")
	if _, err := os.Stat(fontDir); os.IsNotExist(err) {
		return []ProjectFontInfo{}, nil
	}

	// Parse font mappings from system.def and fight.def
	fontSlotMap := parseAllFontMappings(projectDir)

	var fonts []ProjectFontInfo
	fontExts := map[string]bool{
		".fnt": true, ".def": true, ".ttf": true, ".otf": true, ".woff2": true,
	}

	err := filepath.Walk(fontDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if fontExts[ext] {
			rel, _ := filepath.Rel(projectDir, path)
			rel = filepath.ToSlash(rel)

			isSymlink := false
			if fi, lErr := os.Lstat(path); lErr == nil && (fi.Mode()&os.ModeSymlink != 0) {
				isSymlink = true
			}

			baseName := filepath.Base(path)
			normKey := normalizeFontKey(rel)

			slots := fontSlotMap[normKey]
			if len(slots) == 0 {
				slots = fontSlotMap[normalizeFontKey(baseName)]
			}

			fontType := "Bitmap Font"
			switch ext {
			case ".ttf", ".otf", ".woff2":
				fontType = "TrueType (TTF/OTF)"
			case ".def":
				fontType = "Ikemen Font Def"
			case ".fnt":
				fontType = "MUGEN FNT"
			}

			font := ProjectFontInfo{
				RelativePath:       rel,
				FileName:           baseName,
				FontType:           fontType,
				SizeBytes:          info.Size(),
				SystemSlotMappings: slots,
				IsLinkedFromVault:  isSymlink,
			}

			fonts = append(fonts, font)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan font directory: %w", err)
	}

	return fonts, nil
}

// SetSystemFontMapping maps a font to a slot (e.g. "font1", "font2") in system.def or fight.def.
// targetDef can be "system" or "fight".
func SetSystemFontMapping(projectDir, targetDef, fontSlot, fontRelativePath string) error {
	var targetFile string
	if targetDef == "fight" {
		activeFight := getActiveLifebarFromSystemDef(projectDir)
		targetFile = filepath.Join(projectDir, filepath.FromSlash(activeFight))
	} else {
		targetFile = findSystemDef(projectDir)
	}

	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		return fmt.Errorf("target def file not found at %s", targetFile)
	}

	data, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	cleanSlot := strings.ToLower(strings.TrimSpace(fontSlot))
	cleanPath := filepath.ToSlash(strings.TrimSpace(fontRelativePath))

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
				newLines = append(newLines, fmt.Sprintf("%s = %s", cleanSlot, cleanPath))
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
				if k == cleanSlot {
					newLines = append(newLines, fmt.Sprintf("%s = %s", cleanSlot, cleanPath))
					slotUpdated = true
					continue
				}
			}
		}

		newLines = append(newLines, line)
	}

	if !slotUpdated {
		if hasFilesBlock {
			newLines = append(newLines, fmt.Sprintf("%s = %s", cleanSlot, cleanPath))
		} else {
			newLines = append(newLines, "\n[Files]", fmt.Sprintf("%s = %s", cleanSlot, cleanPath))
		}
	}

	return os.WriteFile(targetFile, []byte(strings.Join(newLines, "\n")), 0644)
}

func parseAllFontMappings(projectDir string) map[string][]string {
	res := make(map[string][]string)

	systemDefPath := findSystemDef(projectDir)
	parseFontsFromDef(systemDefPath, "System", res)

	activeFight := getActiveLifebarFromSystemDef(projectDir)
	fightPath := filepath.Join(projectDir, filepath.FromSlash(activeFight))
	parseFontsFromDef(fightPath, "Fight HUD", res)

	return res
}

func parseFontsFromDef(defPath, prefix string, res map[string][]string) {
	file, err := os.Open(defPath)
	if err != nil {
		return
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
				if strings.HasPrefix(k, "font") && v != "" {
					norm := normalizeFontKey(v)
					slotLabel := fmt.Sprintf("%s %s", prefix, k)
					res[norm] = append(res[norm], slotLabel)
				}
			}
		}
	}
}

func normalizeFontKey(p string) string {
	cleaned := filepath.ToSlash(strings.TrimSpace(p))
	cleaned = strings.TrimPrefix(cleaned, "font/")
	return strings.ToLower(cleaned)
}
