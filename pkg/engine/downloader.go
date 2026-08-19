package engine

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const GitHubReleasesURL = "https://api.github.com/repos/ikemen-engine/Ikemen-GO/releases"

// GitHubReleaseResponse maps GitHub's API response
type githubReleaseResponse struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Prerelease  bool   `json:"prerelease"`
	Draft       bool   `json:"draft"`
	PublishedAt string `json:"published_at"`
	Body        string `json:"body"`
	HTMLURL     string `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		Size               int64  `json:"size"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// FetchReleases retrieves release list from GitHub
func FetchReleases() ([]ReleaseInfo, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", GitHubReleasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Ikemen-Studio")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var ghReleases []githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&ghReleases); err != nil {
		return nil, fmt.Errorf("failed to decode releases json: %w", err)
	}

	var results []ReleaseInfo
	for _, gh := range ghReleases {
		if gh.Draft {
			continue
		}
		pubTime, _ := time.Parse(time.RFC3339, gh.PublishedAt)
		var assets []ReleaseAsset
		for _, a := range gh.Assets {
			osName, arch := ParseAssetPlatform(a.Name)
			assets = append(assets, ReleaseAsset{
				Name:        a.Name,
				Size:        a.Size,
				DownloadURL: a.BrowserDownloadURL,
				OS:          osName,
				Arch:        arch,
			})
		}

		results = append(results, ReleaseInfo{
			Tag:          gh.TagName,
			Name:         gh.Name,
			PublishedAt:  pubTime,
			IsPrerelease: gh.Prerelease,
			Body:         gh.Body,
			HTMLURL:      gh.HTMLURL,
			Assets:       assets,
		})
	}

	return results, nil
}

// ParseAssetPlatform detects OS and Arch from the filename
func ParseAssetPlatform(filename string) (osName string, arch string) {
	lower := strings.ToLower(filename)

	if strings.Contains(lower, "linux") {
		osName = "linux"
	} else if strings.Contains(lower, "win") || strings.Contains(lower, "windows") {
		osName = "windows"
	} else if strings.Contains(lower, "mac") || strings.Contains(lower, "darwin") || strings.Contains(lower, "osx") {
		osName = "darwin"
	}

	if strings.Contains(lower, "x86_64") || strings.Contains(lower, "amd64") || strings.Contains(lower, "x64") || strings.Contains(lower, "64bit") || strings.Contains(lower, "win64") {
		arch = "amd64"
	} else if strings.Contains(lower, "arm64") || strings.Contains(lower, "aarch64") {
		arch = "arm64"
	} else if strings.Contains(lower, "386") || strings.Contains(lower, "x86") || strings.Contains(lower, "32bit") || strings.Contains(lower, "win32") {
		arch = "386"
	}

	return osName, arch
}

// FindBestAsset finds matching asset for current host OS and architecture
func FindBestAsset(release ReleaseInfo) (*ReleaseAsset, error) {
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH

	// Exact match first
	for i := range release.Assets {
		a := &release.Assets[i]
		if a.OS == targetOS && (a.Arch == targetArch || (targetArch == "amd64" && a.Arch == "")) {
			return a, nil
		}
	}

	// Fallback to OS match
	for i := range release.Assets {
		a := &release.Assets[i]
		if a.OS == targetOS {
			return a, nil
		}
	}

	return nil, fmt.Errorf("no compatible asset found for OS %s / Arch %s in release %s", targetOS, targetArch, release.Tag)
}

type progressWriter struct {
	ctx        context.Context
	total      int64
	current    int64
	version    string
	onProgress func(DownloadProgress)
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	select {
	case <-pw.ctx.Done():
		return 0, pw.ctx.Err()
	default:
	}

	n := len(p)
	pw.current += int64(n)
	if pw.onProgress != nil {
		percent := 0.0
		if pw.total > 0 {
			percent = float64(pw.current) / float64(pw.total) * 100
		}
		pw.onProgress(DownloadProgress{
			Version:         pw.version,
			Percent:         percent,
			DownloadedBytes: pw.current,
			TotalBytes:      pw.total,
			Status:          "downloading",
		})
	}
	return n, nil
}

