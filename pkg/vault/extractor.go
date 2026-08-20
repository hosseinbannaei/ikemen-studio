package vault

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DiscoveredAsset is a scanned asset found within an unpacked archive or folder.
type DiscoveredAsset struct {
	Category       Category
	FolderRoot     string // Root folder containing the asset files
	DefPath        string // Absolute path to the main .def file
	DefInfo        *DefInfo
	FolderName     string // Clean folder name (e.g. "ryu")
	ScrapedURLs    []string
	ScrapedNotes   string
	TotalSizeBytes int64
}

// IngestPath processes a zip, rar, 7z archive, or folder, and installs detected assets into the vault.
func IngestPath(vaultPath, srcPath, sourcePackage string) (*IngestResult, error) {
	return IngestMultiple(vaultPath, []string{srcPath}, sourcePackage)
}

// IngestMultiple processes multiple archive files or folders and installs all detected assets into the vault.
func IngestMultiple(vaultPath string, srcPaths []string, sourcePackage string) (*IngestResult, error) {
	manifest, err := LoadManifest(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load target vault manifest: %w", err)
	}

	if sourcePackage == "" && len(srcPaths) == 1 {
		sourcePackage = filepath.Base(srcPaths[0])
	} else if sourcePackage == "" {
		sourcePackage = fmt.Sprintf("batch_import_%d", time.Now().Unix())
	}

	// Create temporary staging directory
	stagingDir, err := os.MkdirTemp("", "ikemen_ingest_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	var extractErrors []string
	var discovered []DiscoveredAsset

	// Process each source path in its OWN isolated sub-staging directory
	for i, src := range srcPaths {
		srcBase := filepath.Base(src)
		srcClean := sanitizeFolderName(strings.TrimSuffix(srcBase, filepath.Ext(srcBase)))
		if srcClean == "" {
			srcClean = fmt.Sprintf("pkg_%03d", i)
		}
		subStaging := filepath.Join(stagingDir, fmt.Sprintf("src_%03d_%s", i, srcClean))
		if err := os.MkdirAll(subStaging, 0755); err != nil {
			extractErrors = append(extractErrors, fmt.Sprintf("%s: failed to create sub-staging: %v", srcBase, err))
			continue
		}

		if err := extractOrCopy(src, subStaging); err != nil {
			extractErrors = append(extractErrors, fmt.Sprintf("%s: %v", srcBase, err))
			continue
		}

		// Check if the sub-staging directory has loose character files at its root (e.g. def/cns directly at root)
		ensureLooseFilesWrapped(subStaging, srcClean)

		// Scan this isolated staging directory
		subDiscovered, scanErr := ScanStagingDir(subStaging)
		if scanErr == nil && len(subDiscovered) > 0 {
			discovered = append(discovered, subDiscovered...)
		}
	}

	if len(discovered) == 0 {
		if len(extractErrors) > 0 {
			return nil, fmt.Errorf("failed to extract assets:\n%s", strings.Join(extractErrors, "\n"))
		}
		return nil, fmt.Errorf("no valid MUGEN/Ikemen assets (.def files) found in provided files/folders")
	}

	result := &IngestResult{
		VaultID:        manifest.ID,
		DetectedAssets: make([]VaultAsset, 0, len(discovered)),
		IsMultiAsset:   len(discovered) > 1,
		SourcePackage:  sourcePackage,
		Warnings:       extractErrors,
	}

	// Track imported keys to prevent internal collisions within the same batch
	importedInBatch := make(map[string]bool)

	for _, disc := range discovered {
		categoryDir := "chars"
		switch disc.Category {
		case CategoryStage:
			categoryDir = "stages"
		case CategoryMotif:
			categoryDir = "motifs"
		case CategoryLifebar:
			categoryDir = "lifebars"
		case CategorySound:
			categoryDir = "sound"
		case CategoryFont:
			categoryDir = "fonts"
		case CategoryStoryboard:
			categoryDir = "storyboards"
		}

		targetCategoryPath := filepath.Join(vaultPath, categoryDir)
		if err := os.MkdirAll(targetCategoryPath, 0755); err != nil {
			return nil, err
		}

		// Determine target asset directory name
		cleanName := sanitizeFolderName(disc.FolderName)
		if cleanName == "" && disc.DefInfo != nil {
			cleanName = sanitizeFolderName(disc.DefInfo.Name)
		}
		if cleanName == "" {
			cleanName = "asset_" + fmt.Sprintf("%d", time.Now().UnixNano()%100000)
		}

		assetKey := fmt.Sprintf("%s/%s", categoryDir, cleanName)

		// If collision within the SAME batch from two distinct subfolders with the same name, disambiguate
		if importedInBatch[assetKey] {
			cleanName = fmt.Sprintf("%s_%d", cleanName, time.Now().UnixNano()%1000)
			assetKey = fmt.Sprintf("%s/%s", categoryDir, cleanName)
		}
		importedInBatch[assetKey] = true

		destAssetDir := filepath.Join(targetCategoryPath, cleanName)

		// In-place update / overwrite: If directory already exists on disk, remove old contents cleanly
		if _, err := os.Stat(destAssetDir); err == nil {
			_ = os.RemoveAll(destAssetDir)
		}

		// Copy asset directory from staging to vault
		if err := copyDir(disc.FolderRoot, destAssetDir); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Failed to copy %s: %v", cleanName, err))
			continue
		}

		// Locate SFF file for portrait extraction
		var sffPath string
		if disc.DefInfo.SpriteFile != "" {
			sffClean := strings.ReplaceAll(disc.DefInfo.SpriteFile, "\\", "/")
			sffPath = filepath.Join(destAssetDir, sffClean)
			if _, err := os.Stat(sffPath); err != nil {
				sffPath = ResolvePathCaseInsensitive(destAssetDir, sffClean)
			}
		}

		if sffPath == "" {
			sffPath = findFirstExt(destAssetDir, ".sff")
		}

		// Extract portrait and generate preview thumbnail
		relPreview, base64Preview, _ := ExtractAndCachePortrait(vaultPath, assetKey, sffPath)

		// Auto-tags
		tags := make([]string, 0)
		if disc.DefInfo.MugenVersion != "" {
			tags = append(tags, "MUGEN "+disc.DefInfo.MugenVersion)
		} else if disc.DefInfo.IkemenVersion != "" {
			tags = append(tags, "Ikemen "+disc.DefInfo.IkemenVersion)
		}

		sourceURL := ""
		if len(disc.ScrapedURLs) > 0 {
			sourceURL = disc.ScrapedURLs[0]
		}
		if sourceURL == "" && len(disc.DefInfo.FoundURLs) > 0 {
			sourceURL = disc.DefInfo.FoundURLs[0]
		}

		notes := disc.ScrapedNotes
		if notes == "" && len(disc.DefInfo.Comments) > 0 {
			notes = strings.Join(disc.DefInfo.Comments, "\n")
		}

		sizeBytes := calculateDirSize(destAssetDir)

		vaultAsset := VaultAsset{
			Key:           assetKey,
			Category:      disc.Category,
			DisplayName:   disc.DefInfo.DisplayName,
			Author:        disc.DefInfo.Author,
			VersionDate:   disc.DefInfo.VersionDate,
			MugenVersion:  disc.DefInfo.MugenVersion,
			IkemenVersion: disc.DefInfo.IkemenVersion,
			SourceURL:     sourceURL,
			SourcePackage: sourcePackage,
			License:       "Unknown / Fan-made",
			Tags:          tags,
			PreviewImage:  relPreview,
			PreviewBase64: base64Preview,
			Notes:         notes,
			SizeBytes:     sizeBytes,
			AddedAt:       time.Now().UTC(),
		}

		// Preserve existing user tags / notes if updating an existing asset
		if existing, exists := manifest.Assets[assetKey]; exists {
			if len(existing.Tags) > 0 && len(vaultAsset.Tags) == 0 {
				vaultAsset.Tags = existing.Tags
			}
			if existing.Notes != "" && vaultAsset.Notes == "" {
				vaultAsset.Notes = existing.Notes
			}
			if existing.SourceURL != "" && vaultAsset.SourceURL == "" {
				vaultAsset.SourceURL = existing.SourceURL
			}
			vaultAsset.AddedAt = existing.AddedAt
		}

		manifest.Assets[assetKey] = vaultAsset
		result.DetectedAssets = append(result.DetectedAssets, vaultAsset)
		result.ImportedCount++
	}

	if err := SaveManifest(vaultPath, manifest); err != nil {
		return nil, fmt.Errorf("failed to save updated vault manifest: %w", err)
	}

	return result, nil
}

