package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadGameConfig parses save/config.ini from project directory into key-value map
func LoadGameConfig(projectDir string) (map[string]string, error) {
	if projectDir == "" {
		return nil, fmt.Errorf("project directory cannot be empty")
	}

	iniPath := filepath.Join(projectDir, "save", "config.ini")
	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		// Return standard defaults if config.ini doesn't exist yet
		return map[string]string{
			"Width":        "1280",
			"Height":       "720",
			"Fullscreen":   "0",
			"VolumeMaster": "80",
			"VolumeBgm":    "80",
			"VolumeSfx":    "80",
			"Vsync":        "1",
			"RenderMode":   "OpenGL 3.3",
		}, nil
	}

	file, err := os.Open(iniPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	configMap := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// Strip inline comments if any
			if idx := strings.Index(val, ";"); idx != -1 {
				val = strings.TrimSpace(val[:idx])
			}
			configMap[key] = val
		}
	}

	return configMap, scanner.Err()
}

// SaveGameConfig updates save/config.ini while preserving comments and file structure
func SaveGameConfig(projectDir string, updates map[string]string) error {
	if projectDir == "" {
		return fmt.Errorf("project directory cannot be empty")
	}

	saveDir := filepath.Join(projectDir, "save")
	_ = os.MkdirAll(saveDir, 0755)
	iniPath := filepath.Join(saveDir, "config.ini")

	var lines []string
	updatedKeys := make(map[string]bool)

	if _, err := os.Stat(iniPath); err == nil {
		file, err := os.Open(iniPath)
		if err == nil {
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				rawLine := scanner.Text()
				trimmed := strings.TrimSpace(rawLine)

				if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "[") {
					lines = append(lines, rawLine)
					continue
				}

				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					matched := false
					for uk, uv := range updates {
						if strings.EqualFold(uk, key) {
							lines = append(lines, fmt.Sprintf("%s = %s", key, uv))
							updatedKeys[strings.ToLower(uk)] = true
							matched = true
							break
						}
					}
					if !matched {
						lines = append(lines, rawLine)
					}
				} else {
					lines = append(lines, rawLine)
				}
			}
			file.Close()
		}
	}

	// Append any new keys that weren't in the original file
	for uk, uv := range updates {
		if !updatedKeys[strings.ToLower(uk)] {
			lines = append(lines, fmt.Sprintf("%s = %s", uk, uv))
		}
	}

	content := strings.Join(lines, "\r\n") + "\r\n"
	return os.WriteFile(iniPath, []byte(content), 0644)
}
