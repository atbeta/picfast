//go:build novips

package service

import (
	"bytes"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/rwcarlsen/goexif/exif"
	xdraw "golang.org/x/image/draw"
)

// defaultJPEGQuality mirrors govips NewJpegExportParams (quality 80/100).
const defaultJPEGQuality = 80

// maxDecodePixels bounds decoded image dimensions. Unlike libvips, which
// shrinks on load, stdlib decoding allocates the full bitmap — a small crafted
// file with huge dimensions (decompression bomb) could OOM a constrained
// instance. 64MP ≈ 256MB of NRGBA worst case. A var so tests can lower it.
var maxDecodePixels int64 = 64_000_000

// exifOrientation reads the EXIF Orientation tag (1-8). Non-JPEG input or
// missing/invalid tags return 1 (normal).
func exifOrientation(data []byte) int {
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		return 1
	}
	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}
	o, err := tag.Int(0)
	if err != nil || o < 1 || o > 8 {
		return 1
	}
	return o
}

// toNRGBA converts any decoded image to NRGBA for pixel-level manipulation.
func toNRGBA(src image.Image) *image.NRGBA {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	xdraw.Draw(dst, b, src, b.Min, xdraw.Src)
	return dst
}

// orientNRGBA returns the image with EXIF orientation o applied so pixels read
// top-left first, matching libvips AutoRotate.
func orientNRGBA(src *image.NRGBA, o int) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	var dst *image.NRGBA
	if o >= 5 && o <= 8 {
		dst = image.NewNRGBA(image.Rect(0, 0, h, w))
	} else {
		dst = image.NewNRGBA(image.Rect(0, 0, w, h))
	}
	for y := 0; y < dst.Bounds().Dy(); y++ {
		for x := 0; x < dst.Bounds().Dx(); x++ {
			var sx, sy int
			switch o {
			case 2: // flip horizontal
				sx, sy = w-1-x, y
			case 3: // rotate 180
				sx, sy = w-1-x, h-1-y
			case 4: // flip vertical
				sx, sy = x, h-1-y
			case 5: // transpose
				sx, sy = y, x
			case 6: // rotate 90 CW
				sx, sy = y, h-1-x
			case 7: // transverse
				sx, sy = w-1-y, h-1-x
			case 8: // rotate 90 CCW
				sx, sy = w-1-y, x
			default:
				sx, sy = x, y
			}
			dst.SetNRGBA(x, y, src.NRGBAAt(sx+b.Min.X, sy+b.Min.Y))
		}
	}
	return dst
}

// decodeAutoOriented decodes image data and applies EXIF orientation,
// matching the vips backend's AutoRotate import behavior. Inputs whose header
// reports more than maxDecodePixels are rejected before any pixel allocation.
func decodeAutoOriented(data []byte) (image.Image, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if int64(cfg.Width)*int64(cfg.Height) > maxDecodePixels {
		return nil, ErrImageTooLarge
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if o := exifOrientation(data); o > 1 {
		return orientNRGBA(toNRGBA(img), o), nil
	}
	return img, nil
}

// fitWithin computes the target size fitting src inside maxW x maxH while
// preserving aspect ratio; it never upscales. Zero/negative maxes are ignored.
func fitWithin(srcW, srcH, maxW, maxH int) (int, int) {
	scale := 1.0
	if maxW > 0 && srcW > maxW {
		if s := float64(maxW) / float64(srcW); s < scale {
			scale = s
		}
	}
	if maxH > 0 && srcH > maxH {
		if s := float64(maxH) / float64(srcH); s < scale {
			scale = s
		}
	}
	if scale >= 1.0 {
		return srcW, srcH
	}
	w := int(float64(srcW)*scale + 0.5)
	h := int(float64(srcH)*scale + 0.5)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}

// resizeFit scales src down to fit within maxW x maxH keeping aspect ratio.
// Returns src unchanged when it already fits.
func resizeFit(src image.Image, maxW, maxH int) image.Image {
	b := src.Bounds()
	w, h := fitWithin(b.Dx(), b.Dy(), maxW, maxH)
	if w == b.Dx() && h == b.Dy() {
		return src
	}
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, b, xdraw.Over, nil)
	return dst
}

