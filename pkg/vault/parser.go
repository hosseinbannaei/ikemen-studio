package vault

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	urlRegex     = regexp.MustCompile(`https?://[^\s<>"\']+`)
	sectionRegex = regexp.MustCompile(`^\s*\[\s*([^\]]+)\s*\]`)
	keyValRegex  = regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)
)

// DefInfo holds extracted metadata from a MUGEN / Ikemen .def file.
type DefInfo struct {
	Category      Category
	Name          string
	DisplayName   string
	Author        string
	VersionDate   string
	MugenVersion  string
	IkemenVersion string
	SpriteFile    string // Relative to .def file, normalized with /
	SoundFile     string // Relative to .def file, normalized with /
	AnimFile      string // Relative to .def file, normalized with /
	CmdFile       string // Relative to .def file, normalized with /
	CnsFile       string // Relative to .def file, normalized with /
	FoundURLs     []string
	Comments      []string
	IsStoryboard  bool // e.g. ending.def, intro.def (contains [SceneDef])
	IsLifebar     bool // e.g. fight.def, lifebar.def (contains [Lifebar] or [Round])
	IsFont        bool // e.g. font.def, name14.def (contains [Font] or .fnt)
	IsCommand     bool // e.g. command.def (contains only [Command])
	HasFilesBlock bool // true if [Files] section with cns/cmd/sprite exists
	IsValidAsset  bool // true if verified as a valid character, stage, motif, or sound
}

// ParseDefFile reads and extracts metadata from a character, stage, lifebar, or motif .def file.
func ParseDefFile(defPath string) (*DefInfo, error) {
	file, err := os.Open(defPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &DefInfo{
		FoundURLs: make([]string, 0),
		Comments:  make([]string, 0),
	}

	scanner := bufio.NewScanner(file)
	currentSection := ""
	hasStageInfo := false
	hasTitleInfo := false
	hasLifebarInfo := false
	hasSceneDef := false
	hasFiles := false
	hasCharInfo := false
	hasFontInfo := false
	hasCmdInfo := false

	baseName := strings.ToLower(filepath.Base(defPath))
	baseNoExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Preliminary name-based checks
	if baseName == "fight.def" || baseName == "lifebar.def" {
		hasLifebarInfo = true
	}
	if baseName == "intro.def" || baseName == "ending.def" || baseName == "credits.def" ||
		baseName == "introduction.def" || baseName == "cutscene.def" || baseName == "logo.def" {
		hasSceneDef = true
	}
	if strings.HasPrefix(baseName, "font") || strings.HasPrefix(baseName, "name14") || strings.HasPrefix(baseName, "f-") {
		hasFontInfo = true
	}
	if baseName == "command.def" || baseName == "cmd.def" {
		hasCmdInfo = true
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for URLs across all lines including comments
		if matches := urlRegex.FindAllString(line, -1); len(matches) > 0 {
			for _, u := range matches {
				u = strings.TrimRight(u, ".,;)\"'>")
				if !containsString(info.FoundURLs, u) {
					info.FoundURLs = append(info.FoundURLs, u)
				}
			}
		}

		if strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			commentText := strings.TrimSpace(strings.TrimLeft(trimmed, ";# "))
			if commentText != "" && len(info.Comments) < 20 {
				info.Comments = append(info.Comments, commentText)
			}
			continue
		}

		if match := sectionRegex.FindStringSubmatch(trimmed); len(match) > 1 {
			// Strip any trailing comments inside or after section bracket
			rawSec := match[1]
			if idx := strings.Index(rawSec, ";"); idx != -1 {
				rawSec = rawSec[:idx]
			}
			if idx := strings.Index(rawSec, "#"); idx != -1 {
				rawSec = rawSec[:idx]
			}
			currentSection = strings.ToLower(strings.TrimSpace(rawSec))

			if strings.HasPrefix(currentSection, "stageinfo") || strings.HasPrefix(currentSection, "bgdef") {
				hasStageInfo = true
			}
			if strings.HasPrefix(currentSection, "title info") || strings.HasPrefix(currentSection, "select info") || strings.HasPrefix(currentSection, "option info") {
				hasTitleInfo = true
			}
			if strings.HasPrefix(currentSection, "lifebar") || strings.HasPrefix(currentSection, "simul lifebar") ||
				strings.HasPrefix(currentSection, "tag lifebar") || strings.HasPrefix(currentSection, "round") ||
				strings.HasPrefix(currentSection, "winicon") || strings.HasPrefix(currentSection, "combo") {
				hasLifebarInfo = true
			}
			if strings.HasPrefix(currentSection, "scenedef") || strings.HasPrefix(currentSection, "scene ") || currentSection == "scene" {
				hasSceneDef = true
			}
			if currentSection == "files" {
				hasFiles = true
			}
			if currentSection == "info" {
				hasCharInfo = true
			}
			if strings.HasPrefix(currentSection, "font") || strings.HasPrefix(currentSection, "def") {
				hasFontInfo = true
			}
			if currentSection == "command" || currentSection == "statedef" {
				hasCmdInfo = true
			}
			continue
		}

		if match := keyValRegex.FindStringSubmatch(line); len(match) > 2 {
			rawKey := strings.ToLower(strings.TrimSpace(match[1]))
			val := cleanDefValue(match[2])

			switch currentSection {
			case "info":
				switch rawKey {
				case "name":
					info.Name = val
				case "displayname":
					info.DisplayName = val
				case "author":
					info.Author = val
				case "versiondate":
					info.VersionDate = val
				case "mugenversion":
					info.MugenVersion = val
				case "ikemenversion":
					info.IkemenVersion = val
				}
			case "files":
				hasFiles = true
				switch rawKey {
				case "sprite", "sff":
					info.SpriteFile = normalizeDefPath(val)
				case "sound", "snd":
					info.SoundFile = normalizeDefPath(val)
				case "anim", "air":
					info.AnimFile = normalizeDefPath(val)
				case "cmd":
					info.CmdFile = normalizeDefPath(val)
				case "cns":
					info.CnsFile = normalizeDefPath(val)
				}
			case "stageinfo", "bgdef":
				if rawKey == "name" && info.Name == "" {
					info.Name = val
				}
				if rawKey == "displayname" && info.DisplayName == "" {
					info.DisplayName = val
				}
				if rawKey == "author" && info.Author == "" {
					info.Author = val
				}
				if (rawKey == "spr" || rawKey == "sprite" || rawKey == "sff") && info.SpriteFile == "" {
					info.SpriteFile = normalizeDefPath(val)
				}
				if (rawKey == "bgmusic" || rawKey == "music") && info.SoundFile == "" {
					info.SoundFile = normalizeDefPath(val)
				}
			}
		}
	}

	info.IsStoryboard = hasSceneDef && !hasFiles && !hasStageInfo
	info.IsLifebar = hasLifebarInfo
	info.IsFont = hasFontInfo && !hasFiles && !hasStageInfo
	info.IsCommand = hasCmdInfo && !hasFiles && !hasCharInfo && !hasStageInfo
	info.HasFilesBlock = hasFiles

	// Categorize asset
	if hasStageInfo {
		info.Category = CategoryStage
		info.IsValidAsset = true
	} else if hasTitleInfo || hasLifebarInfo {
		info.Category = CategoryMotif
		info.IsValidAsset = true
	} else if info.IsStoryboard || info.IsFont || info.IsCommand {
		info.IsValidAsset = false
	} else if hasFiles || hasCharInfo || info.SpriteFile != "" || info.CmdFile != "" || info.CnsFile != "" {
		info.Category = CategoryFighter
		info.IsValidAsset = true
	} else {
		// Auxiliary / unknown .def
		info.Category = CategoryFighter
		info.IsValidAsset = (baseNoExt != "intro" && baseNoExt != "ending" && baseNoExt != "credits" && baseNoExt != "command")
	}

	if info.DisplayName == "" {
		info.DisplayName = info.Name
	}
	if info.DisplayName == "" {
		info.DisplayName = baseNoExt
	}
	if info.Author == "" {
		info.Author = "Unknown"
	}

	return info, scanner.Err()
}

