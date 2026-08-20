package vault

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestVaultManifestCRUD(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "vault_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	manifest := &VaultManifest{
		Version:     "1.0",
		ID:          "vault-test-1",
		Name:        "Test Vault",
		Description: "Unit testing vault",
		Assets:      make(map[string]VaultAsset),
	}

	manifest.Assets["chars/kfm"] = VaultAsset{
		Key:         "chars/kfm",
		Category:    CategoryFighter,
		DisplayName: "Kung Fu Man",
		Author:      "Elecbyte",
	}

	if err := SaveManifest(tempDir, manifest); err != nil {
		t.Fatalf("SaveManifest failed: %v", err)
	}

	loaded, err := LoadManifest(tempDir)
	if err != nil {
		t.Fatalf("LoadManifest failed: %v", err)
	}

	if loaded.Name != "Test Vault" {
		t.Errorf("Expected name 'Test Vault', got '%s'", loaded.Name)
	}

	if len(loaded.Assets) != 1 {
		t.Fatalf("Expected 1 asset, got %d", len(loaded.Assets))
	}

	asset := loaded.Assets["chars/kfm"]
	if asset.DisplayName != "Kung Fu Man" || asset.Author != "Elecbyte" {
		t.Errorf("Asset data mismatch: %+v", asset)
	}
}

func TestParseDefFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "def_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	defContent := `; Sample character def
; Check out http://mugen.example.com/kfm for updates
[Info]
name = "kfm"
displayname = "Kung Fu Man"
versiondate = "04,14,2002"
mugenversion = "1.0"
author = "Elecbyte"
pal.defaults = 1,2,3

[Files]
cmd     = kfm.cmd
cns     = kfm.cns
st      = kfm.cns
stcommon = common1.cns
sprite  = kfm.sff
anim    = kfm.air
sound   = kfm.snd
`
	defPath := filepath.Join(tempDir, "kfm.def")
	if err := os.WriteFile(defPath, []byte(defContent), 0644); err != nil {
		t.Fatal(err)
	}

	info, err := ParseDefFile(defPath)
	if err != nil {
		t.Fatalf("ParseDefFile failed: %v", err)
	}

	if info.Category != CategoryFighter {
		t.Errorf("Expected category fighters, got %s", info.Category)
	}
	if info.DisplayName != "Kung Fu Man" {
		t.Errorf("Expected displayname 'Kung Fu Man', got '%s'", info.DisplayName)
	}
	if info.Author != "Elecbyte" {
		t.Errorf("Expected author 'Elecbyte', got '%s'", info.Author)
	}
	if len(info.FoundURLs) == 0 || info.FoundURLs[0] != "http://mugen.example.com/kfm" {
		t.Errorf("URL scraping failed: %+v", info.FoundURLs)
	}
}

func TestIngestZipArchive(t *testing.T) {
	vaultDir, err := os.MkdirTemp("", "vault_ingest_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(vaultDir)

	// Init vault manifest
	manifest := &VaultManifest{
		Version: "1.0",
		ID:      "vault-123",
		Name:    "Ingest Test Vault",
		Assets:  make(map[string]VaultAsset),
	}
	_ = SaveManifest(vaultDir, manifest)

	// Create dummy character zip
	zipPath := filepath.Join(vaultDir, "ryu_package.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}

	zw := zip.NewWriter(zipFile)
	defEntry, _ := zw.Create("ryu/ryu.def")
	_, _ = defEntry.Write([]byte(`[Info]
name = "Ryu"
displayname = "Ryu Hoshi"
author = "CapcomDev"
mugenversion = "1.1"

