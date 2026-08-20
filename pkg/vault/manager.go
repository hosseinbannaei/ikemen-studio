package vault

import (
	"fmt"
	"io"
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

// CleanAndRepairVault scans, diagnoses, deduplicates, and repairs an existing vault.
func (vm *VaultManager) CleanAndRepairVault(vaultID string) (*VaultCleanReport, error) {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vaultID == "" || vaultID == "all" {
		vaultID = "vault-default"
	}

	_, vaultPath, err := vm.GetVault(vaultID)
	if err != nil {
		return nil, err
	}

	manifest, err := LoadManifest(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load vault manifest: %w", err)
	}

	report := &VaultCleanReport{
		VaultID: vaultID,
		Details: make([]string, 0),
	}

	// 1. Scan disk folders and identify duplicates
	charsDir := filepath.Join(vaultPath, "chars")
	if entries, err := os.ReadDir(charsDir); err == nil {
		// Map canonical name -> list of folder names on disk
		folderMap := make(map[string][]string)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			folderName := e.Name()
			// Check if ends with _<number>
			baseName := folderName
			lastUnderscore := strings.LastIndex(folderName, "_")
			if lastUnderscore != -1 && lastUnderscore < len(folderName)-1 {
				numPart := folderName[lastUnderscore+1:]
				isNum := true
				for _, ch := range numPart {
					if ch < '0' || ch > '9' {
						isNum = false
						break
					}
				}
				if isNum {
					baseName = folderName[:lastUnderscore]
				}
			}
			folderMap[baseName] = append(folderMap[baseName], folderName)
		}

		// Process duplicate timestamp folders
		for baseName, folders := range folderMap {
			if len(folders) > 1 {
				// Keep the cleanest folder name (prefer baseName without suffix)
				keepFolder := ""
				for _, f := range folders {
					if strings.EqualFold(f, baseName) {
						keepFolder = f
						break
					}
				}
				if keepFolder == "" {
					keepFolder = folders[0]
				}

				for _, f := range folders {
					if f != keepFolder {
						dupPath := filepath.Join(charsDir, f)
						_ = os.RemoveAll(dupPath)
						dupKey := fmt.Sprintf("chars/%s", f)
						delete(manifest.Assets, dupKey)
						// Remove preview
						_ = os.Remove(filepath.Join(vaultPath, ".previews", fmt.Sprintf("chars_%s.png", f)))
						report.RemovedDuplicates++
						report.Details = append(report.Details, fmt.Sprintf("Removed duplicate character: %s (kept %s)", f, keepFolder))
					}
				}
			}
		}

		// Also check by display name + author
		nameAuthorMap := make(map[string][]string)
		for key, asset := range manifest.Assets {
			if asset.Category == CategoryFighter {
				combo := fmt.Sprintf("%s|%s", strings.ToLower(strings.TrimSpace(asset.DisplayName)), strings.ToLower(strings.TrimSpace(asset.Author)))
				if asset.DisplayName != "" && asset.DisplayName != "Unknown" {
					nameAuthorMap[combo] = append(nameAuthorMap[combo], key)
				}
			}
		}
		for combo, keys := range nameAuthorMap {
			if len(keys) > 1 {
				keepKey := keys[0]
				for _, k := range keys[1:] {
					diskP := filepath.Join(vaultPath, k)
					_ = os.RemoveAll(diskP)
					delete(manifest.Assets, k)
					safePrev := strings.ReplaceAll(k, "/", "_") + ".png"
					_ = os.Remove(filepath.Join(vaultPath, ".previews", safePrev))
					report.RemovedDuplicates++
					report.Details = append(report.Details, fmt.Sprintf("Removed duplicate character by metadata (%s): %s (kept %s)", combo, k, keepKey))
				}
			}
		}
	}

	// 2. Clean cross-contaminations & fix nesting
	if entries, err := os.ReadDir(charsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			charDir := filepath.Join(charsDir, e.Name())

			// Check for nested subdirectories that contain separate characters
			subEntries, subErr := os.ReadDir(charDir)
			if subErr != nil {
				continue
			}

			for _, sub := range subEntries {
				if !sub.IsDir() {
					continue
				}
				subName := strings.ToLower(sub.Name())

				// Ignore standard valid subdirectories
				if subName == "palettes" || subName == "pal" || subName == "states" ||
					subName == "storyboard" || subName == "sound" || subName == "intro" ||
					subName == "ending" || subName == "cutscene" || subName == "txt" ||
					subName == "id" || subName == "media" || subName == "code" ||
					subName == "common" || subName == "add" || subName == "backup" ||
					subName == "ai" || subName == "introending" || subName == "portraits" {
					continue
				}

				subDirPath := filepath.Join(charDir, sub.Name())

				// Check if subDirPath contains a character .def
				hasSubDef := false
				var subDefs []string
				_ = filepath.Walk(subDirPath, func(p string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".def") {
						if defInfo, parseErr := ParseDefFile(p); parseErr == nil && defInfo != nil && defInfo.IsValidAsset {
							hasSubDef = true
							subDefs = append(subDefs, p)
						}
					}
					return nil
				})

				if hasSubDef {
					// Check if this is an identical nested folder wrapper (e.g. venom_01/Venom 01/)
					cleanSub := sanitizeFolderName(sub.Name())
					cleanParent := sanitizeFolderName(e.Name())
					if cleanSub == cleanParent {
						// Flatten contents up to parent
						innerFiles, _ := os.ReadDir(subDirPath)
						for _, inF := range innerFiles {
							oldPath := filepath.Join(subDirPath, inF.Name())
							newPath := filepath.Join(charDir, inF.Name())
							if _, exists := os.Stat(newPath); os.IsNotExist(exists) {
								_ = os.Rename(oldPath, newPath)
							}
						}
						_ = os.RemoveAll(subDirPath)
						report.CleanedContaminations++
						report.Details = append(report.Details, fmt.Sprintf("Flattened redundant nested wrapper folder %s in %s", sub.Name(), e.Name()))
					} else {
						// Foreign cross-contaminated directory (e.g. Black Cat1 inside ingridyamazaki)
						_ = os.RemoveAll(subDirPath)
						report.CleanedContaminations++
						report.Details = append(report.Details, fmt.Sprintf("Removed foreign cross-contaminated folder %s from %s", sub.Name(), e.Name()))
					}
				}
			}
		}
	}

	// 3. Prune missing assets and re-index disk assets
	for key := range manifest.Assets {
		assetDiskPath := filepath.Join(vaultPath, key)
		if _, err := os.Stat(assetDiskPath); os.IsNotExist(err) {
			delete(manifest.Assets, key)
			report.PrunedMissing++
			report.Details = append(report.Details, fmt.Sprintf("Pruned missing asset record: %s", key))
		}
	}

	// Re-scan categories to pick up any valid unindexed disk assets
	for _, catDir := range []string{"chars", "stages", "motifs", "sound"} {
		catPath := filepath.Join(vaultPath, catDir)
		catEntries, err := os.ReadDir(catPath)
		if err != nil {
			continue
		}

		for _, e := range catEntries {
			if !e.IsDir() && catDir == "chars" {
				continue
			}
			folderName := e.Name()
			assetKey := fmt.Sprintf("%s/%s", catDir, folderName)

			assetDiskPath := filepath.Join(catPath, folderName)

			// If not in manifest or metadata missing, index it
			existing, inManifest := manifest.Assets[assetKey]
			if !inManifest || existing.DisplayName == "" || existing.DisplayName == "Unknown" {
				// Find primary .def
				var primaryDef string
				_ = filepath.Walk(assetDiskPath, func(p string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(p), ".def") {
						if defInfo, parseErr := ParseDefFile(p); parseErr == nil && defInfo != nil && defInfo.IsValidAsset {
							primaryDef = p
							return io.EOF
						}
					}
					return nil
				})

				if primaryDef != "" {
					defInfo, _ := ParseDefFile(primaryDef)
					if defInfo != nil {
						cat := CategoryFighter
						if catDir == "stages" {
							cat = CategoryStage
						} else if catDir == "motifs" {
							cat = CategoryMotif
						} else if catDir == "sound" {
							cat = CategorySound
						}

						sffPath := ""
						if defInfo.SpriteFile != "" {
							sffClean := strings.ReplaceAll(defInfo.SpriteFile, "\\", "/")
							sffPath = filepath.Join(assetDiskPath, sffClean)
							if _, err := os.Stat(sffPath); err != nil {
								sffPath = ResolvePathCaseInsensitive(assetDiskPath, sffClean)
							}
						}
						if sffPath == "" {
							sffPath = findFirstExt(assetDiskPath, ".sff")
						}

						relPrev, b64, _ := ExtractAndCachePortrait(vaultPath, assetKey, sffPath)
						urls, notes := ScanFolderReadmes(assetDiskPath)

						manifest.Assets[assetKey] = VaultAsset{
							Key:           assetKey,
							Category:      cat,
							DisplayName:   defInfo.DisplayName,
							Author:        defInfo.Author,
							VersionDate:   defInfo.VersionDate,
							MugenVersion:  defInfo.MugenVersion,
							IkemenVersion: defInfo.IkemenVersion,
							SourceURL:     strings.Join(urls, ", "),
							SourcePackage: "vault_recovery",
							License:       "Unknown / Fan-made",
							PreviewImage:  relPrev,
							PreviewBase64: b64,
							Notes:         notes,
							SizeBytes:     calculateDirSize(assetDiskPath),
							AddedAt:       time.Now().UTC(),
						}
						report.Details = append(report.Details, fmt.Sprintf("Re-indexed unindexed asset: %s (%s)", assetKey, defInfo.DisplayName))
					}
				}
			}
		}
	}

	// 4. Regenerate missing preview images & base64
	for key, asset := range manifest.Assets {
		fullPreviewPath := filepath.Join(vaultPath, asset.PreviewImage)
		needRegen := false
		if asset.PreviewImage == "" {
			needRegen = true
		} else if _, statErr := os.Stat(fullPreviewPath); statErr != nil {
			needRegen = true
		}

		if needRegen {
			assetDiskPath := filepath.Join(vaultPath, key)
			sffPath := findFirstExt(assetDiskPath, ".sff")
			relPrev, b64, err := ExtractAndCachePortrait(vaultPath, key, sffPath)
			if err == nil {
				asset.PreviewImage = relPrev
				asset.PreviewBase64 = b64
				manifest.Assets[key] = asset
				report.RegeneratedPreviews++
			}
		}
	}

	report.TotalAssetsNow = len(manifest.Assets)
	_ = SaveManifest(vaultPath, manifest)

	return report, nil
}

