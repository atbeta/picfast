package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func generateValidPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	return buf.Bytes()
}

func TestComputePHash(t *testing.T) {
	pngData := generateValidPNG(t)

	h1, err := ComputePHash(pngData)
	if err != nil {
		t.Fatalf("ComputePHash failed: %v", err)
	}
	if h1 == 0 {
		t.Error("expected non-zero phash")
	}

	h2, err := ComputePHash(pngData)
	if err != nil {
		t.Fatalf("ComputePHash failed: %v", err)
	}
	if h1 != h2 {
		t.Errorf("phash not deterministic: %d != %d", h1, h2)
	}
}

func TestComputePHashInvalidData(t *testing.T) {
	_, err := ComputePHash([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid data")
	}
}
