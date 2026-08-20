package vault

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const ManifestFilename = "vault.json"
const CurrentManifestVersion = "1.0"

// GenerateVaultID produces a unique identifier for a new vault.
func GenerateVaultID() string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("vault-%d", time.Now().UnixNano())
	}
	return "vault-" + hex.EncodeToString(bytes)
}

// LoadManifest reads and parses vault.json from the given vault directory.
func LoadManifest(vaultPath string) (*VaultManifest, error) {
	manifestPath := filepath.Join(vaultPath, ManifestFilename)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vault manifest at %s: %w", manifestPath, err)
	}

	var manifest VaultManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse vault manifest at %s: %w", manifestPath, err)
	}

	if manifest.Assets == nil {
		manifest.Assets = make(map[string]VaultAsset)
	}

	return &manifest, nil
}

// SaveManifest writes the vault manifest atomically to vault.json in the vault directory.
func SaveManifest(vaultPath string, manifest *VaultManifest) error {
	manifest.UpdatedAt = time.Now().UTC()
	if manifest.Version == "" {
		manifest.Version = CurrentManifestVersion
	}
	if manifest.Assets == nil {
		manifest.Assets = make(map[string]VaultAsset)
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal vault manifest: %w", err)
	}

	manifestPath := filepath.Join(vaultPath, ManifestFilename)
	tempPath := filepath.Join(vaultPath, fmt.Sprintf(".vault_%d.tmp", time.Now().UnixNano()))

	// Ensure vault directory exists
	if err := os.MkdirAll(vaultPath, 0755); err != nil {
		return fmt.Errorf("failed to create vault directory %s: %w", vaultPath, err)
	}

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary vault manifest: %w", err)
	}

	if err := os.Rename(tempPath, manifestPath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to commit vault manifest to %s: %w", manifestPath, err)
	}

	return nil
}
