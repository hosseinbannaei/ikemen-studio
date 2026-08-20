package vault

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractAndCachePortrait attempts to find and extract Sprite 9000,0 or 9000,1 from an SFF file,
// or loose image files (portrait.png, icon.png, *.png), saves it as a PNG, and returns the relative path and base64 URI.
func ExtractAndCachePortrait(vaultDir, assetKey, sffPath string) (relPath string, base64URI string, err error) {
	safeName := strings.ReplaceAll(assetKey, "/", "_") + ".png"
	previewsDir := filepath.Join(vaultDir, ".previews")
	if err := os.MkdirAll(previewsDir, 0755); err != nil {
		return "", "", err
	}
	outPath := filepath.Join(previewsDir, safeName)
	relPath = filepath.Join(".previews", safeName)

	// If cached PNG already exists on disk, reuse immediately
	if cachedData, err := os.ReadFile(outPath); err == nil && len(cachedData) > 100 {
		return relPath, "data:image/png;base64," + base64.StdEncoding.EncodeToString(cachedData), nil
	}

	var img image.Image
	assetDir := filepath.Join(vaultDir, assetKey)

	// 1. Try SFF extraction
	if sffPath != "" {
		sffClean := strings.ReplaceAll(sffPath, "\\", "/")
		targetSff := sffClean
		if !filepath.IsAbs(targetSff) {
			targetSff = filepath.Join(assetDir, sffClean)
		}
		if _, statErr := os.Stat(targetSff); statErr != nil {
			targetSff = ResolvePathCaseInsensitive(assetDir, sffClean)
		}
		if targetSff != "" {
			if fileImg, parseErr := ParseSFFPortrait(targetSff); parseErr == nil && fileImg != nil {
				img = fileImg
			}
		}
	}

	// 2. If SFF failed or not present, search character directory for any SFF file
	if img == nil {
		anySff := findFirstExt(assetDir, ".sff")
		if anySff != "" {
			if fileImg, parseErr := ParseSFFPortrait(anySff); parseErr == nil && fileImg != nil {
				img = fileImg
			}
		}
	}

	// 3. If SFF failed, search character directory for loose portrait images
	if img == nil {
		img = findLooseImage(assetDir)
	}

	// 4. Fallback to procedural avatar if no image found
	if img == nil {
		img = generatePlaceholderAvatar(assetKey)
	}

	// Write PNG file
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", "", err
	}
	defer outFile.Close()

	var buf bytes.Buffer
	mw := io.MultiWriter(outFile, &buf)
	if err := png.Encode(mw, img); err != nil {
		return "", "", err
	}

	base64URI = "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
	return relPath, base64URI, nil
}

// ResolvePathCaseInsensitive traverses path components step by step to resolve case differences on Linux/macOS.
func ResolvePathCaseInsensitive(baseDir, relPath string) string {
	cleanRel := strings.ReplaceAll(relPath, "\\", "/")
	parts := strings.Split(filepath.Clean(cleanRel), "/")

	current := baseDir
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			current = filepath.Dir(current)
			continue
		}

		entries, err := os.ReadDir(current)
		if err != nil {
			return ""
		}

		found := false
		lowerPart := strings.ToLower(part)
		for _, e := range entries {
			if strings.ToLower(e.Name()) == lowerPart {
				current = filepath.Join(current, e.Name())
				found = true
				break
			}
		}
		if !found {
			return ""
		}
	}

	return current
}

func findLooseImage(dir string) image.Image {
	if _, err := os.Stat(dir); err != nil {
		return nil
	}

	// Prioritized candidates
	priorityNames := []string{"portrait.png", "big.png", "icon.png", "small.png", "preview.png", "face.png"}
	for _, name := range priorityNames {
		p := ResolvePathCaseInsensitive(dir, name)
		if p != "" {
			if f, err := os.Open(p); err == nil {
				defer f.Close()
				if decoded, _, err := image.Decode(f); err == nil {
					return decoded
				}
			}
		}
	}

	// Any png/jpg in the directory
	var foundImg image.Image
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && foundImg == nil {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				if f, err := os.Open(path); err == nil {
					defer f.Close()
					if decoded, _, err := image.Decode(f); err == nil {
						foundImg = decoded
						return io.EOF
					}
				}
			}
		}
		return nil
	})

	return foundImg
}


