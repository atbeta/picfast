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
