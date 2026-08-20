package project

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"ikemen-studio/pkg/config"
	"ikemen-studio/pkg/vault"
)

// ProjectMotifInfo represents detailed metadata about an installed screenpack/motif.
type ProjectMotifInfo struct {
	Key               string `json:"key"`
	DisplayName       string `json:"display_name"`
	Author            string `json:"author"`
	Version           string `json:"version"`
	Resolution        string `json:"resolution"`
	GridColumns       int    `json:"grid_columns"`
	GridRows          int    `json:"grid_rows"`
	TotalSlots        int    `json:"total_slots"`
	IsActive          bool   `json:"is_active"`
	SpriteFile        string `json:"sprite_file"`
	SoundFile         string `json:"sound_file"`
	PreviewBase64     string `json:"preview_base64"`
	IsLinkedFromVault bool   `json:"is_linked_from_vault"`
}

// GetProjectMotifs scans data/ and subdirectories for system.def files and detects the active motif.
func GetProjectMotifs(projectDir string) ([]ProjectMotifInfo, error) {
	dataDir := filepath.Join(projectDir, "data")
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return []ProjectMotifInfo{}, nil
	}

	// 1. Determine currently active motif from save/config.ini
	activeMotifPath := getActiveMotifFromConfig(projectDir)

	var motifs []ProjectMotifInfo

	// 2. Discover all system.def files
	err := filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		baseName := strings.ToLower(filepath.Base(path))
		if baseName == "system.def" {
			rel, _ := filepath.Rel(projectDir, path)
			rel = filepath.ToSlash(rel)

			motifMeta := parseMotifSystemDef(path)
			if motifMeta == nil {
				return nil
			}

			// Check if active
			isActive := normalizeMotifKey(rel) == normalizeMotifKey(activeMotifPath)
			if activeMotifPath == "" && rel == "data/system.def" {
				isActive = true
			}

			// Symlink detection
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

			// Portrait / Logo preview extraction
			previewBase64 := ""
			if motifMeta.SpriteFile != "" {
				fullSff := filepath.Join(filepath.Dir(path), motifMeta.SpriteFile)
				if _, sErr := os.Stat(fullSff); sErr == nil {
					_, b64, _ := vault.ExtractAndCachePortrait(filepath.Dir(path), filepath.Base(filepath.Dir(path)), fullSff)
					previewBase64 = b64
				}
			}

			motif := ProjectMotifInfo{
				Key:               rel,
				DisplayName:       motifMeta.DisplayName,
				Author:            motifMeta.Author,
				Version:           motifMeta.MugenVersion,
				Resolution:        motifMeta.Resolution,
				GridColumns:       motifMeta.GridColumns,
				GridRows:          motifMeta.GridRows,
				TotalSlots:        motifMeta.GridColumns * motifMeta.GridRows,
				IsActive:          isActive,
				SpriteFile:        motifMeta.SpriteFile,
				SoundFile:         motifMeta.SoundFile,
				PreviewBase64:     previewBase64,
				IsLinkedFromVault: isSymlink,
			}

			motifs = append(motifs, motif)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan motifs: %w", err)
	}

	return motifs, nil
}

// SetActiveMotif updates the Motif path in save/config.ini and ensures system configuration integrity.
func SetActiveMotif(projectDir, motifRelativePath string) error {
	cleanPath := filepath.ToSlash(strings.TrimSpace(motifRelativePath))
	fullPath := filepath.Join(projectDir, filepath.FromSlash(cleanPath))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("motif system.def not found at %s", cleanPath)
	}

	cfgPath := filepath.Join(projectDir, "save", "config.ini")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		// If config.ini does not exist, create it with motif setting
		_ = os.MkdirAll(filepath.Dir(cfgPath), 0755)
		_ = os.WriteFile(cfgPath, []byte(fmt.Sprintf("[Options]\nMotif = %s\n", cleanPath)), 0644)
		return nil
	}

	return config.SaveGameConfig(projectDir, map[string]string{
		"Motif": cleanPath,
	})
}


type motifDefMeta struct {
	DisplayName  string
	Author       string
	MugenVersion string
	Resolution   string
	GridColumns  int
	GridRows     int
	SpriteFile   string
	SoundFile    string
}

func parseMotifSystemDef(defPath string) *motifDefMeta {
	file, err := os.Open(defPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	parentFolder := filepath.Base(filepath.Dir(defPath))
	if parentFolder == "data" {
		parentFolder = "Default Screenpack"
	}

	meta := &motifDefMeta{
		DisplayName:  parentFolder,
		Author:       "Elecbyte / Ikemen",
		MugenVersion: "1.0",
		Resolution:   "1280x720",
		GridColumns:  8,
		GridRows:     10,
	}

	scanner := bufio.NewScanner(file)
	currentSec := ""
	sectionRe := regexp.MustCompile(`^\s*\[\s*([^\]]+)\s*\]`)
	keyValRe := regexp.MustCompile(`^\s*([^=;]+)\s*=\s*(.*)$`)

	resX, resY := 0, 0

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
					if meta.DisplayName == "Default Screenpack" || meta.DisplayName == parentFolder {
						meta.DisplayName = v
					}
				case "displayname":
					meta.DisplayName = v
				case "author":
					meta.Author = v
				case "mugenversion", "ikemenversion":
					meta.MugenVersion = v
				}
			case "select info":
				switch k {
				case "columns", "rows":
					if val, pErr := strconv.Atoi(v); pErr == nil && val > 0 {
						if k == "columns" {
							meta.GridColumns = val
						} else {
							meta.GridRows = val
						}
					}
				}
			case "files":
				switch k {
				case "spr", "sprite", "sff":
					meta.SpriteFile = filepath.ToSlash(v)
				case "snd", "sound":
					meta.SoundFile = filepath.ToSlash(v)
				}
			case "title info", "system info":
				switch k {
				case "localcoord":
					parts := strings.Split(v, ",")
					if len(parts) >= 2 {
						rx, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
						ry, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
						if rx > 0 && ry > 0 {
							resX, resY = rx, ry
						}
					}
				}
			}
		}
	}

	if resX > 0 && resY > 0 {
		meta.Resolution = fmt.Sprintf("%dx%d", resX, resY)
	}

	return meta
}

func getActiveMotifFromConfig(projectDir string) string {
	cfg, err := config.LoadGameConfig(projectDir)
	if err != nil {
		return "data/system.def"
	}
	if motif, ok := cfg["motif"]; ok && motif != "" {
		return filepath.ToSlash(motif)
	}
	if motif, ok := cfg["Motif"]; ok && motif != "" {
		return filepath.ToSlash(motif)
	}
	return "data/system.def"
}


func normalizeMotifKey(p string) string {
	cleaned := filepath.ToSlash(strings.TrimSpace(p))
	return strings.ToLower(cleaned)
}
