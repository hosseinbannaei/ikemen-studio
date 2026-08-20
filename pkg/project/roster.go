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

// RosterCharacterSlot represents a single slot in the character select screen grid.
type RosterCharacterSlot struct {
	Index           int    `json:"index"`
	Type            string `json:"type"` // "character", "randomselect", "empty", "disabled"
	Character       string `json:"character"`
	DisplayName     string `json:"display_name"`
	Author          string `json:"author"`
	PortraitBase64  string `json:"portrait_base64"`
	HomeStage       string `json:"home_stage"`
	Music           string `json:"music"`
	Order           int    `json:"order"`
	IncludeInArcade bool   `json:"include_in_arcade"`
	RawLine         string `json:"raw_line"`
}

// RosterGridInfo holds the layout dimensions of the select screen.
type RosterGridInfo struct {
	Rows           int  `json:"rows"`
	Columns        int  `json:"columns"`
	Wrapping       bool `json:"wrapping"`
	ShowEmptyBoxes bool `json:"show_empty_boxes"`
}

// RosterAvailableCharacter represents a character found on disk available for placing.
type RosterAvailableCharacter struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Author         string `json:"author"`
	PortraitBase64 string `json:"portrait_base64"`
	IsLinked       bool   `json:"is_linked"`
}

// ProjectRoster contains the full structured roster and available assets.
type ProjectRoster struct {
	Grid                RosterGridInfo             `json:"grid"`
	Slots               []RosterCharacterSlot      `json:"slots"`
	ExtraStages         []string                   `json:"extra_stages"`
	AvailableCharacters []RosterAvailableCharacter `json:"available_characters"`
	AvailableStages     []string                   `json:"available_stages"`
}

// GetProjectRoster reads system.def, select.def, chars/, and stages/ to build the full roster.
func GetProjectRoster(projectDir string) (*ProjectRoster, error) {
	systemDefPath := findSystemDef(projectDir)
	gridInfo := ParseSystemDefGrid(systemDefPath)

	selectDefPath := filepath.Join(projectDir, "data", "select.def")
	slots, extraStages, err := ParseSelectDef(selectDefPath, projectDir)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to parse select.def: %w", err)
	}

	availChars := ScanAvailableCharacters(projectDir)
	availStages := ScanAvailableStages(projectDir)

	return &ProjectRoster{
		Grid:                gridInfo,
		Slots:               slots,
		ExtraStages:         extraStages,
		AvailableCharacters: availChars,
		AvailableStages:     availStages,
	}, nil
}

// SaveProjectRoster writes the updated roster slots and extra stages to data/select.def.
func SaveProjectRoster(projectDir string, roster ProjectRoster) error {
	selectDefPath := filepath.Join(projectDir, "data", "select.def")
	if err := os.MkdirAll(filepath.Dir(selectDefPath), 0755); err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString(";---------------------------------------------------------------------\n")
	sb.WriteString("; Characters and Stages Configuration (Managed by Ikemen Studio)\n")
	sb.WriteString(";---------------------------------------------------------------------\n\n")

	sb.WriteString("[Characters]\n")
	for _, slot := range roster.Slots {
		line := formatSlotLine(slot)
		if line != "" {
			sb.WriteString(line + "\n")
		}
	}

	sb.WriteString("\n[ExtraStages]\n")
	for _, stage := range roster.ExtraStages {
		st := strings.TrimSpace(stage)
		if st != "" {
			if !strings.HasPrefix(strings.ToLower(st), "stages/") && !strings.Contains(st, "/") {
				st = "stages/" + st
			}
			if !strings.HasSuffix(strings.ToLower(st), ".def") {
				st = st + ".def"
			}
			sb.WriteString(st + "\n")
		}
	}

	sb.WriteString("\n[Options]\n")
	sb.WriteString("; Arcade mode matches per order\n")
	sb.WriteString("arcade.maxmatches = 6,1,1,0,0,0,0,0,0,0\n")
	sb.WriteString("team.maxmatches = 4,1,1,0,0,0,0,0,0,0\n")

	return os.WriteFile(selectDefPath, []byte(sb.String()), 0644)
}

