package project

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const ManifestFileName = "ikemen-project.json"

// EngineConfig holds version requirements for a project
type EngineConfig struct {
	Version string `json:"version"`
	Channel string `json:"channel"`
}

// ProjectManifest represents the root manifest of an Ikemen GO Studio project
type ProjectManifest struct {
	Name      string       `json:"name"`
	Version   string       `json:"version"`
	Engine    EngineConfig `json:"engine"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Author    string       `json:"author"`
	Path      string       `json:"path"` // Absolute path to project root
}

// Validate checks the project manifest for required fields
func (m *ProjectManifest) Validate() error {
	if strings.TrimSpace(m.Name) == "" {
		return errors.New("project name cannot be empty")
	}
	if strings.TrimSpace(m.Version) == "" {
		return errors.New("project version cannot be empty")
	}
	if strings.TrimSpace(m.Engine.Version) == "" {
		return errors.New("engine version cannot be empty")
	}
	return nil
}

// LoadManifest reads and parses ikemen-project.json from the specified directory
func LoadManifest(projectDir string) (*ProjectManifest, error) {
	manifestPath := filepath.Join(projectDir, ManifestFileName)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no %s found in directory %s", ManifestFileName, projectDir)
		}
		return nil, fmt.Errorf("failed to read %s: %w", ManifestFileName, err)
	}

	var manifest ProjectManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("invalid %s format: %w", ManifestFileName, err)
	}

	manifest.Path = projectDir

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project manifest: %w", err)
	}

	return &manifest, nil
}

// SaveManifest writes the project manifest to ikemen-project.json in the specified directory
func SaveManifest(projectDir string, manifest *ProjectManifest) error {
	if err := manifest.Validate(); err != nil {
		return err
	}

	manifest.UpdatedAt = time.Now().UTC()
	manifest.Path = projectDir

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project manifest: %w", err)
	}

	manifestPath := filepath.Join(projectDir, ManifestFileName)
	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", ManifestFileName, err)
	}

	return nil
}
