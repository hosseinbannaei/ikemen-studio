package project

import (
	"bytes"
	"debug/buildinfo"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectLegacyEngineVersion analyzes a legacy game directory to identify its baseline engine version
func DetectLegacyEngineVersion(dir string) string {
	if dir == "" {
		return "nightly"
	}

	// 1. Check folder name for dev/nightly hints
	dirBase := strings.ToLower(filepath.Base(dir))
	if strings.Contains(dirBase, "nightly") || strings.Contains(dirBase, "dev-") || strings.Contains(dirBase, "-dev") {
		return "nightly"
	}

	// 2. Check version.txt if present
	for _, vFile := range []string{"version.txt", "VERSION.txt", "VERSION", "save/version.txt"} {
		vPath := filepath.Join(dir, vFile)
		if data, err := os.ReadFile(vPath); err == nil {
			content := strings.TrimSpace(string(data))
			if content != "" {
				return normalizeVersionTag(content)
			}
		}
	}

	// 3. Scan engine executable if present in legacy directory
	exeCandidates := []string{
		"Ikemen_GO_Linux", "Ikemen_GO.exe", "Ikemen_GO", "Ikemen_GO.command",
		"Ikemen_GO_x86_64", "Ikemen_GO_arm64", "Ikemen_GO_mac", "Ikemen.exe", "ikemen.exe", "ikemen",
	}

	for _, exeName := range exeCandidates {
		exePath := filepath.Join(dir, exeName)
		if fi, err := os.Stat(exePath); err == nil && !fi.IsDir() && fi.Size() > 0 {
			if ver := extractVersionFromBinary(exePath); ver != "" {
				return ver
			}
		}
	}

	// 4. Scan system.def or select.def
	for _, defFile := range []string{"data/system.def", "system.def"} {
		defPath := filepath.Join(dir, defFile)
		if data, err := os.ReadFile(defPath); err == nil {
			str := string(data)
			re := regexp.MustCompile(`(?i)(?:ikemenversion|version|mugenversion)\s*=\s*([^\r\n]+)`)
			if matches := re.FindStringSubmatch(str); len(matches) > 1 {
				val := strings.TrimSpace(matches[1])
				if strings.Contains(val, "nightly") || strings.Contains(val, "dev") {
					return "nightly"
				} else if strings.Contains(val, "0.99") {
					return "v0.99.0"
				} else if strings.Contains(val, "0.98") {
					return "v0.98.2"
				}
			}
		}
	}

	// 5. Check for presence of modern features (gltf 3D, zss, shaders)
	hasModernFeatures := false
	_ = filepath.Walk(filepath.Join(dir, "data"), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".zss" || ext == ".vert" || ext == ".frag" || ext == ".gltf" || ext == ".glb" {
				hasModernFeatures = true
			}
		}
		return nil
	})

	if hasModernFeatures {
		return "nightly"
	}

	return "v0.99.0"
}

func extractVersionFromBinary(exePath string) string {
	// First, try standard Go buildinfo inspection
	if bi, err := buildinfo.ReadFile(exePath); err == nil && bi != nil {
		// Check for devel / nightly
		if bi.Main.Version == "(devel)" || bi.Main.Version == "devel" || bi.Main.Version == "" {
			// Check dependencies for modern nightly packages (e.g. vulkan, gltf)
			for _, dep := range bi.Deps {
				if strings.Contains(dep.Path, "vulkan") || strings.Contains(dep.Path, "gltf") || strings.Contains(dep.Path, "ggpo") {
					return "nightly"
				}
			}
			return "nightly"
		}
		if bi.Main.Version != "" {
			return normalizeVersionTag(bi.Main.Version)
		}
	}

	file, err := os.Open(exePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Read up to 8MB of the binary to search for version string patterns
	buf := make([]byte, 8*1024*1024)
	n, _ := file.Read(buf)
	data := buf[:n]

	// Check nightly / devel indicators FIRST
	if bytes.Contains(data, []byte("nightly")) ||
		bytes.Contains(data, []byte("Nightly")) ||
		bytes.Contains(data, []byte("(devel)")) ||
		bytes.Contains(data, []byte("github.com/Eiton/vulkan")) ||
		bytes.Contains(data, []byte("github.com/qmuntal/gltf")) {
		return "nightly"
	}

	if bytes.Contains(data, []byte("v1.0.0")) {
		return "v1.0.0-rc.3"
	}
	if bytes.Contains(data, []byte("v0.99.0")) {
		return "v0.99.0"
	}
	if bytes.Contains(data, []byte("v0.98.2")) {
		return "v0.98.2"
	}

	return ""
}

func normalizeVersionTag(raw string) string {
	raw = strings.TrimSpace(raw)
	if strings.EqualFold(raw, "nightly") || strings.EqualFold(raw, "dev") || strings.EqualFold(raw, "develop") {
		return "nightly"
	}
	if !strings.HasPrefix(raw, "v") && (strings.Contains(raw, "0.") || strings.Contains(raw, "1.")) {
		return "v" + raw
	}
	return raw
}

