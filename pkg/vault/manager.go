package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ikemen-studio/pkg/config"
)

// VaultManager handles multi-vault registries, assets, and operations.
type VaultManager struct {
	mu sync.RWMutex
}

// NewVaultManager creates a new VaultManager and ensures the default vault exists.
func NewVaultManager() (*VaultManager, error) {
	vm := &VaultManager{}
	_, err := vm.EnsureDefaultVault()
	if err != nil {
		// Log but don't hard fail if initial disk setup has permissions issue
		fmt.Printf("Warning: failed to initialize default vault: %v\n", err)
	}
	return vm, nil
}

// EnsureDefaultVault makes sure the standard default vault exists and is registered.
func (vm *VaultManager) EnsureDefaultVault() (*VaultInfo, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	defaultBase := config.GetDefaultVaultsDir()
	defaultVaultPath := filepath.Join(defaultBase, "default")

	settings, err := config.LoadSettings()
	if err != nil {
		return nil, err
	}

	// Check if already registered
	isRegistered := false
	for _, p := range settings.RegisteredVaults {
		if strings.EqualFold(filepath.Clean(p), filepath.Clean(defaultVaultPath)) {
			isRegistered = true
			break
		}
	}

	// Check if manifest exists at default path
	manifestPath := filepath.Join(defaultVaultPath, ManifestFilename)
	var manifest *VaultManifest
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Create default vault
		_ = os.MkdirAll(filepath.Join(defaultVaultPath, "chars"), 0755)
		_ = os.MkdirAll(filepath.Join(defaultVaultPath, "stages"), 0755)
		_ = os.MkdirAll(filepath.Join(defaultVaultPath, "motifs"), 0755)
		_ = os.MkdirAll(filepath.Join(defaultVaultPath, "sound"), 0755)

		manifest = &VaultManifest{
			Version:     CurrentManifestVersion,
			ID:          "vault-default",
			Name:        "Default Vault",
			Description: "Standard global asset library for Ikemen Studio",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Assets:      make(map[string]VaultAsset),
		}
		if err := SaveManifest(defaultVaultPath, manifest); err != nil {
			return nil, err
		}
	} else {
		manifest, _ = LoadManifest(defaultVaultPath)
	}

	if !isRegistered {
		settings.RegisteredVaults = append([]string{defaultVaultPath}, settings.RegisteredVaults...)
		_ = config.SaveSettings(settings)
	}

	return &VaultInfo{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Description: manifest.Description,
		Path:        defaultVaultPath,
		AssetCount:  len(manifest.Assets),
		SizeBytes:   calculateDirSize(defaultVaultPath),
		IsDefault:   true,
		CreatedAt:   manifest.CreatedAt,
	}, nil
}

// GetVaults returns all registered vaults with summary statistics.
func (vm *VaultManager) GetVaults() ([]VaultInfo, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	settings, err := config.LoadSettings()
	if err != nil {
		return nil, err
	}

	defaultBase := config.GetDefaultVaultsDir()
	defaultVaultPath := filepath.Clean(filepath.Join(defaultBase, "default"))

	var vaults []VaultInfo
	for _, p := range settings.RegisteredVaults {
		cleanP := filepath.Clean(p)
		manifest, err := LoadManifest(cleanP)
		if err != nil {
			continue // Skip invalid/deleted paths
		}

		manifestChanged := false
		for key, asset := range manifest.Assets {
			assetDiskPath := filepath.Join(cleanP, asset.Key)
			if _, err := os.Stat(assetDiskPath); os.IsNotExist(err) {
				delete(manifest.Assets, key)
				manifestChanged = true
			}
		}
		if manifestChanged {
			_ = SaveManifest(cleanP, manifest)
		}

		isDef := strings.EqualFold(cleanP, defaultVaultPath) || manifest.ID == "vault-default"

		vaults = append(vaults, VaultInfo{
			ID:          manifest.ID,
			Name:        manifest.Name,
			Description: manifest.Description,
			Path:        cleanP,
			AssetCount:  len(manifest.Assets),
			SizeBytes:   calculateDirSize(cleanP),
			IsDefault:   isDef,
			CreatedAt:   manifest.CreatedAt,
		})
	}

	return vaults, nil
}

// GetVault finds a vault by ID and returns its metadata and path.
func (vm *VaultManager) GetVault(vaultID string) (*VaultInfo, string, error) {
	vaults, err := vm.GetVaults()
	if err != nil {
		return nil, "", err
	}

	if vaultID == "" || vaultID == "all" || vaultID == "vault-default" {
		for _, v := range vaults {
			if v.IsDefault || v.ID == "vault-default" {
				return &v, v.Path, nil
			}
		}
		if len(vaults) > 0 {
			return &vaults[0], vaults[0].Path, nil
		}
	}

	for _, v := range vaults {
		if v.ID == vaultID {
			return &v, v.Path, nil
		}
	}

	return nil, "", fmt.Errorf("vault with ID %s not found", vaultID)
}

