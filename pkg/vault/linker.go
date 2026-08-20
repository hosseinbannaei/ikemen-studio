package vault

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LinkAssetToProject connects a vault asset to a target project and updates select.def.
func LinkAssetToProject(projectPath, vaultPath, assetKey string, strategy LinkStrategy) error {
	if strategy == "" {
		strategy = LinkStrategySymlink
	}

	srcAssetPath := filepath.Join(vaultPath, assetKey)
	destAssetPath := filepath.Join(projectPath, assetKey)

	if _, err := os.Stat(srcAssetPath); os.IsNotExist(err) {
		return fmt.Errorf("source asset %s does not exist in vault", assetKey)
	}

	// Ensure destination parent directory exists (e.g. projectPath/chars or stages)
	if err := os.MkdirAll(filepath.Dir(destAssetPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// If destination already exists, remove it if it's a symlink or return error
	if fi, err := os.Lstat(destAssetPath); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			_ = os.Remove(destAssetPath)
		} else {
			// Already exists as regular directory or file
			return fmt.Errorf("asset %s already exists in project", assetKey)
		}
	}

	var linkErr error
	switch strategy {
	case LinkStrategySymlink:
		// Attempt symlink
		linkErr = os.Symlink(srcAssetPath, destAssetPath)
		if linkErr != nil {
			// Graceful fallback to copy
			linkErr = copyDir(srcAssetPath, destAssetPath)
		}
	case LinkStrategyHardlink:
		linkErr = hardlinkDir(srcAssetPath, destAssetPath)
		if linkErr != nil {
			linkErr = copyDir(srcAssetPath, destAssetPath)
		}
	case LinkStrategyCopy:
		linkErr = copyDir(srcAssetPath, destAssetPath)
	default:
		linkErr = copyDir(srcAssetPath, destAssetPath)
	}

	if linkErr != nil {
		return fmt.Errorf("failed to link asset %s to project: %w", assetKey, linkErr)
	}

	// Update project's data/select.def
	_ = registerInSelectDef(projectPath, assetKey)

	return nil
}

// RemoveAssetFromProject removes the linked directory from the project and updates select.def.
func RemoveAssetFromProject(projectPath, assetKey string) error {
	destAssetPath := filepath.Join(projectPath, assetKey)

	// Remove symlink or directory
	if fi, err := os.Lstat(destAssetPath); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(destAssetPath); err != nil {
				return fmt.Errorf("failed to remove symlink %s: %w", destAssetPath, err)
			}
		} else if fi.IsDir() {
			if err := os.RemoveAll(destAssetPath); err != nil {
				return fmt.Errorf("failed to remove directory %s: %w", destAssetPath, err)
			}
		} else {
			_ = os.Remove(destAssetPath)
		}
	}

	// Comment out from select.def
	_ = unregisterFromSelectDef(projectPath, assetKey)

	return nil
}

func hardlinkDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dest, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return os.Link(path, target)
	})
}

func registerInSelectDef(projectPath, assetKey string) error {
	selectDefPath := filepath.Join(projectPath, "data", "select.def")
	if _, err := os.Stat(selectDefPath); os.IsNotExist(err) {
		return nil // No select.def present
	}

	parts := strings.Split(filepath.ToSlash(assetKey), "/")
	if len(parts) < 2 {
		return nil
	}

	category := parts[0]
	assetName := parts[1]

	data, err := os.ReadFile(selectDefPath)
	if err != nil {
		return err
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	var newLines []string

	if category == "chars" {
		// Check if already in lines
		for _, l := range lines {
			trimmed := strings.TrimSpace(l)
			if strings.EqualFold(trimmed, assetName) || strings.HasPrefix(strings.ToLower(trimmed), strings.ToLower(assetName)+",") {
				return nil // Already present
			}
		}

		inserted := false
		for _, l := range lines {
			newLines = append(newLines, l)
			trimmed := strings.ToLower(strings.TrimSpace(l))
			if strings.HasPrefix(trimmed, "[characters]") {
				newLines = append(newLines, assetName)
				inserted = true
			}
		}
		if !inserted {
			newLines = append(newLines, assetName)
		}
	} else if category == "stages" {
		stageEntry := fmt.Sprintf("stages/%s.def", assetName)
		// Check if already present
		for _, l := range lines {
			if strings.Contains(strings.ToLower(l), strings.ToLower(assetName)) {
				return nil
			}
		}

		inserted := false
		for _, l := range lines {
			newLines = append(newLines, l)
			trimmed := strings.ToLower(strings.TrimSpace(l))
			if strings.HasPrefix(trimmed, "[extrastages]") {
				newLines = append(newLines, stageEntry)
				inserted = true
			}
		}
		if !inserted {
			newLines = append(newLines, "[ExtraStages]", stageEntry)
		}
	} else {
		return nil
	}

	return os.WriteFile(selectDefPath, []byte(strings.Join(newLines, "\n")), 0644)
}

func unregisterFromSelectDef(projectPath, assetKey string) error {
	selectDefPath := filepath.Join(projectPath, "data", "select.def")
	if _, err := os.Stat(selectDefPath); os.IsNotExist(err) {
		return nil
	}

	parts := strings.Split(filepath.ToSlash(assetKey), "/")
	if len(parts) < 2 {
		return nil
	}
	assetName := strings.ToLower(parts[1])

	file, err := os.Open(selectDefPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var newLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "[") {
			newLines = append(newLines, line)
			continue
		}

		// If line matches assetName, comment it out
		cleanLine := strings.ToLower(trimmed)
		if strings.HasPrefix(cleanLine, assetName) {
			newLines = append(newLines, "; "+line+" (Unlinked from Vault)")
		} else {
			newLines = append(newLines, line)
		}
	}

	return os.WriteFile(selectDefPath, []byte(strings.Join(newLines, "\n")), 0644)
}
