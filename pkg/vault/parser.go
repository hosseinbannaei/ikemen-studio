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
	SpriteFile    string // Relative to .def file
	SoundFile     string
	FoundURLs     []string
	Comments      []string
	IsStoryboard  bool // e.g. ending.def, intro.def (contains [SceneDef])
	IsLifebar     bool // e.g. fight.def, lifebar.def (contains [Lifebar] or [Round])
	HasFilesBlock bool // true if [Files] section with cns/cmd/sprite exists
}

// ParseDefFile reads and extracts metadata from a character, stage, lifebar, or motif .def file.
func ParseDefFile(defPath string) (*DefInfo, error) {
	file, err := os.Open(defPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info := &DefInfo{
		Category:  CategoryFighter,
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

	baseName := strings.ToLower(filepath.Base(defPath))
	if baseName == "fight.def" || baseName == "lifebar.def" {
		hasLifebarInfo = true
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
			currentSection = strings.ToLower(strings.TrimSpace(match[1]))
			if strings.HasPrefix(currentSection, "stageinfo") || strings.HasPrefix(currentSection, "bgdef") {
				hasStageInfo = true
			}
			if strings.HasPrefix(currentSection, "title info") || strings.HasPrefix(currentSection, "select info") {
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
					info.SpriteFile = val
				case "sound", "snd":
					info.SoundFile = val
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
					info.SpriteFile = val
				}
				if (rawKey == "bgmusic" || rawKey == "music") && info.SoundFile == "" {
					info.SoundFile = val
				}
			}
		}
	}

	info.IsStoryboard = hasSceneDef && !hasFiles && !hasStageInfo
	info.IsLifebar = hasLifebarInfo
	info.HasFilesBlock = hasFiles

	if hasStageInfo {
		info.Category = CategoryStage
	} else if hasTitleInfo || hasLifebarInfo {
		info.Category = CategoryMotif
	}

	if info.DisplayName == "" {
		info.DisplayName = info.Name
	}
	if info.DisplayName == "" {
		// Fallback to base name of the .def file
		base := filepath.Base(defPath)
		info.DisplayName = strings.TrimSuffix(base, filepath.Ext(base))
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
	val = strings.TrimSpace(val)
	// Strip surrounding double quotes
	val = strings.Trim(val, "\"")
	return strings.TrimSpace(val)
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
