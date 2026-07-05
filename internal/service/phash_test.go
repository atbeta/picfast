package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func fillImage(img *image.RGBA, c color.RGBA) {
	for y := range img.Bounds().Dy() {
		for x := range img.Bounds().Dx() {
			img.Set(x, y, c)
		}
	}
}

func generateTestPNG(t *testing.T, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	fillImage(img, c)
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	return buf.Bytes()
}

func generateTestJPEG(t *testing.T, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	fillImage(img, c)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatalf("failed to encode JPEG: %v", err)
	}
	return buf.Bytes()
}

func TestComputePHash(t *testing.T) {
	gray := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	pngData := generateTestPNG(t, gray)

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

func TestComputePHashFormats(t *testing.T) {
	gray := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	pngData := generateTestPNG(t, gray)
	jpgData := generateTestJPEG(t, gray)

	pHash, err := ComputePHash(pngData)
	if err != nil {
		t.Fatalf("PNG phash failed: %v", err)
	}
	jHash, err := ComputePHash(jpgData)
	if err != nil {
		t.Fatalf("JPEG phash failed: %v", err)
	}

	// Same image content in different formats should produce similar hashes
	xor := pHash ^ jHash
	diff := bitCount(xor)
	if diff > 30 {
		t.Errorf("PNG/JPEG phash distance too large: %d bits, hash=%016x vs %016x", diff, pHash, jHash)
	}
}

func TestComputePHashDifferentImages(t *testing.T) {
	red := generateTestPNG(t, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	green := generateTestPNG(t, color.RGBA{R: 0, G: 255, B: 0, A: 255})

	rHash, _ := ComputePHash(red)
	gHash, _ := ComputePHash(green)

	xor := rHash ^ gHash
	diff := bitCount(xor)
	if diff < 5 {
		t.Errorf("different images should have different hashes, got distance %d", diff)
	}
}

func TestComputePHashInvalidData(t *testing.T) {
	_, err := ComputePHash([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid data")
	}
}

func bitCount(x uint64) int {
	c := 0
	for x != 0 {
		x &= x - 1
		c++
	}
	return c
}
