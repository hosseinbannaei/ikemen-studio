package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ListInstalledEngines scans the engines directory (and any standard fallback paths) and returns all valid installed engines
func ListInstalledEngines(enginesDir string) ([]InstalledEngine, error) {
	seenVersions := make(map[string]bool)
	var engines []InstalledEngine

	// Directories to scan in priority order
	candidateDirs := []string{enginesDir}

	// Add standard fallback paths only when not given an explicit custom directory
	if enginesDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			stdDir := filepath.Join(homeDir, ".local", "share", "ikemen-studio", "engines")
			candidateDirs = append(candidateDirs, stdDir)

			// Also check VS Code / Snap sandbox directory if present
			snapGlob := filepath.Join(homeDir, "snap", "code", "*", ".local", "share", "ikemen-studio", "engines")
			if matches, err := filepath.Glob(snapGlob); err == nil {
				for _, m := range matches {
					candidateDirs = append(candidateDirs, m)
				}
			}
		}
	}


	for _, dir := range candidateDirs {
		if dir == "" {
			continue
		}
		if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || strings.Contains(entry.Name(), ".tmp") {
				continue
			}

			version := entry.Name()
			if seenVersions[strings.ToLower(version)] {
				continue
			}

			enginePath := filepath.Join(dir, version)
			execPath := FindEngineExecutable(enginePath)
			if execPath == "" {
				continue
			}

			// Ensure binary is executable on Unix
			if runtime.GOOS != "windows" {
				_ = os.Chmod(execPath, 0755)
			}

			info, err := entry.Info()
			var installedAt time.Time
			if err == nil {
				installedAt = info.ModTime()
			}

			channel := "stable"
			if strings.Contains(strings.ToLower(version), "nightly") || strings.Contains(strings.ToLower(version), "pre") || strings.Contains(strings.ToLower(version), "rc") || strings.Contains(strings.ToLower(version), "dev") {
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
			seenVersions[strings.ToLower(version)] = true
		}
	}

	return engines, nil
}


// FindEngineExecutable finds the binary inside an engine directory
func FindEngineExecutable(enginePath string) string {
	var candidates []string
	if runtime.GOOS == "windows" {
		candidates = []string{
			filepath.Join(enginePath, "Ikemen_GO_Windows.exe"),
			filepath.Join(enginePath, "Ikemen_GO.exe"),
			filepath.Join(enginePath, "ikemen_go.exe"),
			filepath.Join(enginePath, "Ikemen.exe"),
			filepath.Join(enginePath, "ikemen.exe"),
		}
	} else if runtime.GOOS == "darwin" {
		candidates = []string{
			filepath.Join(enginePath, "Ikemen_GO_MacOS"),
			filepath.Join(enginePath, "Ikemen_GO.app", "Contents", "MacOS", "Ikemen_GO"),
			filepath.Join(enginePath, "Ikemen_GO"),
			filepath.Join(enginePath, "ikemen_go"),
			filepath.Join(enginePath, "Ikemen"),
		}
	} else {
		candidates = []string{
			filepath.Join(enginePath, "Ikemen_GO_Linux"),
			filepath.Join(enginePath, "ikemen_go_linux"),
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
		if (strings.HasPrefix(name, "ikemen_go") || strings.HasPrefix(name, "ikemen")) &&
			!strings.HasSuffix(name, ".png") &&
			!strings.HasSuffix(name, ".desktop") &&
			!strings.HasSuffix(name, ".txt") &&
			!strings.HasSuffix(name, ".md") &&
			!strings.HasSuffix(name, ".command") &&
			found == "" {
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