// ensureLooseFilesWrapped checks if staging directory has loose character files at its root.
// If so, moves them into an enclosing folder so they don't pollute parent staging scans.
func ensureLooseFilesWrapped(stagingDir, archiveName string) {
	entries, err := os.ReadDir(stagingDir)
	if err != nil {
		return
	}

	hasLooseDef := false
	var looseItems []string
	subDirCount := 0

	for _, e := range entries {
		if !e.IsDir() {
			looseItems = append(looseItems, e.Name())
			if strings.HasSuffix(strings.ToLower(e.Name()), ".def") {
				hasLooseDef = true
			}
		} else {
			subDirCount++
		}
	}

	// If there are loose .def files directly at the root, wrap everything into archiveName subfolder
	if hasLooseDef {
		wrapFolder := filepath.Join(stagingDir, archiveName)
		if err := os.MkdirAll(wrapFolder, 0755); err != nil {
			return
		}

		for _, e := range entries {
			if e.Name() == archiveName {
				continue
			}
			oldP := filepath.Join(stagingDir, e.Name())
			newP := filepath.Join(wrapFolder, e.Name())
			_ = os.Rename(oldP, newP)
		}
	}
}

// ScanStagingDir recursively looks for .def files and isolates the enclosing asset roots.
func ScanStagingDir(stagingDir string) ([]DiscoveredAsset, error) {
	var discovered []DiscoveredAsset

	type candidateDef struct {
		path    string
		defInfo *DefInfo
	}
	folderDefs := make(map[string][]candidateDef)

	// Auxiliary folders to ignore as standalone character roots
	auxiliaryDirs := map[string]bool{
		"storyboard": true, "intro": true, "ending": true, "cutscene": true,
		"states": true, "pal": true, "palettes": true, "media": true, "id": true,
		"code": true, "common": true, "add": true, "backup": true, "ai": true,
		"introending": true, "txt": true, "sound": true, "portraits": true,
	}

	err := filepath.Walk(stagingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) == ".def" {
			folderRoot := filepath.Dir(path)
			folderName := strings.ToLower(filepath.Base(folderRoot))

			// Skip if inside an auxiliary folder
			if auxiliaryDirs[folderName] {
				return nil
			}

			defInfo, parseErr := ParseDefFile(path)
			if parseErr == nil && defInfo != nil {
				folderDefs[folderRoot] = append(folderDefs[folderRoot], candidateDef{
					path:    path,
					defInfo: defInfo,
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// For each directory root, select the best primary .def file
	for folderRoot, candidates := range folderDefs {
		folderName := filepath.Base(folderRoot)
		if folderRoot == stagingDir || strings.HasPrefix(folderName, "src_") {
			if len(candidates) > 0 {
				folderName = strings.TrimSuffix(filepath.Base(candidates[0].path), filepath.Ext(candidates[0].path))
			}
		}

		var bestDef *candidateDef
		bestScore := -999

		for _, c := range candidates {
			score := scoreDefCandidate(c.path, c.defInfo, folderName)
			if score > bestScore {
				bestScore = score
				candCopy := c
				bestDef = &candCopy
			}
		}

		// If best score is <= 0 or not a valid asset, skip
		if bestDef == nil || bestScore <= 0 || !bestDef.defInfo.IsValidAsset {
			continue
		}

		urls, notes := ScanFolderReadmes(folderRoot)
		sizeBytes := calculateDirSize(folderRoot)

		discovered = append(discovered, DiscoveredAsset{
			Category:       bestDef.defInfo.Category,
			FolderRoot:     folderRoot,
			DefPath:        bestDef.path,
			DefInfo:        bestDef.defInfo,
			FolderName:     folderName,
			ScrapedURLs:    urls,
			ScrapedNotes:   notes,
			TotalSizeBytes: sizeBytes,
		})
	}

	return discovered, nil
}

func scoreDefCandidate(defPath string, info *DefInfo, folderName string) int {
	base := strings.ToLower(strings.TrimSuffix(filepath.Base(defPath), filepath.Ext(defPath)))

	// Discard invalid / command defs completely
	if info.IsCommand || !info.IsValidAsset || base == "command" || base == "cmd" {
		return -1000
	}

	score := 10

	// Match folder name (e.g. kfm.def inside kfm/ or Black Cat1.def inside Black Cat1/)
	cleanBase := sanitizeFolderName(base)
	cleanFolder := sanitizeFolderName(folderName)
	if strings.EqualFold(cleanBase, cleanFolder) || strings.EqualFold(base, folderName) {
		score += 60
	}

	if info.HasFilesBlock {
		score += 40
	}

	if info.SpriteFile != "" {
		score += 20
	}

	if info.CmdFile != "" || info.CnsFile != "" || info.AnimFile != "" {
		score += 20
	}

	switch info.Category {
	case CategoryFighter:
		score += 30
	case CategoryStage:
		score += 40
	case CategoryMotif:
		score += 40
	case CategoryLifebar:
		score += 40
	case CategoryStoryboard:
		score += 35
	case CategoryFont:
		score += 25
	}

	if info.DisplayName != "" && !strings.EqualFold(info.DisplayName, "Unknown") {
		score += 10
	}

	return score
}

func isArchive(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".zip" || ext == ".rar" || ext == ".7z" || ext == ".tar" ||
		ext == ".gz" || ext == ".tgz" || ext == ".bz2" || ext == ".xz"
}

func extractOrCopy(src, dest string) error {
	fi, err := os.Stat(src)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		// Scan directory for nested archives first
		var nestedArchives []string
		_ = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && isArchive(path) {
				nestedArchives = append(nestedArchives, path)
			}
			return nil
		})

		if len(nestedArchives) > 0 {
			for _, arch := range nestedArchives {
				base := strings.TrimSuffix(filepath.Base(arch), filepath.Ext(arch))
				archDest := filepath.Join(dest, "archive_"+sanitizeFolderName(base))
				_ = os.MkdirAll(archDest, 0755)
				_ = extractSingleArchive(arch, archDest)
			}
		}

		// Also copy loose folders/files
		return copyDir(src, dest)
	}

	return extractSingleArchive(src, dest)
}

func extractSingleArchive(src, dest string) error {
	ext := strings.ToLower(filepath.Ext(src))

	if ext == ".zip" {
		if err := unzip(src, dest); err == nil {
			return nil
		}
		return extractCLI(src, dest)
	} else if ext == ".gz" || ext == ".tgz" || ext == ".tar" {
		if err := untar(src, dest); err == nil {
			return nil
		}
		return extractCLI(src, dest)
	} else if ext == ".rar" || ext == ".7z" || ext == ".bz2" || ext == ".xz" {
		return extractCLI(src, dest)
	}

	return fmt.Errorf("unsupported file format: %s", ext)
}

func extractCLI(src, dest string) error {
	_ = os.MkdirAll(dest, 0755)

	// Try 7z / 7za
	if p, err := exec.LookPath("7z"); err == nil {
		cmd := exec.Command(p, "x", "-y", fmt.Sprintf("-o%s", dest), src)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	if p, err := exec.LookPath("7za"); err == nil {
		cmd := exec.Command(p, "x", "-y", fmt.Sprintf("-o%s", dest), src)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// Try unar
	if p, err := exec.LookPath("unar"); err == nil {
		cmd := exec.Command(p, "-o", dest, "-f", src)
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// Try unrar
	if p, err := exec.LookPath("unrar"); err == nil {
		cmd := exec.Command(p, "x", "-y", "-o+", src, dest+string(os.PathSeparator))
		if out, err := cmd.CombinedOutput(); err == nil {
			return nil
		} else {
			_ = out
		}
	}

	// Try unzip (for zip fallback)
	if strings.HasSuffix(strings.ToLower(src), ".zip") {
		if p, err := exec.LookPath("unzip"); err == nil {
			cmd := exec.Command(p, "-o", src, "-d", dest)
			if out, err := cmd.CombinedOutput(); err == nil {
				return nil
			} else {
				_ = out
			}
		}
	}

	return fmt.Errorf("could not extract %s: no compatible extraction utility (7z, unar, unrar) found", filepath.Base(src))
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) && fpath != dest {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	var tr *tar.Reader
	if strings.HasSuffix(src, ".gz") || strings.HasSuffix(src, ".tgz") {
		gzr, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer gzr.Close()
		tr = tar.NewReader(gzr)
	} else {
		tr = tar.NewReader(file)
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) && target != dest {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Don't copy archive files directly into staging asset folders
		if !info.IsDir() && isArchive(path) {
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, relPath)
		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		return copyFile(path, targetPath)
	})
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func calculateDirSize(dirPath string) int64 {
	var size int64
	_ = filepath.Walk(dirPath, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

func findFirstExt(dir, ext string) string {
	var result string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			if strings.ToLower(filepath.Ext(path)) == strings.ToLower(ext) {
				result = path
				return io.EOF
			}
		}
		return nil
	})
	return result
}

func sanitizeFolderName(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	result := name
	for _, ch := range invalid {
		result = strings.ReplaceAll(result, ch, "_")
	}
	return strings.ToLower(strings.Trim(result, "._"))
}