func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	if quality <= 0 || quality >= 100 {
		quality = defaultJPEGQuality
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeGIF(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := gif.Encode(&buf, img, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodePureGo exports img in the given normalized format, mirroring the vips
// exporter registry: unknown formats fall back to JPEG, WebP is unsupported.
func encodePureGo(img image.Image, format string, quality int) ([]byte, error) {
	switch normalizeExportFormat(format) {
	case "jpeg":
		return encodeJPEG(img, quality)
	case "png":
		return encodePNG(img)
	case "gif":
		return encodeGIF(img)
	case "webp":
		return nil, ErrFormatUnsupported
	default:
		return encodeJPEG(img, quality)
	}
}

// ReadImageDimensions loads an image and returns its width/height post-EXIF
// orientation. Undecodable input returns an error.
func ReadImageDimensions(data []byte) (int, int, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0, err
	}
	w, h := cfg.Width, cfg.Height
	if o := exifOrientation(data); o >= 5 && o <= 8 {
		w, h = h, w
	}
	return w, h, nil
}

func ProcessImage(data []byte, saveFormat string, quality int, stripExif bool) (*ProcessedImage, error) {
	img, err := decodeAutoOriented(data)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	width, height := b.Dx(), b.Dy()

	// Same no-op contract as the vips backend: nothing requested means the
	// original bytes pass through untouched (the decode above still validates
	// the input and reports dimensions).
	if !stripExif && saveFormat == "" && (quality <= 0 || quality >= 100) {
		return &ProcessedImage{Data: data, Width: width, Height: height}, nil
	}

	out, err := encodePureGo(img, saveFormat, quality)
	if err != nil {
		return nil, err
	}
	return &ProcessedImage{Data: out, Width: width, Height: height}, nil
}

// ResizeImageToWidth resizes an image to the target width while keeping aspect
// ratio. Unsupported or no-op cases return the original bytes.
func ResizeImageToWidth(data []byte, formatHint string, width int) ([]byte, error) {
	if width <= 0 {
		return data, nil
	}

	format := normalizeExportFormat(formatHint)
	if format == "gif" {
		return data, nil // keep animated/static gif behavior unchanged for now
	}

	img, err := decodeAutoOriented(data)
	if err != nil {
		return nil, err
	}

	origW := img.Bounds().Dx()
	if origW <= 0 || width >= origW {
		return data, nil
	}

	resized := resizeFit(img, width, 0)
	return encodePureGo(resized, format, 0)
}

// ProcessImageOnTheFly applies resize, quality, and format conversion.
// GIF images are skipped (returned as-is). On any processing error the
// original data is returned unchanged so callers can always fall back to the
// source image.
//
// The returned ok flag reports whether any processing was actually applied.
// Callers MUST NOT assume the output format matches the requested format when
// ok is false.
func ProcessImageOnTheFly(data []byte, sourceFormat string, p ProcessingParams) ([]byte, bool) {
	if p.IsEmpty() {
		return data, false
	}

	format := normalizeExportFormat(sourceFormat)
	if format == "gif" {
		return data, false
	}

	needResize := p.Width > 0 || p.Height > 0
	needQuality := p.Quality > 0 && p.Quality < 100
	needFormat := p.Format != "" && normalizeExportFormat(p.Format) != format

	if !needResize && !needQuality && !needFormat {
		return data, false
	}

	exportFormat := format
	if needFormat {
		exportFormat = normalizeExportFormat(p.Format)
	}
	if exportFormat == "webp" {
		return data, false // WebP output unsupported by this backend
	}

	img, err := decodeAutoOriented(data)
	if err != nil {
		return data, false
	}

	if needResize {
		img = resizeFit(img, p.Width, p.Height)
	}

	quality := 100
	if needQuality {
		quality = p.Quality
	}

	out, err := encodePureGo(img, exportFormat, quality)
	if err != nil {
		return data, false
	}
	return out, true
}
