package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// EngineBackupInfo represents a saved snapshot of a project's engine runtime files
type EngineBackupInfo struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
}

// SwitchEngine migrates a project to a new engine version while creating a safe backup of previous runtime files
func SwitchEngine(projectDir, newEngineDir, newVersion string) error {
	if projectDir == "" {
		return fmt.Errorf("project directory cannot be empty")
	}
	if newEngineDir == "" {
		return fmt.Errorf("new engine directory cannot be empty")
	}
	if newVersion == "" {
		return fmt.Errorf("new engine version cannot be empty")
	}

	manifest, err := LoadManifest(projectDir)
	if err != nil {
		return fmt.Errorf("failed to load project manifest: %w", err)
	}

	oldVersion := manifest.Engine.Version

	// 1. Create a backup of current runtime files (external, lib)
	backupsDir := filepath.Join(projectDir, "save", "backups")
	_ = os.MkdirAll(backupsDir, 0755)

	backupID := fmt.Sprintf("engine_%s_%d", strings.ReplaceAll(oldVersion, "/", "_"), time.Now().Unix())
	backupFolder := filepath.Join(backupsDir, backupID)
	_ = os.MkdirAll(backupFolder, 0755)

	for _, folder := range []string{"external", "lib"} {
		src := filepath.Join(projectDir, folder)
		if fi, err := os.Stat(src); err == nil && fi.IsDir() {
			_ = copyDirRecursive(src, filepath.Join(backupFolder, folder))
		}
	}

	// Save backup metadata
	meta := EngineBackupInfo{
		ID:        backupID,
		Version:   oldVersion,
		Timestamp: time.Now().UTC(),
		Path:      backupFolder,
	}
	metaBytes, _ := json.MarshalIndent(meta, "", "  ")
	_ = os.WriteFile(filepath.Join(backupFolder, "backup_info.json"), metaBytes, 0644)

	// 2. Remove old runtime folders in project (external, lib)
	_ = os.RemoveAll(filepath.Join(projectDir, "external"))
	_ = os.RemoveAll(filepath.Join(projectDir, "lib"))

	// 3. Copy new runtime files from newEngineDir
	for _, folder := range []string{"external", "lib"} {
		src := filepath.Join(newEngineDir, folder)
		if fi, err := os.Stat(src); err == nil && fi.IsDir() {
			_ = copyDirRecursive(src, filepath.Join(projectDir, folder))
		}
	}

	// 4. Update manifest
	channel := "stable"
	if strings.Contains(strings.ToLower(newVersion), "nightly") || strings.Contains(strings.ToLower(newVersion), "pre") || strings.Contains(strings.ToLower(newVersion), "rc") {
		channel = "nightly"
	}

	manifest.Engine.Version = newVersion
	manifest.Engine.Channel = channel
	manifest.UpdatedAt = time.Now().UTC()

	return SaveManifest(projectDir, manifest)
}

// GetEngineBackups returns the list of available backups for rollback
func GetEngineBackups(projectDir string) ([]EngineBackupInfo, error) {
	backupsDir := filepath.Join(projectDir, "save", "backups")
	if _, err := os.Stat(backupsDir); os.IsNotExist(err) {
		return []EngineBackupInfo{}, nil
	}

	entries, err := os.ReadDir(backupsDir)
	if err != nil {
		return nil, err
	}

	var backups []EngineBackupInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		infoPath := filepath.Join(backupsDir, e.Name(), "backup_info.json")
		data, err := os.ReadFile(infoPath)
		if err != nil {
			continue
		}

		var info EngineBackupInfo
		if err := json.Unmarshal(data, &info); err == nil {
			info.Path = filepath.Join(backupsDir, e.Name())
			backups = append(backups, info)
		}
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// RollbackEngine restores a previous runtime backup and updates the manifest
func RollbackEngine(projectDir, backupID string) error {
	backupsDir := filepath.Join(projectDir, "save", "backups")
	backupFolder := filepath.Join(backupsDir, backupID)

	infoPath := filepath.Join(backupFolder, "backup_info.json")
	data, err := os.ReadFile(infoPath)
	if err != nil {
		return fmt.Errorf("backup info not found: %w", err)
	}

	var info EngineBackupInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return fmt.Errorf("invalid backup info: %w", err)
	}

	// 1. Remove current runtime folders in project
	_ = os.RemoveAll(filepath.Join(projectDir, "external"))
	_ = os.RemoveAll(filepath.Join(projectDir, "lib"))

	// 2. Restore folders from backup
	for _, folder := range []string{"external", "lib"} {
		src := filepath.Join(backupFolder, folder)
		if fi, err := os.Stat(src); err == nil && fi.IsDir() {
			_ = copyDirRecursive(src, filepath.Join(projectDir, folder))
		}
	}

	// 3. Update manifest
	manifest, err := LoadManifest(projectDir)
	if err != nil {
		return err
	}

	channel := "stable"
	if strings.Contains(strings.ToLower(info.Version), "nightly") || strings.Contains(strings.ToLower(info.Version), "pre") || strings.Contains(strings.ToLower(info.Version), "rc") {
		channel = "nightly"
	}

	manifest.Engine.Version = info.Version
	manifest.Engine.Channel = channel
	manifest.UpdatedAt = time.Now().UTC()

	return SaveManifest(projectDir, manifest)
}