// ParseSFFPortrait parses an SFF file (v1 or v2) and extracts Sprite (9000,0) or (9000,1).
func ParseSFFPortrait(sffPath string) (image.Image, error) {
	data, err := os.ReadFile(sffPath)
	if err != nil {
		return nil, err
	}

	if len(data) < 32 {
		return nil, fmt.Errorf("file too small for SFF")
	}

	sig := string(data[:11])
	if !strings.HasPrefix(sig, "ElecbyteSpr") {
		return nil, fmt.Errorf("invalid SFF signature: %s", sig)
	}

	verHi := data[15]
	if verHi == 1 || verHi == 0 {
		return parseSFFv1(data)
	} else if verHi == 2 {
		return parseSFFv2(data)
	}

	return nil, fmt.Errorf("unsupported SFF version: %d", verHi)
}

// parseSFFv1 handles SFF v1.01 subfiles (PCX formatted)
func parseSFFv1(data []byte) (image.Image, error) {
	if len(data) < 512 {
		return nil, fmt.Errorf("SFF v1 data too short")
	}

	firstOffset := binary.LittleEndian.Uint32(data[24:28])
	currOffset := int(firstOffset)

	var (
		fallbackImg image.Image
		sharedPal   []color.Color
	)

	// Traverse linked list of sprites
	maxIter := 5000
	for currOffset > 0 && currOffset+32 <= len(data) && maxIter > 0 {
		maxIter--
		nextOffset := int(binary.LittleEndian.Uint32(data[currOffset : currOffset+4]))
		subLen := int(binary.LittleEndian.Uint32(data[currOffset+4 : currOffset+8]))
		group := binary.LittleEndian.Uint16(data[currOffset+12 : currOffset+14])
		imageNo := binary.LittleEndian.Uint16(data[currOffset+14 : currOffset+16])
		samePal := data[currOffset+18]

		pcxStart := currOffset + 32
		pcxEnd := pcxStart + subLen
		if pcxEnd > len(data) {
			pcxEnd = len(data)
		}

		if subLen > 128 && pcxStart < len(data) {
			pcxData := data[pcxStart:pcxEnd]
			decodedImg, pal, err := decodePCX(pcxData, sharedPal)
			if err == nil && decodedImg != nil {
				if samePal == 0 && len(pal) > 0 {
					sharedPal = pal
				}

				// Large portrait (9000, 1) is highest priority
				if group == 9000 && imageNo == 1 {
					return decodedImg, nil
				}
				// (9000, 0) or other sprite as fallback
				if group == 9000 && imageNo == 0 {
					fallbackImg = decodedImg
				} else if fallbackImg == nil {
					fallbackImg = decodedImg
				}
			}
		}

		if nextOffset == 0 || nextOffset == currOffset {
			break
		}
		currOffset = nextOffset
	}

	if fallbackImg != nil {
		return fallbackImg, nil
	}

	return nil, fmt.Errorf("no suitable portrait found in SFF v1")
}

// parseSFFv2 handles SFF v2.00 / 2.01 subfiles
func parseSFFv2(data []byte) (image.Image, error) {
	if len(data) < 512 {
		return nil, fmt.Errorf("SFF v2 data too short")
	}

	spriteOffset := int(binary.LittleEndian.Uint32(data[36:40]))
	totalSprites := int(binary.LittleEndian.Uint32(data[40:44]))

	if spriteOffset <= 0 || spriteOffset >= len(data) {
		return nil, fmt.Errorf("invalid SFF v2 sprite offset")
	}

	var fallbackImg image.Image

	for i := 0; i < totalSprites && (spriteOffset+28) <= len(data); i++ {
		group := binary.LittleEndian.Uint16(data[spriteOffset : spriteOffset+2])
		item := binary.LittleEndian.Uint16(data[spriteOffset+2 : spriteOffset+4])
		format := data[spriteOffset+12]
		dataOffset := int(binary.LittleEndian.Uint32(data[spriteOffset+16 : spriteOffset+20]))
		dataLen := int(binary.LittleEndian.Uint32(data[spriteOffset+20 : spriteOffset+24]))

		spriteOffset += 28

		// Check if PNG format (10=PNG8, 11=PNG24, 12=PNG32)
		if (group == 9000 && (item == 1 || item == 0)) || fallbackImg == nil {
			if dataOffset > 0 && dataOffset+dataLen <= len(data) && dataLen > 8 {
				raw := data[dataOffset : dataOffset+dataLen]
				if format >= 10 && format <= 12 || bytes.HasPrefix(raw, []byte("\x89PNG")) {
					img, err := png.Decode(bytes.NewReader(raw))
					if err == nil {
						if group == 9000 && item == 1 {
							return img, nil
						}
						if group == 9000 && item == 0 {
							fallbackImg = img
						} else if fallbackImg == nil {
							fallbackImg = img
						}
					}
				}
			}
		}
	}

	if fallbackImg != nil {
		return fallbackImg, nil
	}

	return nil, fmt.Errorf("no portrait found in SFF v2")
}

