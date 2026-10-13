//go:build novips

package service

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/rwcarlsen/goexif/exif"
	"golang.org/x/image/bmp"
)

// tinyWebP is an 8x6 WebP generated with libvips for decode coverage (the
// pure-Go backend decodes but never encodes WebP).
const tinyWebP = "UklGRiIBAABXRUJQVlA4WAoAAAAIAAAABwAABQAAVlA4IEIAAADwAQCdASoIAAYAAgA0JbACdLoAArl9WmgA/vXyaL1/78rJYJLL2P/HYlbfnF53WRvK8//y5RXDwdeHOH53/6cIAABFWElGugAAAEV4aWYAAElJKgAIAAAABgASAQMAAQAAAAEAAAAaAQUAAQAAAFYAAAAbAQUAAQAAAF4AAAAoAQMAAQAAAAIAAAATAgMAAQAAAAEAAABphwQAAQAAAGYAAAAAAAAALxkBAOgDAAAvGQEA6AMAAAYAAJAHAAQAAAAwMjEwAZEHAAQAAAABAgMAAKAHAAQAAAAwMTAwAaADAAEAAAD//wAAAqAEAAEAAAAIAAAAA6AEAAEAAAAGAAAAAAAAAA=="

// asymmetricImage returns a 3x2 NRGBA whose pixels are all distinct so
// orientation transforms can be verified pixel-by-pixel.
func asymmetricImage() *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	colors := [2][3]color.NRGBA{
		{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}},
		{{255, 255, 0, 255}, {0, 255, 255, 255}, {255, 0, 255, 255}},
	}
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			img.SetNRGBA(x, y, colors[y][x])
		}
	}
	return img
}

// jpegWithOrientation encodes img as JPEG and injects an APP1 EXIF segment
// carrying the given Orientation tag (1-8).
func jpegWithOrientation(t *testing.T, img image.Image, orientation int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("jpeg encode: %v", err)
	}
	src := buf.Bytes()

	// Minimal TIFF directory with a single SHORT tag 0x0112 (Orientation).
	tiff := []byte{
		'I', 'I', 0x2A, 0x00, // little-endian, magic 42
		0x08, 0x00, 0x00, 0x00, // IFD offset
		0x01, 0x00, // 1 entry
		0x12, 0x01, 0x03, 0x00, // tag 0x0112, type SHORT
		0x01, 0x00, 0x00, 0x00, // count 1
		byte(orientation), 0x00, 0x00, 0x00, // value
		0x00, 0x00, 0x00, 0x00, // next IFD
	}
	seg := append([]byte("Exif\x00\x00"), tiff...)
	segLen := len(seg) + 2

	out := make([]byte, 0, len(src)+segLen+4)
	out = append(out, src[:2]...) // SOI
	out = append(out, 0xFF, 0xE1, byte(segLen>>8), byte(segLen&0xFF))
	out = append(out, seg...)
	out = append(out, src[2:]...)
	return out
}

func TestOrientNRGBAAllOrientations(t *testing.T) {
	src := asymmetricImage()
	c := func(x, y int) color.NRGBA { return src.NRGBAAt(x, y) }
	a, b, cc := c(0, 0), c(1, 0), c(2, 0)
	d, e, f := c(0, 1), c(1, 1), c(2, 1)

	cases := []struct {
		orient int
		w, h   int
		grid   [][]color.NRGBA // expected [row][col]
	}{
		{1, 3, 2, [][]color.NRGBA{{a, b, cc}, {d, e, f}}},
		{2, 3, 2, [][]color.NRGBA{{cc, b, a}, {f, e, d}}},
		{3, 3, 2, [][]color.NRGBA{{f, e, d}, {cc, b, a}}},
		{4, 3, 2, [][]color.NRGBA{{d, e, f}, {a, b, cc}}},
		{5, 2, 3, [][]color.NRGBA{{a, d}, {b, e}, {cc, f}}},
		{6, 2, 3, [][]color.NRGBA{{d, a}, {e, b}, {f, cc}}},
		{7, 2, 3, [][]color.NRGBA{{f, cc}, {e, b}, {d, a}}},
		{8, 2, 3, [][]color.NRGBA{{cc, f}, {b, e}, {a, d}}},
	}
	for _, tc := range cases {
		got := orientNRGBA(asymmetricImage(), tc.orient)
		if got.Bounds().Dx() != tc.w || got.Bounds().Dy() != tc.h {
			t.Fatalf("orientation %d: size %dx%d, want %dx%d", tc.orient, got.Bounds().Dx(), got.Bounds().Dy(), tc.w, tc.h)
		}
		for y := 0; y < tc.h; y++ {
			for x := 0; x < tc.w; x++ {
				if got.NRGBAAt(x, y) != tc.grid[y][x] {
					t.Fatalf("orientation %d pixel (%d,%d): got %v want %v", tc.orient, x, y, got.NRGBAAt(x, y), tc.grid[y][x])
				}
			}
		}
	}
}

