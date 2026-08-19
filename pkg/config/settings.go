package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type Settings struct {
	EnginesDir     string   `json:"enginesDir"`
	Theme          string   `json:"theme"`
	RecentProjects []string `json:"recentProjects"`
	DefaultChannel string   `json:"defaultChannel"`
}

var (
	mu             sync.Mutex
	cachedSettings *Settings
)

// GetDefaultEnginesDir returns the platform-specific default directory for engine downloads
func GetDefaultEnginesDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	switch runtime.GOOS {
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			return filepath.Join(localAppData, "ikemen-studio", "engines")
		}
		return filepath.Join(homeDir, "AppData", "Local", "ikemen-studio", "engines")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "ikemen-studio", "engines")
	default: // linux, bsd, etc.
		xdgData := os.Getenv("XDG_DATA_HOME")
		if xdgData != "" {
			return filepath.Join(xdgData, "ikemen-studio", "engines")
		}
		return filepath.Join(homeDir, ".local", "share", "ikemen-studio", "engines")
	}
}

// GetConfigPath returns the platform-specific configuration file path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, "ikemen-studio", "settings.json")
		}
		return filepath.Join(homeDir, "AppData", "Roaming", "ikemen-studio", "settings.json")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "ikemen-studio", "settings.json")
	default: // linux, etc.
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			return filepath.Join(xdgConfig, "ikemen-studio", "settings.json")
		}
		return filepath.Join(homeDir, ".config", "ikemen-studio", "settings.json")
	}
}

// DefaultSettings returns fresh default configuration
func DefaultSettings() Settings {
	return Settings{
		EnginesDir:     GetDefaultEnginesDir(),
		Theme:          "dark",
		RecentProjects: []string{},
		DefaultChannel: "stable",
	}
}

// LoadSettings reads configuration from disk or creates default
func LoadSettings() (*Settings, error) {
	mu.Lock()
	defer mu.Unlock()

	configPath := GetConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaults := DefaultSettings()
		cachedSettings = &defaults
		_ = saveSettingsLocked(cachedSettings)
		return cachedSettings, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		defaults := DefaultSettings()
		cachedSettings = &defaults
		return cachedSettings, err
	}

	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		defaults := DefaultSettings()
		cachedSettings = &defaults
		return cachedSettings, err
	}

	if s.EnginesDir == "" {
		s.EnginesDir = GetDefaultEnginesDir()
	}
	if s.Theme == "" {
		s.Theme = "dark"
	}
	if s.DefaultChannel == "" {
		s.DefaultChannel = "stable"
	}
	if s.RecentProjects == nil {
		s.RecentProjects = []string{}
	}

	cachedSettings = &s
	return cachedSettings, nil
}

// SaveSettings writes configuration to disk
func SaveSettings(s *Settings) error {
	mu.Lock()
	defer mu.Unlock()
	return saveSettingsLocked(s)
}

func saveSettingsLocked(s *Settings) error {
	configPath := GetConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return err
	}

	cachedSettings = s
	return nil
}
