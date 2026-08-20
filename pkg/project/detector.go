package project

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DetectLegacyEngineVersion analyzes a legacy game directory to identify its baseline engine version
func DetectLegacyEngineVersion(dir string) string {
	if dir == "" {
		return "v0.99.0"
	}

	// 1. Check version.txt if present
	for _, vFile := range []string{"version.txt", "VERSION.txt", "VERSION", "save/version.txt"} {
		vPath := filepath.Join(dir, vFile)
		if data, err := os.ReadFile(vPath); err == nil {
			content := strings.TrimSpace(string(data))
			if content != "" {
				return normalizeVersionTag(content)
			}
		}
	}

	// 2. Scan engine executable if present in legacy directory
	for _, exeName := range []string{"Ikemen_GO.exe", "Ikemen_GO", "Ikemen.exe", "ikemen.exe"} {
		exePath := filepath.Join(dir, exeName)
		if fi, err := os.Stat(exePath); err == nil && !fi.IsDir() && fi.Size() > 0 {
			if ver := extractVersionFromBinary(exePath); ver != "" {
				return ver
			}
		}
	}

	// 3. Scan system.def or select.def
	for _, defFile := range []string{"data/system.def", "system.def"} {
		defPath := filepath.Join(dir, defFile)
		if data, err := os.ReadFile(defPath); err == nil {
			str := string(data)
			re := regexp.MustCompile(`(?i)(?:ikemenversion|version|mugenversion)\s*=\s*([^\r\n]+)`)
			if matches := re.FindStringSubmatch(str); len(matches) > 1 {
				val := strings.TrimSpace(matches[1])
				if strings.Contains(val, "0.99") {
					return "v0.99.0"
				} else if strings.Contains(val, "0.98") {
					return "v0.98.2"
				}
			}
		}
	}

	// 4. Check for presence of zss files (introduced in modern Ikemen)
	hasZss := false
	_ = filepath.Walk(filepath.Join(dir, "data"), func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".zss") {
			hasZss = true
		}
		return nil
	})

	if hasZss {
		return "v0.99.0"
	}

	return "v0.99.0"
}

func extractVersionFromBinary(exePath string) string {
	file, err := os.Open(exePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Read up to 8MB of the binary to search for version string patterns
	buf := make([]byte, 8*1024*1024)
	n, _ := file.Read(buf)
	data := buf[:n]

	// Patterns: v0.99.0, v0.98.2, v1.0.0-rc.3, nightly
	if bytes.Contains(data, []byte("v1.0.0")) {
		return "v1.0.0-rc.3"
	}
	if bytes.Contains(data, []byte("v0.99.0")) || bytes.Contains(data, []byte("0.99.0")) {
		return "v0.99.0"
	}
	if bytes.Contains(data, []byte("v0.98.2")) || bytes.Contains(data, []byte("0.98.2")) {
		return "v0.98.2"
	}
	if bytes.Contains(data, []byte("nightly")) || bytes.Contains(data, []byte("Nightly")) {
		return "nightly"
	}

	return ""
}

func normalizeVersionTag(raw string) string {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "v") && (strings.Contains(raw, "0.") || strings.Contains(raw, "1.")) {
		return "v" + raw
	}
	return raw
}
