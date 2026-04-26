package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

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
	processed, err := ProcessImage(data, "jpeg", 80)
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
	processed, err := ProcessImage(data, "png", 100)
	if err != nil {
		t.Fatalf("process image failed: %v", err)
	}
	if len(processed.Data) == 0 {
		t.Fatal("processed data is empty")
	}
}

func TestProcessImageWebPKeepOriginal(t *testing.T) {
	// WebP is not supported by imaging, should return original
	data := createTestImage("jpeg", 100, 100)
	processed, err := ProcessImage(data, "webp", 90)
	if err != nil {
		t.Fatalf("process image failed: %v", err)
	}
	if len(processed.Data) == 0 {
		t.Fatal("processed data is empty")
	}
}

func TestProcessImageInvalidData(t *testing.T) {
	_, err := ProcessImage([]byte("not-an-image"), "jpeg", 80)
	if err == nil {
		t.Fatal("expected error for invalid image data")
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
