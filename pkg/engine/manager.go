package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ListInstalledEngines scans the engines directory and returns all valid installed engines
func ListInstalledEngines(enginesDir string) ([]InstalledEngine, error) {
	if _, err := os.Stat(enginesDir); os.IsNotExist(err) {
		return []InstalledEngine{}, nil
	}

	entries, err := os.ReadDir(enginesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read engines directory: %w", err)
	}

	var engines []InstalledEngine
	for _, entry := range entries {
		if !entry.IsDir() || strings.Contains(entry.Name(), ".tmp") {
			continue
		}

		version := entry.Name()
		enginePath := filepath.Join(enginesDir, version)
		execPath := FindEngineExecutable(enginePath)
		if execPath == "" {
			// Not a fully installed engine or missing binary
			continue
		}

		info, err := entry.Info()
		var installedAt time.Time
		if err == nil {
			installedAt = info.ModTime()
		}

		channel := "stable"
		if strings.Contains(strings.ToLower(version), "nightly") || strings.Contains(strings.ToLower(version), "pre") || strings.Contains(strings.ToLower(version), "rc") {
			channel = "nightly"
		}

		size := calculateDirSize(enginePath)

		engines = append(engines, InstalledEngine{
			Version:        version,
			Path:           enginePath,
			ExecutablePath: execPath,
			InstalledAt:    installedAt,
			Channel:        channel,
			Size:           size,
		})
	}

	return engines, nil
}

// FindEngineExecutable finds the binary inside an engine directory
func FindEngineExecutable(enginePath string) string {
	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			filepath.Join(enginePath, "Ikemen_GO.exe"),
			filepath.Join(enginePath, "ikemen_go.exe"),
			filepath.Join(enginePath, "Ikemen.exe"),
			filepath.Join(enginePath, "ikemen.exe"),
		}
	} else if runtime.GOOS == "darwin" {
		candidates = []string{
			filepath.Join(enginePath, "Ikemen_GO.app", "Contents", "MacOS", "Ikemen_GO"),
			filepath.Join(enginePath, "Ikemen_GO"),
			filepath.Join(enginePath, "ikemen_go"),
			filepath.Join(enginePath, "Ikemen"),
		}
	} else {
		candidates = []string{
			filepath.Join(enginePath, "Ikemen_GO"),
			filepath.Join(enginePath, "ikemen_go"),
			filepath.Join(enginePath, "Ikemen_GO.x86_64"),
			filepath.Join(enginePath, "Ikemen"),
			filepath.Join(enginePath, "ikemen"),
		}
	}

	for _, c := range candidates {
		if fi, err := os.Stat(c); err == nil && !fi.IsDir() {
			return c
		}
	}

	// Fallback recursive search if not at root
	var found string
	_ = filepath.Walk(enginePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		name := strings.ToLower(info.Name())
		if (name == "ikemen_go" || name == "ikemen_go.exe") && found == "" {
			found = path
		}
		return nil
	})

	return found
}

// DeleteEngine removes an installed engine directory
func DeleteEngine(enginesDir, version string) error {
	if version == "" || version == "." || version == ".." {
		return fmt.Errorf("invalid engine version specified: %s", version)
	}

	targetPath := filepath.Join(enginesDir, version)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("engine version %s does not exist", version)
	}

	return os.RemoveAll(targetPath)
}

func calculateDirSize(path string) int64 {
	var size int64
	_ = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