[Files]
sprite = ryu.sff
`))

	sffEntry, _ := zw.Create("ryu/ryu.sff")
	_, _ = sffEntry.Write([]byte("ElecbyteSpr\x00\x00\x01\x00\x01"))

	readmeEntry, _ := zw.Create("ryu/readme.txt")
	_, _ = readmeEntry.Write([]byte("Downloaded from https://capcom.com/fighters\nEnjoy!"))

	_ = zw.Close()
	_ = zipFile.Close()

	// Ingest zip
	result, err := IngestPath(vaultDir, zipPath, "ryu_package.zip")
	if err != nil {
		t.Fatalf("IngestPath failed: %v", err)
	}

	if result.ImportedCount != 1 {
		t.Fatalf("Expected 1 imported asset, got %d", result.ImportedCount)
	}

	loaded, err := LoadManifest(vaultDir)
	if err != nil {
		t.Fatal(err)
	}

	var foundAsset *VaultAsset
	for _, a := range loaded.Assets {
		if a.DisplayName == "Ryu Hoshi" {
			foundAsset = &a
			break
		}
	}

	if foundAsset == nil {
		t.Fatal("Ingested asset not found in manifest")
	}

	if foundAsset.Author != "CapcomDev" {
		t.Errorf("Expected author 'CapcomDev', got '%s'", foundAsset.Author)
	}
	if foundAsset.SourceURL != "https://capcom.com/fighters" {
		t.Errorf("Expected source URL 'https://capcom.com/fighters', got '%s'", foundAsset.SourceURL)
	}
}

func TestLinkAssetToProject(t *testing.T) {
	tempRoot, err := os.MkdirTemp("", "link_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempRoot)

	vaultDir := filepath.Join(tempRoot, "vault")
	projectDir := filepath.Join(tempRoot, "project")

	_ = os.MkdirAll(filepath.Join(vaultDir, "chars", "kfm"), 0755)
	_ = os.WriteFile(filepath.Join(vaultDir, "chars", "kfm", "kfm.def"), []byte("[Info]\nname = kfm"), 0644)

	_ = os.MkdirAll(filepath.Join(projectDir, "data"), 0755)
	_ = os.WriteFile(filepath.Join(projectDir, "data", "select.def"), []byte("[Characters]\n[ExtraStages]\n"), 0644)

	err = LinkAssetToProject(projectDir, vaultDir, "chars/kfm", LinkStrategySymlink)
	if err != nil {
		t.Fatalf("LinkAssetToProject failed: %v", err)
	}

	// Verify project character folder exists
	destDef := filepath.Join(projectDir, "chars", "kfm", "kfm.def")
	if _, err := os.Stat(destDef); err != nil {
		t.Fatalf("Linked character file not accessible: %v", err)
	}

	// Verify select.def has kfm
	selectDefData, err := os.ReadFile(filepath.Join(projectDir, "data", "select.def"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsSubstring(string(selectDefData), "kfm") {
		t.Errorf("kfm not found in select.def:\n%s", string(selectDefData))
	}

	// Test unlinking
	err = RemoveAssetFromProject(projectDir, "chars/kfm")
	if err != nil {
		t.Fatalf("RemoveAssetFromProject failed: %v", err)
	}

	// Verify vault asset is completely unharmed
	if _, err := os.Stat(filepath.Join(vaultDir, "chars", "kfm", "kfm.def")); err != nil {
		t.Fatalf("Vault asset was deleted on project unlinking!")
	}
}

func TestIngestRecursiveFolderWithArchives(t *testing.T) {
	tempRoot, err := os.MkdirTemp("", "bulk_ingest_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempRoot)

	vaultDir := filepath.Join(tempRoot, "vault")
	downloadsDir := filepath.Join(tempRoot, "downloads")
	subfolderDir := filepath.Join(downloadsDir, "batch_01")

	_ = os.MkdirAll(vaultDir, 0755)
	_ = os.MkdirAll(subfolderDir, 0755)

	manifest := &VaultManifest{
		Version: "1.0",
		ID:      "vault-bulk",
		Name:    "Bulk Test Vault",
		Assets:  make(map[string]VaultAsset),
	}
	_ = SaveManifest(vaultDir, manifest)

	// Create root zip: cammy.zip
	cammyZipPath := filepath.Join(downloadsDir, "cammy.zip")
	cammyZip, _ := os.Create(cammyZipPath)
	zw1 := zip.NewWriter(cammyZip)
	f1, _ := zw1.Create("cammy/cammy.def")
	_, _ = f1.Write([]byte("[Info]\nname = Cammy\ndisplayname = Cammy White\nauthor = Pots"))
	_ = zw1.Close()
	_ = cammyZip.Close()

	// Create nested zip in batch_01: spiderman.zip
	spideyZipPath := filepath.Join(subfolderDir, "spiderman.zip")
	spideyZip, _ := os.Create(spideyZipPath)
	zw2 := zip.NewWriter(spideyZip)
	f2, _ := zw2.Create("spiderman/spiderman.def")
	_, _ = f2.Write([]byte("[Info]\nname = Spiderman\ndisplayname = Spider-Man\nauthor = MarvelDev"))
	_ = zw2.Close()
	_ = spideyZip.Close()

	// Run Ingest on the entire parent folder downloadsDir
	res, err := IngestPath(vaultDir, downloadsDir, "all_chars")
	if err != nil {
		t.Fatalf("IngestPath on folder with nested archives failed: %v", err)
	}

	if res.ImportedCount != 2 {
		t.Fatalf("Expected 2 imported assets, got %d", res.ImportedCount)
	}

	loaded, err := LoadManifest(vaultDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(loaded.Assets) != 2 {
		t.Errorf("Expected 2 assets in vault manifest, got %d", len(loaded.Assets))
	}
}

func containsSubstring(str, substr string) bool {
	return filepath.Clean(str) != "" && len(str) >= len(substr) && (str == substr || len(str) > 0 && len(substr) > 0 && len(str) >= len(substr) && (stringContains(str, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

