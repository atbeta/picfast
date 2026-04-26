package service

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func createTestPNG(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{100, 150, 200, 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestApplyWatermark(t *testing.T) {
	data := createTestPNG(400, 300)

	cfg := WatermarkConfig{
		Text:     "TestWatermark",
		Position: "bottom-right",
		FontSize: 16,
		Color:    "#FFFFFF",
		Opacity:  0.8,
	}

	result, err := ApplyWatermark(data, cfg, "png", 90)
	if err != nil {
		t.Fatalf("apply watermark failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("watermark result is empty")
	}

	// Decode to verify valid image
	img, _, err := image.Decode(bytes.NewReader(result))
	if err != nil {
		t.Fatalf("decode watermarked image failed: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Fatalf("unexpected dimensions: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestApplyWatermarkJPEG(t *testing.T) {
	data := createTestPNG(200, 200)

	cfg := WatermarkConfig{
		Text:     "JPEG",
		Position: "center",
		FontSize: 12,
		Color:    "#FF0000",
		Opacity:  0.5,
	}

	result, err := ApplyWatermark(data, cfg, "jpeg", 85)
	if err != nil {
		t.Fatalf("apply watermark failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("watermark result is empty")
	}
}

func TestApplyWatermarkInvalidImage(t *testing.T) {
	cfg := WatermarkConfig{Text: "X"}
	_, err := ApplyWatermark([]byte("not-an-image"), cfg, "png", 90)
	if err == nil {
		t.Fatal("expected error for invalid image")
	}
}

func TestParseColor(t *testing.T) {
	cases := []struct {
		hex     string
		opacity float64
		wantR   uint8
		wantA   uint8
	}{
		{"#FFFFFF", 1.0, 255, 255},
		{"#FF0000", 0.5, 255, 127},
		{"#00FF00AA", 1.0, 0, 170},
	}

	for _, c := range cases {
		col, err := parseColor(c.hex, c.opacity)
		if err != nil {
			t.Fatalf("parseColor(%q, %f) error: %v", c.hex, c.opacity, err)
		}
		rgba, ok := col.(color.RGBA)
		if !ok {
			t.Fatal("expected color.RGBA")
		}
		if rgba.R != c.wantR {
			t.Fatalf("parseColor(%q) R=%d, want %d", c.hex, rgba.R, c.wantR)
		}
		if rgba.A != c.wantA {
			t.Fatalf("parseColor(%q) A=%d, want %d", c.hex, rgba.A, c.wantA)
		}
	}
}

func TestCalcWatermarkPosition(t *testing.T) {
	tests := []struct {
		pos       string
		wantX     int
		wantY     int
	}{
		{"top-left", 10, 26},
		{"top-right", 110, 26},
		{"bottom-left", 10, 190},
		{"bottom-right", 110, 190},
		{"center", 60, 108},
	}

	for _, tt := range tests {
		x, y := calcWatermarkPosition(200, 200, 80, 16, tt.pos)
		if x != tt.wantX || y != tt.wantY {
			t.Fatalf("position %s: got (%d,%d), want (%d,%d)", tt.pos, x, y, tt.wantX, tt.wantY)
		}
	}
}
