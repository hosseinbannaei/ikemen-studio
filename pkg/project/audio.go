package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectAudioInfo represents an audio file (.mp3, .ogg, .wav) with its current game event and stage assignments.
type ProjectAudioInfo struct {
	RelativePath      string   `json:"relative_path"`
	FileName          string   `json:"file_name"`
	Format            string   `json:"format"`
	SizeBytes         int64    `json:"size_bytes"`
	AssignedEvents    []string `json:"assigned_events"`
	AssignedStages    []string `json:"assigned_stages"`
	IsLinkedFromVault bool     `json:"is_linked_from_vault"`
}

// GetProjectAudio scans sound/ and subfolders to list all BGM tracks and inspects their system assignments.
func GetProjectAudio(projectDir string) ([]ProjectAudioInfo, error) {
	soundDir := filepath.Join(projectDir, "sound")
	if _, err := os.Stat(soundDir); os.IsNotExist(err) {
		return []ProjectAudioInfo{}, nil
	}

	// 1. Parse system.def [Music] assignments
	systemMusicMap := parseSystemDefMusic(projectDir)

	// 2. Parse stage music assignments
	stageMusicMap := parseAllStageMusic(projectDir)

	var audioList []ProjectAudioInfo
	audioExts := map[string]bool{
		".mp3": true, ".ogg": true, ".wav": true, ".flac": true, ".mid": true,
	}

	err := filepath.Walk(soundDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if audioExts[ext] {
			rel, _ := filepath.Rel(projectDir, path)
			rel = filepath.ToSlash(rel)

			isSymlink := false
			if fi, lErr := os.Lstat(path); lErr == nil && (fi.Mode()&os.ModeSymlink != 0) {
				isSymlink = true
			}

			normKey := normalizeAudioKey(rel)
			baseName := filepath.Base(path)

			events := systemMusicMap[normKey]
			if len(events) == 0 {
				events = systemMusicMap[normalizeAudioKey(baseName)]
			}

			stages := stageMusicMap[normKey]
			if len(stages) == 0 {
				stages = stageMusicMap[normalizeAudioKey(baseName)]
			}

			audio := ProjectAudioInfo{
				RelativePath:      rel,
				FileName:          baseName,
				Format:            strings.TrimPrefix(ext, "."),
				SizeBytes:         info.Size(),
				AssignedEvents:    events,
				AssignedStages:    stages,
				IsLinkedFromVault: isSymlink,
			}

			audioList = append(audioList, audio)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan sound directory: %w", err)
	}

	return audioList, nil
}

// SetSystemBGM sets background music for Title, Select, VS, or Victory in system.def.
// eventType can be "title", "select", "vs", "victory".
func SetSystemBGM(projectDir, eventType, audioRelativePath string) error {
	systemDefPath := findSystemDef(projectDir)
	if _, err := os.Stat(systemDefPath); os.IsNotExist(err) {
		return fmt.Errorf("system.def not found at %s", systemDefPath)
	}

	data, err := os.ReadFile(systemDefPath)
	if err != nil {
		return err
	}

	cleanKey := strings.ToLower(strings.TrimSpace(eventType))
	if !strings.HasSuffix(cleanKey, ".bgm") {
		cleanKey = cleanKey + ".bgm"
	}

	cleanAudioPath := filepath.ToSlash(strings.TrimSpace(audioRelativePath))

	lines := strings.Split(string(data), "\n")
	var newLines []string
	inMusic := false
	bgmUpdated := false
	hasMusicBlock := false

	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)

		if strings.HasPrefix(lower, "[") {
			if inMusic && !bgmUpdated {
				newLines = append(newLines, fmt.Sprintf("%s = %s", cleanKey, cleanAudioPath))
				bgmUpdated = true
			}
			inMusic = strings.HasPrefix(lower, "[music]")
			if inMusic {
				hasMusicBlock = true
			}
			newLines = append(newLines, line)
			continue
		}

		if inMusic && !bgmUpdated {
			if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
				k := strings.ToLower(strings.TrimSpace(m[1]))
				if k == cleanKey {
					newLines = append(newLines, fmt.Sprintf("%s = %s", cleanKey, cleanAudioPath))
					bgmUpdated = true
					continue
				}
			}
		}

		newLines = append(newLines, line)
	}

	if !bgmUpdated {
		if hasMusicBlock {
			newLines = append(newLines, fmt.Sprintf("%s = %s", cleanKey, cleanAudioPath))
		} else {
			newLines = append(newLines, "\n[Music]", fmt.Sprintf("%s = %s", cleanKey, cleanAudioPath))
		}
	}

	return os.WriteFile(systemDefPath, []byte(strings.Join(newLines, "\n")), 0644)
}

// DeleteProjectAudio removes an audio file from the project's sound/ directory.
func DeleteProjectAudio(projectDir, audioRelativePath string) error {
	fullPath := filepath.Join(projectDir, filepath.FromSlash(audioRelativePath))
	return os.Remove(fullPath)
}

func parseSystemDefMusic(projectDir string) map[string][]string {
	res := make(map[string][]string)
	systemDefPath := findSystemDef(projectDir)

	file, err := os.Open(systemDefPath)
	if err != nil {
		return res
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inMusic := false
	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}

		lower := strings.ToLower(line)
		if strings.HasPrefix(lower, "[") {
			inMusic = strings.HasPrefix(lower, "[music]")
			continue
		}

		if inMusic {
			if m := keyValRe.FindStringSubmatch(line); len(m) > 2 {
				k := strings.ToLower(strings.TrimSpace(m[1]))
				v := cleanVal(m[2])
				if v != "" {
					label := k
					switch k {
					case "title.bgm":
						label = "Title Screen"
					case "select.bgm":
						label = "Character Select"
					case "vs.bgm":
						label = "VS Screen"
					case "victory.bgm":
						label = "Victory Screen"
					}
					norm := normalizeAudioKey(v)
					res[norm] = append(res[norm], label)
				}
			}
		}
	}

	return res
}

func parseAllStageMusic(projectDir string) map[string][]string {
	res := make(map[string][]string)
	stagesDir := filepath.Join(projectDir, "stages")

	_ = filepath.Walk(stagesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".def" {
			meta := parseStageDef(path)
			if meta != nil && meta.SoundFile != "" {
				norm := normalizeAudioKey(meta.SoundFile)
				stageBase := filepath.Base(path)
				res[norm] = append(res[norm], stageBase)
			}
		}
		return nil
	})

	return res
}

func normalizeAudioKey(p string) string {
	cleaned := filepath.ToSlash(strings.TrimSpace(p))
	cleaned = strings.TrimPrefix(cleaned, "sound/")
	return strings.ToLower(cleaned)
}
