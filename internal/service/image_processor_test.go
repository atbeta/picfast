package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
)

func TestMain(m *testing.M) {
	vips.Startup(nil)
	code := m.Run()
	vips.Shutdown()
	os.Exit(code)
}

func createTestImage(format string, width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{50, 100, 150, 255})
		}
	}
	var buf bytes.Buffer
	if format == "png" {
		png.Encode(&buf, img)
	} else {
		jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	}
	return buf.Bytes()
}

func TestProcessImageJPEG(t *testing.T) {
	data := createTestImage("jpeg", 200, 200)
	processed, err := ProcessImage(data, "jpeg", 80, false)
	if err != nil {
		t.Fatalf("process image failed: %v", err)
	}
	if len(processed.Data) == 0 {
		t.Fatal("processed data is empty")
	}
	if processed.Width != 200 || processed.Height != 200 {
		t.Fatalf("unexpected dimensions: %dx%d", processed.Width, processed.Height)
	}
}

func TestProcessImagePNG(t *testing.T) {
	data := createTestImage("png", 100, 100)
	processed, err := ProcessImage(data, "png", 100, false)
	if err != nil {
		t.Fatalf("process image failed: %v", err)
	}
	if len(processed.Data) == 0 {
		t.Fatal("processed data is empty")
	}
}

func TestProcessImageWebP(t *testing.T) {
	// govips supports webp, should encode successfully
	data := createTestImage("jpeg", 100, 100)
	processed, err := ProcessImage(data, "webp", 90, false)
	if err != nil {
		t.Fatalf("process image failed: %v", err)
	}
	if len(processed.Data) == 0 {
		t.Fatal("processed data is empty")
	}
}

func TestProcessImageStripExif(t *testing.T) {
	data := createTestImage("jpeg", 100, 100)
	processed, err := ProcessImage(data, "", 0, true)
	if err != nil {
		t.Fatalf("process image failed: %v", err)
	}
	if len(processed.Data) == 0 {
		t.Fatal("processed data is empty")
	}
}

func TestProcessImageInvalidData(t *testing.T) {
	_, err := ProcessImage([]byte("not-an-image"), "jpeg", 80, false)
	if err == nil {
		t.Fatal("expected error for invalid image data")
	}
}

func TestResizeImageToWidth(t *testing.T) {
	data := createTestImage("jpeg", 400, 200)
	out, err := ResizeImageToWidth(data, "jpeg", 200)
	if err != nil {
		t.Fatalf("resize image failed: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("resized image is empty")
	}

	w, h, err := DecodeImageDimensions(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode resized dimensions failed: %v", err)
	}
	if w != 200 || h != 100 {
		t.Fatalf("unexpected resized dimensions: %dx%d", w, h)
	}
}

func TestDecodeImageDimensions(t *testing.T) {
	data := createTestImage("png", 300, 200)
	w, h, err := DecodeImageDimensions(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode dimensions failed: %v", err)
	}
	if w != 300 || h != 200 {
		t.Fatalf("unexpected dimensions: %dx%d", w, h)
	}
}

func TestProcessImageOnTheFlyEmpty(t *testing.T) {
	data := createTestImage("jpeg", 400, 200)
	out := ProcessImageOnTheFly(data, "jpeg", ProcessingParams{})
	if !bytes.Equal(out, data) {
		t.Fatal("empty params should return original data unchanged")
	}
}

func TestProcessImageOnTheFlyWidthOnly(t *testing.T) {
	data := createTestImage("jpeg", 400, 200)
	out := ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Width: 200})
	if len(out) == 0 {
		t.Fatal("output is empty")
	}
	w, h, err := DecodeImageDimensions(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode dimensions: %v", err)
	}
	if w != 200 || h != 100 {
		t.Fatalf("unexpected dimensions: %dx%d", w, h)
	}
}

func TestProcessImageOnTheFlyHeightOnly(t *testing.T) {
	data := createTestImage("jpeg", 400, 200)
	out := ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Height: 100})
	if len(out) == 0 {
		t.Fatal("output is empty")
	}
	w, h, err := DecodeImageDimensions(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode dimensions: %v", err)
	}
	if w != 200 || h != 100 {
		t.Fatalf("unexpected dimensions: %dx%d", w, h)
	}
}

func TestProcessImageOnTheFlyWidthAndHeight(t *testing.T) {
	data := createTestImage("jpeg", 400, 200)
	// 400x200 → fit in 100x100 → scale=min(100/400, 100/200)=0.25 → 100x50
	out := ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Width: 100, Height: 100})
	if len(out) == 0 {
		t.Fatal("output is empty")
	}
	w, h, err := DecodeImageDimensions(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode dimensions: %v", err)
	}
	if w != 100 || h != 50 {
		t.Fatalf("unexpected dimensions: %dx%d", w, h)
	}
}

func TestProcessImageOnTheFlyQuality(t *testing.T) {
	data := createTestImage("jpeg", 200, 200)
	out := ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Quality: 50})
	if len(out) == 0 {
		t.Fatal("output is empty")
	}
	// Quality compression should produce smaller output
	if len(out) >= len(data) {
		t.Fatalf("expected smaller output: got %d, original %d", len(out), len(data))
	}
}

func TestProcessImageOnTheFlyFormat(t *testing.T) {
	data := createTestImage("jpeg", 200, 200)
	out := ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Format: "webp"})
	if len(out) == 0 {
		t.Fatal("output is empty")
	}
	// Output should be valid webp (starts with RIFF)
	if len(out) < 4 || string(out[:4]) != "RIFF" {
		t.Fatal("output does not look like webp")
	}
}

func TestProcessImageOnTheFlyGifSkipped(t *testing.T) {
	data := createTestImage("jpeg", 200, 200)
	out := ProcessImageOnTheFly(data, "gif", ProcessingParams{Width: 100})
	if !bytes.Equal(out, data) {
		t.Fatal("gif should be returned unchanged")
	}
}

func TestMimeTypeForFormat(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{"jpeg", "image/jpeg"},
		{"jpg", "image/jpeg"},
		{"png", "image/png"},
		{"webp", "image/webp"},
		{"gif", "image/gif"},
		{"unknown", ""},
		{"", "image/jpeg"},
	}
	for _, tt := range tests {
		if got := MimeTypeForFormat(tt.format); got != tt.want {
			t.Errorf("MimeTypeForFormat(%q) = %q, want %q", tt.format, got, tt.want)
		}
	}
}