func formatSlotLine(slot RosterCharacterSlot) string {
	switch slot.Type {
	case "empty":
		return "empty"
	case "randomselect":
		return "randomselect"
	case "disabled":
		if slot.Character != "" {
			return "; " + slot.Character + " (Disabled)"
		}
		return ""
	case "character":
		if slot.Character == "" {
			return "empty"
		}
		var parts []string
		parts = append(parts, slot.Character)

		if slot.HomeStage != "" {
			stageVal := slot.HomeStage
			if !strings.HasPrefix(strings.ToLower(stageVal), "stages/") && !strings.Contains(stageVal, "/") {
				stageVal = "stages/" + stageVal
			}
			if !strings.HasSuffix(strings.ToLower(stageVal), ".def") {
				stageVal = stageVal + ".def"
			}
			parts = append(parts, stageVal)
		}

		if slot.Music != "" {
			parts = append(parts, fmt.Sprintf("music=%s", slot.Music))
		}

		if slot.Order > 0 {
			parts = append(parts, fmt.Sprintf("order=%d", slot.Order))
		}

		if !slot.IncludeInArcade && slot.Order == 0 {
			parts = append(parts, "includestage=0")
		}

		return strings.Join(parts, ", ")
	}
	return ""
}

// ParseSystemDefGrid reads [Select Info] rows and columns from system.def.
func ParseSystemDefGrid(systemDefPath string) RosterGridInfo {
	grid := RosterGridInfo{
		Rows:           2,
		Columns:        5,
		Wrapping:       true,
		ShowEmptyBoxes: true,
	}

	if systemDefPath == "" {
		return grid
	}

	data, err := os.ReadFile(systemDefPath)
	if err != nil {
		return grid
	}

	inSelectInfo := false
	scanner := bufio.NewScanner(strings.NewReader(string(data)))

	reRows := regexp.MustCompile(`(?i)^\s*rows\s*=\s*(\d+)`)
	reCols := regexp.MustCompile(`(?i)^\s*columns\s*=\s*(\d+)`)
	reWrap := regexp.MustCompile(`(?i)^\s*wrapping\s*=\s*(\d+)`)
	reShow := regexp.MustCompile(`(?i)^\s*showemptyboxes\s*=\s*(\d+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			sec := strings.ToLower(line)
			inSelectInfo = strings.Contains(sec, "select info")
			continue
		}

		if inSelectInfo {
			if m := reRows.FindStringSubmatch(line); len(m) > 1 {
				if v, e := strconv.Atoi(m[1]); e == nil && v > 0 {
					grid.Rows = v
				}
			} else if m := reCols.FindStringSubmatch(line); len(m) > 1 {
				if v, e := strconv.Atoi(m[1]); e == nil && v > 0 {
					grid.Columns = v
				}
			} else if m := reWrap.FindStringSubmatch(line); len(m) > 1 {
				grid.Wrapping = m[1] == "1"
			} else if m := reShow.FindStringSubmatch(line); len(m) > 1 {
				grid.ShowEmptyBoxes = m[1] == "1"
			}
		}
	}

	return grid
}

// ParseSelectDef extracts character slots and extra stages from select.def.
func ParseSelectDef(selectDefPath, projectDir string) ([]RosterCharacterSlot, []string, error) {
	data, err := os.ReadFile(selectDefPath)
	if err != nil {
		return nil, nil, err
	}

	var slots []RosterCharacterSlot
	var extraStages []string

	currentSection := ""
	reachedCharactersList := false
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	slotIdx := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "[") {
			currentSection = strings.ToLower(trimmed)
			continue
		}

		if strings.HasPrefix(currentSection, "[characters]") {
			// Check if tutorial comment or section boundary
			lowerTrimmed := strings.ToLower(trimmed)
			if strings.Contains(lowerTrimmed, "insert your characters below") || strings.Contains(lowerTrimmed, "characters below") {
				reachedCharactersList = true
				continue
			}

			// Check if disabled character (commented out line)
			if strings.HasPrefix(trimmed, ";") {
				if !reachedCharactersList {
					// Still in tutorial header; skip example lines
					continue
				}

				commentText := strings.TrimPrefix(trimmed, ";")
				commentText = strings.TrimSpace(commentText)
				// Check if it matches a character name pattern
				if commentText != "" && !strings.HasPrefix(commentText, "-") && !strings.HasPrefix(commentText, "=") {
					firstToken := strings.Fields(commentText)[0]
					firstToken = strings.Trim(firstToken, ",;")
					if isCharInDir(projectDir, firstToken) {
						charName, dName, author, b64 := resolveCharInfo(projectDir, firstToken)
						slots = append(slots, RosterCharacterSlot{
							Index:          slotIdx,
							Type:           "disabled",
							Character:      charName,
							DisplayName:    dName,
							Author:         author,
							PortraitBase64: b64,
							RawLine:        line,
						})
						slotIdx++
					}
				}
				continue
			}

			// First uncommented character line automatically activates character list mode
			reachedCharactersList = true

			// Clean inline comments
			cleanLine := trimmed
			if idx := strings.Index(cleanLine, ";"); idx != -1 {
				cleanLine = strings.TrimSpace(cleanLine[:idx])
			}
			if cleanLine == "" {
				continue
			}

			lowerClean := strings.ToLower(cleanLine)
			if lowerClean == "empty" {
				slots = append(slots, RosterCharacterSlot{
					Index:   slotIdx,
					Type:    "empty",
					RawLine: line,
				})
				slotIdx++
				continue
			}

			if lowerClean == "randomselect" {
				slots = append(slots, RosterCharacterSlot{
					Index:          slotIdx,
					Type:           "randomselect",
					Character:      "randomselect",
					DisplayName:    "Random Select",
					RawLine:        line,
				})
				slotIdx++
				continue
			}

			// Parse character parameters: name, stage, music=..., order=...
			tokens := strings.Split(cleanLine, ",")
			charName := strings.TrimSpace(tokens[0])
			homeStage := ""
			music := ""
			order := 0
			includeInArcade := true

			for i := 1; i < len(tokens); i++ {
				param := strings.TrimSpace(tokens[i])
				lowerParam := strings.ToLower(param)
				if strings.HasPrefix(lowerParam, "music=") {
					music = strings.TrimPrefix(param, "music=")
					music = strings.TrimPrefix(music, "MUSIC=")
				} else if strings.HasPrefix(lowerParam, "order=") {
					orderStr := strings.TrimPrefix(lowerParam, "order=")
					if o, e := strconv.Atoi(orderStr); e == nil {
						order = o
					}
				} else if strings.HasPrefix(lowerParam, "includestage=") {
					if strings.TrimPrefix(lowerParam, "includestage=") == "0" {
						includeInArcade = false
					}
				} else if strings.HasSuffix(lowerParam, ".def") || strings.HasPrefix(lowerParam, "stages/") {
					homeStage = param
				}
			}

			cName, dName, author, b64 := resolveCharInfo(projectDir, charName)
			slots = append(slots, RosterCharacterSlot{
				Index:           slotIdx,
				Type:            "character",
				Character:       cName,
				DisplayName:     dName,
				Author:          author,
				PortraitBase64:  b64,
				HomeStage:       homeStage,
				Music:           music,
				Order:           order,
				IncludeInArcade: includeInArcade,
				RawLine:         line,
			})
			slotIdx++
		} else if strings.HasPrefix(currentSection, "[extrastages]") {
			if !strings.HasPrefix(trimmed, ";") {
				cleanLine := trimmed
				if idx := strings.Index(cleanLine, ";"); idx != -1 {
					cleanLine = strings.TrimSpace(cleanLine[:idx])
				}
				if cleanLine != "" {
					extraStages = append(extraStages, cleanLine)
				}
			}
		}
	}

	return slots, extraStages, nil
}

// ScanAvailableCharacters finds all character directories in chars/.
func ScanAvailableCharacters(projectDir string) []RosterAvailableCharacter {
	var list []RosterAvailableCharacter
	charsDir := filepath.Join(projectDir, "chars")
	entries, err := os.ReadDir(charsDir)
	if err != nil {
		return list
	}

	for _, e := range entries {
		if !e.IsDir() && e.Type()&os.ModeSymlink == 0 {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") || strings.EqualFold(e.Name(), "randomselect") {
			continue
		}

		cName, dName, author, b64 := resolveCharInfo(projectDir, e.Name())
		isLinked := false
		if fi, err := os.Lstat(filepath.Join(charsDir, e.Name())); err == nil {
			isLinked = fi.Mode()&os.ModeSymlink != 0
		}

		list = append(list, RosterAvailableCharacter{
			Name:           cName,
			DisplayName:    dName,
			Author:         author,
			PortraitBase64: b64,
			IsLinked:       isLinked,
		})
	}

	return list
}

// ScanAvailableStages finds all stage .def files in stages/.
func ScanAvailableStages(projectDir string) []string {
	var list []string
	stagesDir := filepath.Join(projectDir, "stages")
	_ = filepath.Walk(stagesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".def") {
			rel, _ := filepath.Rel(projectDir, path)
			list = append(list, filepath.ToSlash(rel))
		}
		return nil
	})
	return list
}

func resolveCharInfo(projectDir, charName string) (string, string, string, string) {
	cleanName := strings.TrimSpace(charName)
	cleanName = strings.TrimPrefix(cleanName, "chars/")
	if idx := strings.Index(cleanName, "/"); idx != -1 {
		cleanName = cleanName[:idx]
	}

	charDir := filepath.Join(projectDir, "chars", cleanName)
	defFile := filepath.Join(charDir, cleanName+".def")

	if _, err := os.Stat(defFile); os.IsNotExist(err) {
		// Look for any .def in charDir
		_ = filepath.Walk(charDir, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".def") {
				defFile = p
				return ioEOF()
			}
			return nil
		})
	}

	displayName := cleanName
	author := "Unknown"
	var b64Portrait string

	if defInfo, err := vault.ParseDefFile(defFile); err == nil && defInfo != nil {
		if defInfo.DisplayName != "" {
			displayName = defInfo.DisplayName
		}
		if defInfo.Author != "" {
			author = defInfo.Author
		}

		sffPath := ""
		if defInfo.SpriteFile != "" {
			sffPath = filepath.Join(charDir, defInfo.SpriteFile)
		}
		if sffPath == "" || !fileExists(sffPath) {
			_ = filepath.Walk(charDir, func(p string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".sff") {
					sffPath = p
					return ioEOF()
				}
				return nil
			})
		}

		if sffPath != "" {
			_, b64, _ := vault.ExtractAndCachePortrait(projectDir, "chars/"+cleanName, sffPath)
			b64Portrait = b64
		}
	}

	// Disambiguate duplicate character display names like kfm variants
	if strings.EqualFold(displayName, "Kung Fu Man") && cleanName != "kfm" {
		displayName = fmt.Sprintf("Kung Fu Man (%s)", cleanName)
	}

	return cleanName, displayName, author, b64Portrait
}

func isCharInDir(projectDir, name string) bool {
	cleanName := strings.TrimSpace(name)
	cleanName = strings.TrimPrefix(cleanName, "chars/")
	if idx := strings.Index(cleanName, "/"); idx != -1 {
		cleanName = cleanName[:idx]
	}
	_, err := os.Stat(filepath.Join(projectDir, "chars", cleanName))
	return err == nil
}

func findSystemDef(projectDir string) string {
	candidates := []string{
		filepath.Join(projectDir, "data", "system.def"),
		filepath.Join(projectDir, "data", "motif", "system.def"),
	}
	for _, c := range candidates {
		if fileExists(c) {
			return c
		}
	}
	return ""
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func ioEOF() error {
	return os.ErrExist // using sentinel error to halt walk
}
