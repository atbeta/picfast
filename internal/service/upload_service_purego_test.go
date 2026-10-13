//go:build novips

package service

import (
	"bytes"
	"encoding/base64"
	"testing"
)

// A WebP source upload with no format override must not attempt re-encoding:
// the effective target is webp, which this backend cannot encode.
func TestProcessUploadImageWebPSourceSkipsProcessing(t *testing.T) {
	data, err := base64.StdEncoding.DecodeString(tinyWebP)
	if err != nil {
		t.Fatalf("webp fixture: %v", err)
	}
	cfg := imageProcessingConfig{Quality: 85, StripExif: true}
	result := processUploadImage(data, "webp", cfg)
	if result.processed {
		t.Fatal("expected unprocessed result")
	}
	if !bytes.Equal(result.data, data) {
		t.Fatal("expected original bytes returned unchanged")
	}
	if result.width != 8 || result.height != 6 {
		t.Fatalf("expected dims 8x6 from header, got %dx%d", result.width, result.height)
	}
}

// minimalPNG parses headers but fails full decode under the pure-Go backend
// (bad zlib checksum), so ProcessImage errors — dimensions must still be
// filled from the header afterwards.
func TestProcessUploadImageFillsDimsOnProcessFailure(t *testing.T) {
	cfg := imageProcessingConfig{Quality: 50}
	result := processUploadImage(minimalPNG, "png", cfg)
	if result.processed {
		t.Fatal("expected unprocessed result")
	}
	if result.width != 1 || result.height != 1 {
		t.Fatalf("expected dims 1x1 from header, got %dx%d", result.width, result.height)
	}
}