// CreateVault initializes a new vault at the specified directory.
func (vm *VaultManager) CreateVault(name, description, targetPath string) (*VaultInfo, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if targetPath == "" {
		defaultBase := config.GetDefaultVaultsDir()
		safeName := sanitizeFolderName(name)
		targetPath = filepath.Join(defaultBase, safeName)
	}

	targetPath = filepath.Clean(targetPath)
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create vault folder: %w", err)
	}

	_ = os.MkdirAll(filepath.Join(targetPath, "chars"), 0755)
	_ = os.MkdirAll(filepath.Join(targetPath, "stages"), 0755)
	_ = os.MkdirAll(filepath.Join(targetPath, "motifs"), 0755)
	_ = os.MkdirAll(filepath.Join(targetPath, "sound"), 0755)

	manifest := &VaultManifest{
		Version:     CurrentManifestVersion,
		ID:          GenerateVaultID(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		Assets:      make(map[string]VaultAsset),
	}

	if err := SaveManifest(targetPath, manifest); err != nil {
		return nil, err
	}

	// Register in settings
	settings, err := config.LoadSettings()
	if err == nil {
		settings.RegisteredVaults = append(settings.RegisteredVaults, targetPath)
		_ = config.SaveSettings(settings)
	}

	return &VaultInfo{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Description: manifest.Description,
		Path:        targetPath,
		AssetCount:  0,
		SizeBytes:   calculateDirSize(targetPath),
		IsDefault:   false,
		CreatedAt:   manifest.CreatedAt,
	}, nil
}

// RegisterVault connects an existing vault folder on disk to the app settings.
func (vm *VaultManager) RegisterVault(vaultPath string) (*VaultInfo, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	vaultPath = filepath.Clean(vaultPath)
	manifest, err := LoadManifest(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("folder is not a valid vault (missing vault.json): %w", err)
	}

	settings, err := config.LoadSettings()
	if err != nil {
		return nil, err
	}

	for _, p := range settings.RegisteredVaults {
		if strings.EqualFold(filepath.Clean(p), vaultPath) {
			// Already registered
			return &VaultInfo{
				ID:          manifest.ID,
				Name:        manifest.Name,
				Description: manifest.Description,
				Path:        vaultPath,
				AssetCount:  len(manifest.Assets),
				SizeBytes:   calculateDirSize(vaultPath),
				CreatedAt:   manifest.CreatedAt,
			}, nil
		}
	}

	settings.RegisteredVaults = append(settings.RegisteredVaults, vaultPath)
	_ = config.SaveSettings(settings)

	return &VaultInfo{
		ID:          manifest.ID,
		Name:        manifest.Name,
		Description: manifest.Description,
		Path:        vaultPath,
		AssetCount:  len(manifest.Assets),
		SizeBytes:   calculateDirSize(vaultPath),
		CreatedAt:   manifest.CreatedAt,
	}, nil
}

// UnregisterVault removes a vault from the app without deleting its files on disk.
func (vm *VaultManager) UnregisterVault(vaultID string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	settings, err := config.LoadSettings()
	if err != nil {
		return err
	}

	var updated []string
	for _, p := range settings.RegisteredVaults {
		manifest, err := LoadManifest(p)
		if err == nil && manifest.ID == vaultID {
			if manifest.ID == "vault-default" {
				return fmt.Errorf("cannot unregister the default system vault")
			}
			continue
		}
		updated = append(updated, p)
	}

	settings.RegisteredVaults = updated
	return config.SaveSettings(settings)
}

// GetVaultAssets returns all assets for a given vault or all vaults if vaultID is "all" or empty.
func (vm *VaultManager) GetVaultAssets(vaultID string) ([]VaultAsset, error) {
	vm.mu.RLock()
	defer vm.mu.RUnlock()

	var allAssets []VaultAsset

	vaults, err := vm.GetVaults()
	if err != nil {
		return nil, err
	}

	for _, v := range vaults {
		if vaultID != "" && vaultID != "all" && v.ID != vaultID {
			continue
		}

		manifest, err := LoadManifest(v.Path)
		if err != nil {
			continue
		}

		manifestChanged := false
		for key, asset := range manifest.Assets {
			// Verify that the asset folder still exists on disk
			assetDiskPath := filepath.Join(v.Path, asset.Key)
			if _, err := os.Stat(assetDiskPath); os.IsNotExist(err) {
				delete(manifest.Assets, key)
				manifestChanged = true
				continue
			}

			// Populate preview base64 if not already cached
			if asset.PreviewImage != "" && asset.PreviewBase64 == "" {
				fullPreviewPath := filepath.Join(v.Path, asset.PreviewImage)
				if _, err := os.Stat(fullPreviewPath); err == nil {
					_, b64, _ := ExtractAndCachePortrait(v.Path, asset.Key, "")
					asset.PreviewBase64 = b64
				}
			}
			allAssets = append(allAssets, asset)
		}

		if manifestChanged {
			_ = SaveManifest(v.Path, manifest)
		}
	}

	return allAssets, nil
}