func TestDecodeAutoOrientedAppliesExif(t *testing.T) {
	data := jpegWithOrientation(t, asymmetricImage(), 6)
	img, err := decodeAutoOriented(data)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// 3x2 rotated 90 CW becomes 2x3.
	if img.Bounds().Dx() != 2 || img.Bounds().Dy() != 3 {
		t.Fatalf("expected 2x3 after orientation 6, got %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestReadImageDimensionsSwapsOnOrientation(t *testing.T) {
	data := jpegWithOrientation(t, asymmetricImage(), 6)
	w, h, err := ReadImageDimensions(data)
	if err != nil {
		t.Fatalf("ReadImageDimensions: %v", err)
	}
	if w != 2 || h != 3 {
		t.Fatalf("expected swapped 2x3, got %dx%d", w, h)
	}

	plain := createTestImage("jpeg", 40, 20)
	w, h, err = ReadImageDimensions(plain)
	if err != nil || w != 40 || h != 20 {
		t.Fatalf("expected 40x20, got %dx%d err=%v", w, h, err)
	}
}

func TestPureGoDecodeFormats(t *testing.T) {
	src := asymmetricImage()

	var pngBuf, gifBuf, bmpBuf bytes.Buffer
	png.Encode(&pngBuf, src)
	gif.Encode(&gifBuf, src, nil)
	bmp.Encode(&bmpBuf, src)

	webpData, err := base64.StdEncoding.DecodeString(tinyWebP)
	if err != nil {
		t.Fatalf("webp fixture: %v", err)
	}

	cases := []struct {
		name string
		data []byte
		w, h int
	}{
		{"jpeg", createTestImage("jpeg", 10, 7), 10, 7},
		{"png", pngBuf.Bytes(), 3, 2},
		{"gif", gifBuf.Bytes(), 3, 2},
		{"bmp", bmpBuf.Bytes(), 3, 2},
		{"webp", webpData, 8, 6},
	}
	for _, tc := range cases {
		w, h, err := ReadImageDimensions(tc.data)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if w != tc.w || h != tc.h {
			t.Fatalf("%s: got %dx%d want %dx%d", tc.name, w, h, tc.w, tc.h)
		}
	}
}

func TestProcessImagePureGo(t *testing.T) {
	data := createTestImage("jpeg", 100, 80)

	res, err := ProcessImage(data, "jpeg", 50, false)
	if err != nil {
		t.Fatalf("jpeg: %v", err)
	}
	if res.Width != 100 || res.Height != 80 {
		t.Fatalf("dims %dx%d", res.Width, res.Height)
	}
	if _, _, err := image.Decode(bytes.NewReader(res.Data)); err != nil {
		t.Fatalf("output not decodable: %v", err)
	}

	res, err = ProcessImage(data, "png", 0, false)
	if err != nil {
		t.Fatalf("png: %v", err)
	}
	if _, format, err := image.DecodeConfig(bytes.NewReader(res.Data)); err != nil || format != "png" {
		t.Fatalf("expected png output, got %q err=%v", format, err)
	}

	res, err = ProcessImage(data, "gif", 0, false)
	if err != nil {
		t.Fatalf("gif: %v", err)
	}
	if _, format, err := image.DecodeConfig(bytes.NewReader(res.Data)); err != nil || format != "gif" {
		t.Fatalf("expected gif output, got %q err=%v", format, err)
	}

	// Unknown format falls back to JPEG like the vips registry does.
	res, err = ProcessImage(data, "bogus", 0, false)
	if err != nil {
		t.Fatalf("unknown format: %v", err)
	}
	if _, format, err := image.DecodeConfig(bytes.NewReader(res.Data)); err != nil || format != "jpeg" {
		t.Fatalf("expected jpeg fallback, got %q err=%v", format, err)
	}
}

func TestProcessImageWebPUnsupported(t *testing.T) {
	data := createTestImage("jpeg", 50, 50)
	_, err := ProcessImage(data, "webp", 80, false)
	if err != ErrFormatUnsupported {
		t.Fatalf("expected ErrFormatUnsupported, got %v", err)
	}
}

func TestProcessImageNoOpEarlyReturn(t *testing.T) {
	data := createTestImage("jpeg", 30, 30)
	res, err := ProcessImage(data, "", 0, false)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !bytes.Equal(res.Data, data) {
		t.Fatal("expected original bytes on no-op")
	}
}

func TestProcessImageStripExif(t *testing.T) {
	data := jpegWithOrientation(t, asymmetricImage(), 6)
	res, err := ProcessImage(data, "jpeg", 90, true)
	if err != nil {
		t.Fatalf("ProcessImage: %v", err)
	}
	if _, err := exif.Decode(bytes.NewReader(res.Data)); err == nil {
		t.Fatal("expected no EXIF in stripped output")
	}
	// Orientation was applied before encode: 3x2 -> 2x3.
	if res.Width != 2 || res.Height != 3 {
		t.Fatalf("expected 2x3, got %dx%d", res.Width, res.Height)
	}
}

func TestProcessImageOnTheFlyPureGo(t *testing.T) {
	data := createTestImage("jpeg", 400, 200)

	out, ok := ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Width: 200})
	if !ok {
		t.Fatal("expected processed")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if format != "jpeg" || cfg.Width != 200 || cfg.Height != 100 {
		t.Fatalf("got %s %dx%d", format, cfg.Width, cfg.Height)
	}

	// Never upscale.
	out, ok = ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Width: 800})
	if ok {
		cfg, _, _ = image.DecodeConfig(bytes.NewReader(out))
		if cfg.Width > 400 {
			t.Fatalf("upscaled to %d", cfg.Width)
		}
	}

	// f_webp returns original unprocessed under novips.
	out, ok = ProcessImageOnTheFly(data, "jpeg", ProcessingParams{Format: "webp"})
	if ok || !bytes.Equal(out, data) {
		t.Fatal("expected (data,false) for f_webp")
	}

	// GIF source skipped.
	var gifBuf bytes.Buffer
	gif.Encode(&gifBuf, asymmetricImage(), nil)
	gifData := gifBuf.Bytes()
	out, ok = ProcessImageOnTheFly(gifData, "gif", ProcessingParams{Width: 50})
	if ok || !bytes.Equal(out, gifData) {
		t.Fatal("expected gif passthrough")
	}
}