// DownloadAndExtractEngine downloads the asset into a temp folder and renames it to enginesDir/version upon completion
func DownloadAndExtractEngine(ctx context.Context, asset ReleaseAsset, version string, enginesDir string, onProgress func(DownloadProgress)) error {
	if err := os.MkdirAll(enginesDir, 0755); err != nil {
		return fmt.Errorf("failed to create engines directory: %w", err)
	}

	destDir := filepath.Join(enginesDir, version)
	tmpExtractDir := filepath.Join(enginesDir, fmt.Sprintf("%s.tmp-%d", version, time.Now().UnixNano()))

	// Clean up temp extraction dir if process is cancelled or fails
	defer func() {
		if _, err := os.Stat(tmpExtractDir); err == nil {
			_ = os.RemoveAll(tmpExtractDir)
		}
	}()

	tmpFile, err := os.CreateTemp("", "ikemen-download-*"+filepath.Ext(asset.Name))
	if err != nil {
		return fmt.Errorf("failed to create temporary download file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	req, err := http.NewRequestWithContext(ctx, "GET", asset.DownloadURL, nil)
	if err != nil {
		return fmt.Errorf("failed to build download request: %w", err)
	}
	req.Header.Set("User-Agent", "Ikemen-Studio")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		return fmt.Errorf("failed to download engine asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}

	pw := &progressWriter{
		ctx:        ctx,
		total:      resp.ContentLength,
		version:    version,
		onProgress: onProgress,
	}

	if _, err := io.Copy(tmpFile, io.TeeReader(resp.Body, pw)); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		return fmt.Errorf("failed writing downloaded asset: %w", err)
	}

	_ = tmpFile.Close()

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	if onProgress != nil {
		onProgress(DownloadProgress{
			Version: version,
			Percent: 100,
			Status:  "extracting",
		})
	}

	// Create temp extract directory
	if err := os.MkdirAll(tmpExtractDir, 0755); err != nil {
		return fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}

	// Extract archive into tmpExtractDir
	lowerName := strings.ToLower(asset.Name)
	if strings.HasSuffix(lowerName, ".zip") {
		if err := extractZip(tmpFile.Name(), tmpExtractDir); err != nil {
			return fmt.Errorf("failed extracting zip: %w", err)
		}
	} else if strings.HasSuffix(lowerName, ".tar.gz") || strings.HasSuffix(lowerName, ".tgz") {
		if err := extractTarGz(tmpFile.Name(), tmpExtractDir); err != nil {
			return fmt.Errorf("failed extracting tar.gz: %w", err)
		}
	} else {
		return fmt.Errorf("unsupported archive format for %s", asset.Name)
	}

	// Flatten single top-level directory if archive contained one
	if err := flattenSingleDirectory(tmpExtractDir); err != nil {
		return fmt.Errorf("failed organizing extracted files: %w", err)
	}

	// Make executables executable
	_ = fixPermissions(tmpExtractDir)

	select {
	case <-ctx.Done():
		return context.Canceled
	default:
	}

	// If destDir already exists (e.g. overwriting/updating), remove it
	_ = os.RemoveAll(destDir)

	// Move tmpExtractDir to final destDir
	if err := os.Rename(tmpExtractDir, destDir); err != nil {
		return fmt.Errorf("failed moving extracted engine to destination: %w", err)
	}

	if onProgress != nil {
		onProgress(DownloadProgress{
			Version: version,
			Percent: 100,
			Status:  "completed",
		})
	}

	return nil
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		targetPath := filepath.Join(dest, f.Name)
		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(targetPath) != filepath.Clean(dest) {
			continue
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(targetPath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dest, header.Name)
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(dest)+string(os.PathSeparator)) && filepath.Clean(targetPath) != filepath.Clean(dest) {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
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

func flattenSingleDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// If there is only 1 directory and nothing else, move its contents up
	if len(entries) == 1 && entries[0].IsDir() {
		subDir := filepath.Join(dir, entries[0].Name())
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			return err
		}
		for _, se := range subEntries {
			oldPath := filepath.Join(subDir, se.Name())
			newPath := filepath.Join(dir, se.Name())
			if err := os.Rename(oldPath, newPath); err != nil {
				return err
			}
		}
		_ = os.Remove(subDir)
	}
	return nil
}

func fixPermissions(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			name := strings.ToLower(info.Name())
			if name == "ikemen_go" || name == "ikemen_go.exe" || strings.HasSuffix(name, ".sh") || (runtime.GOOS != "windows" && (info.Mode()&0111 != 0 || strings.Contains(name, "ikemen"))) {
				_ = os.Chmod(path, 0755)
			}
		}
		return nil
	})
}