// decodePCX is a lightweight PCX 8-bit RLE decoder
func decodePCX(data []byte, fallbackPal []color.Color) (image.Image, []color.Color, error) {
	if len(data) < 128 {
		return nil, nil, fmt.Errorf("PCX data too short")
	}
	if data[0] != 0x0A {
		return nil, nil, fmt.Errorf("not a PCX file")
	}

	bpp := int(data[3])
	xmin := int(binary.LittleEndian.Uint16(data[4:6]))
	ymin := int(binary.LittleEndian.Uint16(data[6:8]))
	xmax := int(binary.LittleEndian.Uint16(data[8:10]))
	ymax := int(binary.LittleEndian.Uint16(data[10:12]))
	bytesPerLine := int(binary.LittleEndian.Uint16(data[66:68]))

	width := xmax - xmin + 1
	height := ymax - ymin + 1

	if width <= 0 || height <= 0 || width > 2048 || height > 2048 {
		return nil, nil, fmt.Errorf("invalid PCX dimensions: %dx%d", width, height)
	}

	// Read palette: 256 colors at end of PCX file (preceded by 0x0C byte)
	var palette []color.Color
	if len(data) >= 769 && data[len(data)-769] == 0x0C {
		palette = make([]color.Color, 256)
		palBytes := data[len(data)-768:]
		for i := 0; i < 256; i++ {
			r := palBytes[i*3]
			g := palBytes[i*3+1]
			b := palBytes[i*3+2]
			a := uint8(255)
			if i == 0 {
				// MUGEN index 0 is transparent by convention
				a = 0
			}
			palette[i] = color.RGBA{R: r, G: g, B: b, A: a}
		}
	} else if len(fallbackPal) == 256 {
		palette = fallbackPal
	} else {
		// Grayscale fallback
		palette = make([]color.Color, 256)
		for i := 0; i < 256; i++ {
			palette[i] = color.RGBA{R: uint8(i), G: uint8(i), B: uint8(i), A: 255}
		}
	}

	if bpp != 8 {
		// Non-8bit PCX fallback
		return nil, palette, fmt.Errorf("unsupported PCX bpp: %d", bpp)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	rleData := data[128:]
	if len(data) >= 769 && data[len(data)-769] == 0x0C {
		rleData = data[128 : len(data)-769]
	}

	rleIdx := 0
	for y := 0; y < height; y++ {
		x := 0
		for x < bytesPerLine && rleIdx < len(rleData) {
			b := rleData[rleIdx]
			rleIdx++
			var count int
			var val byte
			if (b & 0xC0) == 0xC0 {
				count = int(b & 0x3F)
				if rleIdx < len(rleData) {
					val = rleData[rleIdx]
					rleIdx++
				}
			} else {
				count = 1
				val = b
			}

			for k := 0; k < count; k++ {
				if x < width {
					c := palette[val]
					img.Set(x, y, c)
				}
				x++
			}
		}
	}

	return img, palette, nil
}

// generatePlaceholderAvatar creates a sleek 128x128 placeholder graphic with the asset initials.
func generatePlaceholderAvatar(name string) image.Image {
	const size = 120
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Dark gradient background
	baseName := filepath.Base(name)
	hash := 0
	for _, ch := range baseName {
		hash = (hash*31 + int(ch)) % 360
	}

	// Compute a vibrant hue-based accent color
	accentR, accentG, accentB := hsvToRGB(float64(hash), 0.7, 0.85)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Subtle diagonal gradient
			f := float64(x+y) / float64(size*2)
			r := uint8(float64(20)*(1-f) + float64(accentR)*f*0.3)
			g := uint8(float64(24)*(1-f) + float64(accentG)*f*0.3)
			b := uint8(float64(36)*(1-f) + float64(accentB)*f*0.3)
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	// Inner rounded border frame
	borderCol := color.RGBA{R: accentR, G: accentG, B: accentB, A: 180}
	for i := 4; i < size-4; i++ {
		img.SetRGBA(i, 4, borderCol)
		img.SetRGBA(i, size-5, borderCol)
		img.SetRGBA(4, i, borderCol)
		img.SetRGBA(size-5, i, borderCol)
	}

	// Draw center monogram badge
	badgeRect := image.Rect(size/2-24, size/2-24, size/2+24, size/2+24)
	draw.Draw(img, badgeRect, &image.Uniform{color.RGBA{R: accentR, G: accentG, B: accentB, A: 220}}, image.Point{}, draw.Over)

	return img
}

func hsvToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	x := c * (1 - float64(int(h/60)%2-1))
	if x < 0 {
		x = -x
	}
	m := v - c
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return uint8((r + m) * 255), uint8((g + m) * 255), uint8((b + m) * 255)
}