// UpdateVaultAsset modifies metadata, tags, source URL, or notes of an existing asset.
func (vm *VaultManager) UpdateVaultAsset(vaultID, assetKey string, update AssetMetadataUpdate) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	var vaultPath string
	if vaultID != "" && vaultID != "all" {
		if _, vPath, err := vm.GetVault(vaultID); err == nil {
			vaultPath = vPath
		}
	}

	if vaultPath == "" {
		// Auto-locate owning vault
		vaults, _ := vm.GetVaults()
		for _, v := range vaults {
			if manifest, err := LoadManifest(v.Path); err == nil {
				if _, exists := manifest.Assets[assetKey]; exists {
					vaultPath = v.Path
					break
				}
			}
		}
	}

	if vaultPath == "" {
		return fmt.Errorf("asset %s not found in any registered vault", assetKey)
	}

	manifest, err := LoadManifest(vaultPath)
	if err != nil {
		return err
	}

	asset, exists := manifest.Assets[assetKey]
	if !exists {
		return fmt.Errorf("asset %s not found in vault %s", assetKey, vaultPath)
	}

	if update.DisplayName != "" {
		asset.DisplayName = update.DisplayName
	}
	if update.Author != "" {
		asset.Author = update.Author
	}
	if update.SourceURL != "" {
		asset.SourceURL = update.SourceURL
	}
	if update.License != "" {
		asset.License = update.License
	}
	if update.Tags != nil {
		asset.Tags = update.Tags
	}
	if update.Notes != "" {
		asset.Notes = update.Notes
	}

	manifest.Assets[assetKey] = asset
	return SaveManifest(vaultPath, manifest)
}

// DeleteVaultAsset removes an asset from the vault manifest and disk.
func (vm *VaultManager) DeleteVaultAsset(vaultID, assetKey string) error {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	var vaultPath string
	if vaultID != "" && vaultID != "all" {
		if _, vPath, err := vm.GetVault(vaultID); err == nil {
			vaultPath = vPath
		}
	}

	if vaultPath == "" {
		// Auto-locate owning vault
		vaults, _ := vm.GetVaults()
		for _, v := range vaults {
			if manifest, err := LoadManifest(v.Path); err == nil {
				if _, exists := manifest.Assets[assetKey]; exists {
					vaultPath = v.Path
					break
				}
			}
		}
	}

	if vaultPath == "" {
		return fmt.Errorf("asset %s not found in any registered vault", assetKey)
	}

	manifest, err := LoadManifest(vaultPath)
	if err != nil {
		return err
	}

	asset, exists := manifest.Assets[assetKey]
	if !exists {
		return fmt.Errorf("asset %s not found in vault %s", assetKey, vaultPath)
	}

	// Remove physical folder
	assetDiskPath := filepath.Join(vaultPath, asset.Key)
	_ = os.RemoveAll(assetDiskPath)

	// Remove preview thumbnail
	if asset.PreviewImage != "" {
		_ = os.Remove(filepath.Join(vaultPath, asset.PreviewImage))
	}

	delete(manifest.Assets, assetKey)
	return SaveManifest(vaultPath, manifest)
}

// IngestAsset handles ingestion, supporting targeting current vault or creating a dedicated new vault.
func (vm *VaultManager) IngestAsset(vaultID, srcPath, targetMode string) (*IngestResult, error) {
	return vm.IngestMultiple(vaultID, []string{srcPath}, targetMode)
}

// IngestMultiple handles batch ingestion of multiple archives/folders.
func (vm *VaultManager) IngestMultiple(vaultID string, srcPaths []string, targetMode string) (*IngestResult, error) {
	var targetVaultPath string

	if targetMode == "new_vault" && len(srcPaths) == 1 {
		baseName := strings.TrimSuffix(filepath.Base(srcPaths[0]), filepath.Ext(srcPaths[0]))
		vaultName := strings.ReplaceAll(baseName, "_", " ")
		info, err := vm.CreateVault(vaultName, "Created from "+filepath.Base(srcPaths[0]), "")
		if err != nil {
			return nil, err
		}
		targetVaultPath = info.Path
	} else {
		if vaultID == "" || vaultID == "all" {
			vaultID = "vault-default"
		}
		_, vPath, err := vm.GetVault(vaultID)
		if err != nil {
			return nil, err
		}
		targetVaultPath = vPath
	}

	return IngestMultiple(targetVaultPath, srcPaths, "")
}
