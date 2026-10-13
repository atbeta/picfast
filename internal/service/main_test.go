package service

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	StartupImaging()
	code := m.Run()
	ShutdownImaging()
	os.Exit(code)
}

// createTestImage encodes a solid-color test image in the given format
// ("png" or "jpeg"). Shared by tests that run under both image backends.
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