// ScanFolderReadmes searches a folder for readme/credit files and scrapes additional URLs and notes.
func ScanFolderReadmes(dirPath string) (urls []string, notes string) {
	readmeNames := []string{
		"readme.txt", "read me.txt", "credits.txt", "about.txt",
		"instructions.txt", "readme.md", "credits.md",
	}

	urls = make([]string, 0)
	var noteLines []string

	for _, name := range readmeNames {
		target := filepath.Join(dirPath, name)
		data, err := os.ReadFile(target)
		if err == nil {
			content := string(data)
			matches := urlRegex.FindAllString(content, -1)
			for _, u := range matches {
				u = strings.TrimRight(u, ".,;)\"'>")
				if !containsString(urls, u) {
					urls = append(urls, u)
				}
			}

			// Capture first 500 characters of readme for preview/notes
			if len(noteLines) == 0 {
				lines := strings.Split(content, "\n")
				for i, l := range lines {
					if i > 15 {
						break
					}
					t := strings.TrimSpace(l)
					if t != "" {
						noteLines = append(noteLines, t)
					}
				}
			}
		}
	}

	if len(noteLines) > 0 {
		notes = strings.Join(noteLines, "\n")
	}

	return urls, notes
}

func cleanDefValue(val string) string {
	// Strip trailing inline comments
	if idx := strings.Index(val, ";"); idx != -1 {
		val = val[:idx]
	}
	if idx := strings.Index(val, "#"); idx != -1 {
		val = val[:idx]
	}
	val = strings.TrimSpace(val)
	// Strip surrounding double or single quotes
	val = strings.Trim(val, "\"'")
	return strings.TrimSpace(val)
}

func normalizeDefPath(val string) string {
	cleaned := cleanDefValue(val)
	// Normalize Windows backslashes to forward slashes
	cleaned = strings.ReplaceAll(cleaned, "\\", "/")
	return filepath.Clean(cleaned)
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