func TestRenderThumbnailPureGo(t *testing.T) {
	data := createTestImage("jpeg", 800, 600)
	out, err := RenderThumbnail(data, "jpg")
	if err != nil {
		t.Fatalf("RenderThumbnail: %v", err)
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("thumbnail decode: %v", err)
	}
	if format != "png" {
		t.Fatalf("expected png, got %s", format)
	}
	if cfg.Width > ThumbnailMaxSize || cfg.Height > ThumbnailMaxSize {
		t.Fatalf("thumbnail %dx%d exceeds max", cfg.Width, cfg.Height)
	}
	if cfg.Width != 400 || cfg.Height != 300 {
		t.Fatalf("expected 400x300, got %dx%d", cfg.Width, cfg.Height)
	}

	// Never upscale small images.
	small := createTestImage("jpeg", 100, 50)
	out, err = RenderThumbnail(small, "jpg")
	if err != nil {
		t.Fatalf("RenderThumbnail small: %v", err)
	}
	cfg, _, err = image.DecodeConfig(bytes.NewReader(out))
	if err != nil || cfg.Width != 100 || cfg.Height != 50 {
		t.Fatalf("expected 100x50, got %dx%d err=%v", cfg.Width, cfg.Height, err)
	}

	// Skipped types.
	if _, err := RenderThumbnail(data, "svg"); err != ErrThumbnailSkipped {
		t.Fatalf("expected ErrThumbnailSkipped for svg, got %v", err)
	}
	if _, err := RenderThumbnail(data, "ico"); err != ErrThumbnailSkipped {
		t.Fatalf("expected ErrThumbnailSkipped for ico, got %v", err)
	}

	// Decode failure propagates.
	if _, err := RenderThumbnail([]byte("garbage"), "jpg"); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestApplyWatermarkUnsupported(t *testing.T) {
	_, err := ApplyWatermark(createTestImage("png", 50, 50), WatermarkConfig{Text: "x"}, "png", 80)
	if err != ErrWatermarkUnsupported {
		t.Fatalf("expected ErrWatermarkUnsupported, got %v", err)
	}
}

func TestCapabilitiesPureGo(t *testing.T) {
	c := Capabilities()
	if c.Backend != "go" || c.WebPOutput || c.Watermark {
		t.Fatalf("unexpected capabilities: %+v", c)
	}
}

func TestProcessUploadImageWebPFallbackAppliesQuality(t *testing.T) {
	// A jpeg upload with format=webp configured: under novips the format is
	// kept as jpeg but quality processing still runs.
	data := createTestImage("jpeg", 100, 100)
	cfg := imageProcessingConfig{
		Quality:   20,
		Format:    "webp",
		StripExif: false,
	}
	result := processUploadImage(data, "jpg", cfg)
	if !result.processed {
		t.Fatal("expected processing to run")
	}
	_, format, err := image.DecodeConfig(bytes.NewReader(result.data))
	if err != nil || format != "jpeg" {
		t.Fatalf("expected jpeg output, got %q err=%v", format, err)
	}
}
