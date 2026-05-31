package service

import (
	"bytes"
	"fmt"
	"strings"
)

func validateUploadContent(ext string, fileData []byte) error {
	if len(fileData) == 0 {
		return fmt.Errorf("file is empty")
	}
	if !contentMatchesExtension(ext, fileData) {
		return fmt.Errorf("file content does not match extension .%s", ext)
	}
	return nil
}

func contentMatchesExtension(ext string, data []byte) bool {
	switch ext {
	case "jpg", "jpeg":
		return hasPrefix(data, []byte{0xFF, 0xD8, 0xFF})
	case "png":
		return hasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	case "gif":
		return hasPrefix(data, []byte("GIF87a")) || hasPrefix(data, []byte("GIF89a"))
	case "webp":
		return len(data) >= 12 && hasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP"))
	case "bmp":
		return hasPrefix(data, []byte("BM"))
	case "ico":
		return len(data) >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0x01 && data[3] == 0x00
	case "svg":
		return looksLikeSVG(data)
	case "tif", "tiff":
		return hasPrefix(data, []byte{0x49, 0x49, 0x2A, 0x00}) || hasPrefix(data, []byte{0x4D, 0x4D, 0x00, 0x2A})
	default:
		return true
	}
}

func hasPrefix(data, prefix []byte) bool {
	return len(data) >= len(prefix) && bytes.Equal(data[:len(prefix)], prefix)
}

func looksLikeSVG(data []byte) bool {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return false
	}
	if trimmed[0] == '<' {
		head := strings.ToLower(string(trimmed[:min(len(trimmed), 256)]))
		return strings.Contains(head, "<svg")
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
