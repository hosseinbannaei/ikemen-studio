package engine

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestParseAssetPlatform(t *testing.T) {
	tests := []struct {
		filename string
		wantOS   string
		wantArch string
	}{
		{"Ikemen_GO-v0.99.0-linux-x86_64.tar.gz", "linux", "amd64"},
		{"Ikemen_GO-v0.99.0-windows-x64.zip", "windows", "amd64"},
		{"Ikemen_GO-v0.99.0-macOS-arm64.zip", "darwin", "arm64"},
		{"Ikemen_GO-v0.99.0-linux-arm64.tar.gz", "linux", "arm64"},
		{"Ikemen_GO-v0.99.0-win32.zip", "windows", "386"},
	}

	for _, tt := range tests {
		gotOS, gotArch := ParseAssetPlatform(tt.filename)
		if gotOS != tt.wantOS || gotArch != tt.wantArch {
			t.Errorf("ParseAssetPlatform(%q) = (%q, %q), want (%q, %q)", tt.filename, gotOS, gotArch, tt.wantOS, tt.wantArch)
		}
	}
}

func TestEngineManagerOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-engine-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock engine directory
	versionDir := filepath.Join(tmpDir, "v0.99.0")
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		t.Fatal(err)
	}

	binaryName := "Ikemen_GO"
	if runtime.GOOS == "windows" {
		binaryName = "Ikemen_GO.exe"
	}
	binPath := filepath.Join(versionDir, binaryName)
	if err := os.WriteFile(binPath, []byte("mock binary"), 0755); err != nil {
		t.Fatal(err)
	}

	engines, err := ListInstalledEngines(tmpDir)
	if err != nil {
		t.Fatalf("ListInstalledEngines failed: %v", err)
	}

	if len(engines) != 1 {
		t.Fatalf("expected 1 installed engine, got %d", len(engines))
	}

	if engines[0].Version != "v0.99.0" {
		t.Errorf("expected version v0.99.0, got %s", engines[0].Version)
	}

	if engines[0].ExecutablePath != binPath {
		t.Errorf("expected executable %s, got %s", binPath, engines[0].ExecutablePath)
	}

	// Test deletion
	if err := DeleteEngine(tmpDir, "v0.99.0"); err != nil {
		t.Fatalf("DeleteEngine failed: %v", err)
	}

	enginesAfter, err := ListInstalledEngines(tmpDir)
	if err != nil {
		t.Fatalf("ListInstalledEngines after delete failed: %v", err)
	}
	if len(enginesAfter) != 0 {
		t.Errorf("expected 0 engines after deletion, got %d", len(enginesAfter))
	}
}

func TestExtractZipArchive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ikemen-zip-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "test.zip")
	destDir := filepath.Join(tmpDir, "extracted")

	// Create test zip
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(zipFile)
	f, err := w.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write([]byte("hello ikemen")); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	_ = zipFile.Close()

	if err := extractZip(zipPath, destDir); err != nil {
		t.Fatalf("extractZip failed: %v", err)
	}

	extractedFile := filepath.Join(destDir, "test.txt")
	data, err := os.ReadFile(extractedFile)
	if err != nil {
		t.Fatalf("failed to read extracted file: %v", err)
	}
	if string(data) != "hello ikemen" {
		t.Errorf("expected content 'hello ikemen', got %q", string(data))
	}
}
