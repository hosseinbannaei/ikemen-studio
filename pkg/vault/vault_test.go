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

func TestIngestStoryboardSkipping(t *testing.T) {
	vaultDir, err := os.MkdirTemp("", "vault_sb_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(vaultDir)

	manifest := &VaultManifest{
		Version: "1.0",
		ID:      "vault-sb",
		Name:    "Storyboard Skip Test Vault",
		Assets:  make(map[string]VaultAsset),
	}
	_ = SaveManifest(vaultDir, manifest)

	// Create a zip with kfm.def AND ending.def AND intro.def
	zipPath := filepath.Join(vaultDir, "kfm_with_storyboard.zip")
	zipFile, _ := os.Create(zipPath)
	zw := zip.NewWriter(zipFile)

	// ending.def is a storyboard
	f1, _ := zw.Create("kfm/ending.def")
	_, _ = f1.Write([]byte("[SceneDef]\nspr = ending.sff\n[Scene 0]\ntime = 100"))

	// intro.def is a storyboard
	f2, _ := zw.Create("kfm/intro.def")
	_, _ = f2.Write([]byte("[SceneDef]\nspr = intro.sff\n[Scene 0]\ntime = 100"))

	// kfm.def is the real character
	f3, _ := zw.Create("kfm/kfm.def")
	_, _ = f3.Write([]byte("[Info]\nname = kfm\ndisplayname = Kung Fu Man\nauthor = Elecbyte\n[Files]\ncns = kfm.cns\ncmd = kfm.cmd"))

	_ = zw.Close()
	_ = zipFile.Close()

	res, err := IngestPath(vaultDir, zipPath, "kfm_archive")
	if err != nil {
		t.Fatalf("IngestPath failed: %v", err)
	}

	if res.ImportedCount != 1 {
		t.Fatalf("Expected 1 imported character (ignoring storyboards), got %d", res.ImportedCount)
	}

	if res.DetectedAssets[0].DisplayName != "Kung Fu Man" {
		t.Errorf("Expected asset name 'Kung Fu Man', got '%s'", res.DetectedAssets[0].DisplayName)
	}
}

func TestIngestLooseFilesMultiArchive(t *testing.T) {
	vaultDir, err := os.MkdirTemp("", "vault_loose_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(vaultDir)

	manifest := &VaultManifest{
		Version: "1.0",
		ID:      "vault-loose",
		Name:    "Loose Archive Test Vault",
		Assets:  make(map[string]VaultAsset),
	}
	_ = SaveManifest(vaultDir, manifest)

	// Create 2 loose archives (files directly at root, no enclosing folder)
	cammyZipPath := filepath.Join(vaultDir, "cammy_pots.zip")
	z1, _ := os.Create(cammyZipPath)
	zw1 := zip.NewWriter(z1)
	f1, _ := zw1.Create("cammy_pots.def")
	_, _ = f1.Write([]byte("[Info]\nname = Cammy\ndisplayname = Cammy\nauthor = Pots\n[Files]\ncns = cammy.cns\nsprite = cammy.sff"))
	f1c, _ := zw1.Create("cammy.cns")
	_, _ = f1c.Write([]byte("; cns"))
	_ = zw1.Close()
	_ = z1.Close()

	sakuraZipPath := filepath.Join(vaultDir, "sakura_pots.zip")
	z2, _ := os.Create(sakuraZipPath)
	zw2 := zip.NewWriter(z2)
	f2, _ := zw2.Create("sakura_pots.def")
	_, _ = f2.Write([]byte("[Info]\nname = Sakura\ndisplayname = Sakura\nauthor = Pots\n[Files]\ncns = sakura.cns\nsprite = sakura.sff"))
	f2c, _ := zw2.Create("sakura.cns")
	_, _ = f2c.Write([]byte("; cns"))
	_ = zw2.Close()
	_ = z2.Close()

	// Ingest both loose archives together in one batch
	res, err := IngestMultiple(vaultDir, []string{cammyZipPath, sakuraZipPath}, "loose_batch")
	if err != nil {
		t.Fatalf("IngestMultiple failed: %v", err)
	}

	if res.ImportedCount != 2 {
		t.Fatalf("Expected 2 imported assets, got %d", res.ImportedCount)
	}

	// Verify that cammy folder does NOT contain sakura files
	cammyDir := filepath.Join(vaultDir, "chars", "cammy_pots")
	if _, err := os.Stat(filepath.Join(cammyDir, "sakura_pots.def")); err == nil {
		t.Errorf("Cammy folder was polluted with Sakura files!")
	}

	// Verify that sakura folder does NOT contain cammy files
	sakuraDir := filepath.Join(vaultDir, "chars", "sakura_pots")
	if _, err := os.Stat(filepath.Join(sakuraDir, "cammy_pots.def")); err == nil {
		t.Errorf("Sakura folder was polluted with Cammy files!")
	}
}

func TestIngestDeduplicationInPlace(t *testing.T) {
	vaultDir, err := os.MkdirTemp("", "vault_dedup_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(vaultDir)

	manifest := &VaultManifest{
		Version: "1.0",
		ID:      "vault-dedup",
		Name:    "Dedup Test Vault",
		Assets:  make(map[string]VaultAsset),
	}
	_ = SaveManifest(vaultDir, manifest)

	// Create test zip
	zipPath := filepath.Join(vaultDir, "ryu.zip")
	z, _ := os.Create(zipPath)
	zw := zip.NewWriter(z)
	f, _ := zw.Create("ryu/ryu.def")
	_, _ = f.Write([]byte("[Info]\nname = Ryu\ndisplayname = Ryu\nauthor = Capcom\n[Files]\ncns = ryu.cns"))
	_ = zw.Close()
	_ = z.Close()

	// Ingest first time
	_, err = IngestPath(vaultDir, zipPath, "batch_1")
	if err != nil {
		t.Fatalf("First ingest failed: %v", err)
	}

	// Re-ingest second time (should update in-place without generating ryu_1234)
	res2, err := IngestPath(vaultDir, zipPath, "batch_2")
	if err != nil {
		t.Fatalf("Second ingest failed: %v", err)
	}

	if res2.ImportedCount != 1 {
		t.Fatalf("Expected 1 asset imported, got %d", res2.ImportedCount)
	}

	loaded, _ := LoadManifest(vaultDir)
	if len(loaded.Assets) != 1 {
		t.Fatalf("Expected exactly 1 asset in manifest (no duplicate timestamps), got %d: %+v", len(loaded.Assets), loaded.Assets)
	}

	if _, exists := loaded.Assets["chars/ryu"]; !exists {
		t.Errorf("Expected key 'chars/ryu', got assets: %+v", loaded.Assets)
	}
}

func TestCleanAndRepairVault(t *testing.T) {
	tempBase, err := os.MkdirTemp("", "vault_repair_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempBase)

	vaultDir := filepath.Join(tempBase, "default")
	_ = os.MkdirAll(filepath.Join(vaultDir, "chars", "kfm"), 0755)
	_ = os.MkdirAll(filepath.Join(vaultDir, "chars", "kfm_9999"), 0755) // Duplicate timestamp folder

	// Create kfm def in both
	_ = os.WriteFile(filepath.Join(vaultDir, "chars", "kfm", "kfm.def"), []byte("[Info]\nname = kfm\ndisplayname = Kung Fu Man\nauthor = Elecbyte\n[Files]\ncns=kfm.cns"), 0644)
	_ = os.WriteFile(filepath.Join(vaultDir, "chars", "kfm_9999", "kfm.def"), []byte("[Info]\nname = kfm\ndisplayname = Kung Fu Man\nauthor = Elecbyte\n[Files]\ncns=kfm.cns"), 0644)

	// Create a cross-contaminated foreign folder inside kfm: kfm/Ryu/ryu.def
	_ = os.MkdirAll(filepath.Join(vaultDir, "chars", "kfm", "Ryu"), 0755)
	_ = os.WriteFile(filepath.Join(vaultDir, "chars", "kfm", "Ryu", "ryu.def"), []byte("[Info]\nname = Ryu\ndisplayname = Ryu\nauthor = Capcom\n[Files]\ncns=ryu.cns"), 0644)

	manifest := &VaultManifest{
		Version: "1.0",
		ID:      "vault-repair-test",
		Name:    "Repair Vault",
		Assets: map[string]VaultAsset{
			"chars/kfm": {
				Key:         "chars/kfm",
				Category:    CategoryFighter,
				DisplayName: "Kung Fu Man",
				Author:      "Elecbyte",
			},
			"chars/kfm_9999": {
				Key:         "chars/kfm_9999",
				Category:    CategoryFighter,
				DisplayName: "Kung Fu Man",
				Author:      "Elecbyte",
			},
			"chars/ghost_fighter": {
				Key:         "chars/ghost_fighter",
				Category:    CategoryFighter,
				DisplayName: "Ghost",
			},
		},
	}
	_ = SaveManifest(vaultDir, manifest)

	vm := &VaultManager{}
	_, err = vm.RegisterVault(vaultDir)
	if err != nil {
		t.Fatal(err)
	}

	report, err := vm.CleanAndRepairVault("vault-repair-test")

	if err != nil {
		t.Fatalf("CleanAndRepairVault failed: %v", err)
	}

	if report.RemovedDuplicates == 0 {
		t.Errorf("Expected duplicates to be removed, got %d", report.RemovedDuplicates)
	}
	if report.CleanedContaminations == 0 {
		t.Errorf("Expected contaminated Ryu folder to be removed, got %d", report.CleanedContaminations)
	}
	if report.PrunedMissing == 0 {
		t.Errorf("Expected ghost_fighter to be pruned, got %d", report.PrunedMissing)
	}

	// Verify kfm_9999 is deleted from disk
	if _, err := os.Stat(filepath.Join(vaultDir, "chars", "kfm_9999")); !os.IsNotExist(err) {
		t.Errorf("kfm_9999 still exists on disk!")
	}

	// Verify foreign Ryu folder is deleted from kfm
	if _, err := os.Stat(filepath.Join(vaultDir, "chars", "kfm", "Ryu")); !os.IsNotExist(err) {
		t.Errorf("Contaminated Ryu folder still exists inside kfm!")
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


