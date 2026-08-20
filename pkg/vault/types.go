package vault

import "time"

// Category defines the structural taxonomy of an asset.
type Category string

const (
	CategoryFighter Category = "fighters"
	CategoryStage   Category = "stages"
	CategoryMotif   Category = "motifs"
	CategorySound   Category = "sounds"
)

// LinkStrategy defines how an asset is connected into a project.
type LinkStrategy string

const (
	LinkStrategySymlink  LinkStrategy = "symlink"
	LinkStrategyHardlink LinkStrategy = "hardlink"
	LinkStrategyCopy     LinkStrategy = "copy"
)

// VaultAsset represents a single indexed asset inside a Vault.
type VaultAsset struct {
	Key           string    `json:"key"`            // e.g. "chars/kfm" or "stages/suzaku"
	Category      Category  `json:"category"`       // fighters, stages, motifs, sounds
	DisplayName   string    `json:"display_name"`   // Visual display name
	Author        string    `json:"author"`         // Author name
	VersionDate   string    `json:"version_date"`   // Version / release date
	MugenVersion  string    `json:"mugen_version"`  // e.g. "1.0", "1.1", "04,14,2002"
	IkemenVersion string    `json:"ikemen_version"` // e.g. "0.99"
	SourceURL     string    `json:"source_url"`     // URL where downloaded/hosted
	SourcePackage string    `json:"source_package"` // Original archive filename
	License       string    `json:"license"`        // License / commercial rights
	Tags          []string  `json:"tags"`           // User & auto tags
	PreviewImage  string    `json:"preview_image"`  // Relative path to cached PNG (e.g. ".previews/chars_kfm.png")
	PreviewBase64 string    `json:"preview_base64"` // Base64 data URL for frontend rendering
	Notes         string    `json:"notes"`          // Custom notes or readme excerpt
	SizeBytes     int64     `json:"size_bytes"`     // Total folder size in bytes
	AddedAt       time.Time `json:"added_at"`       // Timestamp when added to vault
}

// VaultManifest represents the vault.json file stored in each vault root.
type VaultManifest struct {
	Version     string                `json:"version"`
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Assets      map[string]VaultAsset `json:"assets"`
}

// VaultInfo is a lightweight descriptor for UI display.
type VaultInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	AssetCount  int       `json:"asset_count"`
	SizeBytes   int64     `json:"size_bytes"`
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
}

// AssetMetadataUpdate holds fields that the user can modify on an existing asset.
type AssetMetadataUpdate struct {
	DisplayName string   `json:"display_name"`
	Author      string   `json:"author"`
	SourceURL   string   `json:"source_url"`
	License     string   `json:"license"`
	Tags        []string `json:"tags"`
	Notes       string   `json:"notes"`
}

// IngestResult details the outcome of an archive or folder ingestion.
type IngestResult struct {
	VaultID        string       `json:"vault_id"`
	DetectedAssets []VaultAsset `json:"detected_assets"`
	IsMultiAsset   bool         `json:"is_multi_asset"`
	SourcePackage  string       `json:"source_package"`
	ImportedCount  int          `json:"imported_count"`
	Warnings       []string     `json:"warnings"`
}

// VaultCleanReport details the outcome of a vault diagnostic and cleanup operation.
type VaultCleanReport struct {
	VaultID                string   `json:"vault_id"`
	RemovedDuplicates      int      `json:"removed_duplicates"`
	CleanedContaminations  int      `json:"cleaned_contaminations"`
	RegeneratedPreviews    int      `json:"regenerated_previews"`
	PrunedMissing          int      `json:"pruned_missing"`
	TotalAssetsNow         int      `json:"total_assets_now"`
	Details                []string `json:"details"`
}

