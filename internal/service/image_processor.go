package service

import (
	"image"
	"io"

	"github.com/davidbyttow/govips/v2/vips"
)

type ProcessedImage struct {
	Data   []byte
	Width  int
	Height int
}

// FormatExporter encodes a vips image to a specific format.
type FormatExporter interface {
	Name() string
	Export(img *vips.ImageRef, quality int, stripMetadata bool) ([]byte, error)
}

var exportRegistry = map[string]FormatExporter{}

// RegisterExporter registers an image format exporter.
func RegisterExporter(e FormatExporter) {
	exportRegistry[e.Name()] = e
}

// ReadImageDimensions loads an image and returns its width/height.
// This is a header-only probe (no pixel re-encoding).
func ReadImageDimensions(data []byte) (int, int, error) {
	params := vips.NewImportParams()
	params.AutoRotate.Set(true)
	params.FailOnError.Set(true)

	img, err := vips.LoadImageFromBuffer(data, params)
	if err != nil {
		return 0, 0, err
	}
	defer img.Close()

	return img.Width(), img.Height(), nil
}

func ProcessImage(data []byte, saveFormat string, quality int, stripExif bool) (*ProcessedImage, error) {
	params := vips.NewImportParams()
	params.AutoRotate.Set(true)
	params.FailOnError.Set(true)

	img, err := vips.LoadImageFromBuffer(data, params)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	width := img.Width()
	height := img.Height()

	// stripExif implies re-encoding because loading to vips.ImageRef drops EXIF metadata.
	// Avoid unnecessary re-encoding only when no changes are requested.
	if !stripExif && saveFormat == "" && (quality <= 0 || quality >= 100) {
		return &ProcessedImage{Data: data, Width: width, Height: height}, nil
	}

	format := normalizeExportFormat(saveFormat)

	exporter, ok := exportRegistry[format]
	if !ok {
		exporter = exportRegistry["jpeg"]
	}

	out, err := exporter.Export(img, quality, stripExif)
	if err != nil {
		return nil, err
	}

	return &ProcessedImage{Data: out, Width: width, Height: height}, nil
}

// ResizeImageToWidth resizes an image to the target width while keeping aspect ratio.
// Unsupported or no-op cases return the original bytes.
func ResizeImageToWidth(data []byte, formatHint string, width int) ([]byte, error) {
	if width <= 0 {
		return data, nil
	}

	format := normalizeExportFormat(formatHint)
	if format == "gif" {
		return data, nil // keep animated/static gif behavior unchanged for now
	}

	params := vips.NewImportParams()
	params.AutoRotate.Set(true)
	params.FailOnError.Set(true)

	img, err := vips.LoadImageFromBuffer(data, params)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	origW := img.Width()
	if origW <= 0 || width >= origW {
		return data, nil
	}

	scale := float64(width) / float64(origW)
	if err := img.Resize(scale, vips.KernelAuto); err != nil {
		return nil, err
	}

	exporter, ok := exportRegistry[format]
	if !ok {
		exporter = exportRegistry["jpeg"]
	}
	return exporter.Export(img, 100, false)
}

// ProcessingParams holds on-the-fly image processing parameters parsed from
// the URL variant suffix (e.g. @w_300,h_200,q_80,f_webp).
type ProcessingParams struct {
	Width   int
	Height  int
	Quality int
	Format  string
}

// IsEmpty reports whether no processing is requested.
func (p ProcessingParams) IsEmpty() bool {
	return p.Width <= 0 && p.Height <= 0 && p.Quality <= 0 && p.Format == ""
}

// ProcessImageOnTheFly applies resize, quality, and format conversion in a
// single vips pipeline.  GIF images are skipped (returned as-is).  On any
// processing error the original data is returned unchanged so callers can
// always fall back to the source image.
func ProcessImageOnTheFly(data []byte, sourceFormat string, p ProcessingParams) []byte {
	if p.IsEmpty() {
		return data
	}

	format := normalizeExportFormat(sourceFormat)
	if format == "gif" {
		return data
	}

	needResize := p.Width > 0 || p.Height > 0
	needQuality := p.Quality > 0 && p.Quality < 100
	needFormat := p.Format != "" && normalizeExportFormat(p.Format) != format

	if !needResize && !needQuality && !needFormat {
		return data
	}

	params := vips.NewImportParams()
	params.AutoRotate.Set(true)
	params.FailOnError.Set(true)

	img, err := vips.LoadImageFromBuffer(data, params)
	if err != nil {
		return data
	}
	defer img.Close()

	origW := img.Width()
	origH := img.Height()

	if needResize && origW > 0 && origH > 0 {
		scaleW, scaleH := 1.0, 1.0
		if p.Width > 0 && p.Width < origW {
			scaleW = float64(p.Width) / float64(origW)
		}
		if p.Height > 0 && p.Height < origH {
			scaleH = float64(p.Height) / float64(origH)
		}
		scale := scaleW
		if scaleH < scaleW {
			scale = scaleH
		}
		if scale < 1.0 {
			if err := img.Resize(scale, vips.KernelAuto); err != nil {
				return data
			}
		}
	}

	exportFormat := format
	if needFormat {
		exportFormat = normalizeExportFormat(p.Format)
	}
	exporter, ok := exportRegistry[exportFormat]
	if !ok {
		exporter = exportRegistry["jpeg"]
	}

	quality := 100
	if needQuality {
		quality = p.Quality
	}

	out, err := exporter.Export(img, quality, false)
	if err != nil {
		return data
	}
	return out
}

func normalizeExportFormat(format string) string {
	switch format {
	case "", "jpg":
		return "jpeg"
	case "tif":
		return "tiff"
	default:
		return format
	}
}

// MimeTypeForFormat returns the MIME type for a given image format string.
// Returns empty string for unknown formats.
func MimeTypeForFormat(format string) string {
	switch normalizeExportFormat(format) {
	case "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "webp":
		return "image/webp"
	case "gif":
		return "image/gif"
	default:
		return ""
	}
}

func DecodeImageDimensions(r io.Reader) (int, int, error) {
	img, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
}

// Format exporter implementations

type jpegExporter struct{}

func (e *jpegExporter) Name() string { return "jpeg" }
func (e *jpegExporter) Export(img *vips.ImageRef, quality int, stripMetadata bool) ([]byte, error) {
	p := vips.NewJpegExportParams()
	if quality > 0 && quality < 100 {
		p.Quality = quality
	}
	if stripMetadata {
		p.StripMetadata = true
	}
	out, _, err := img.ExportJpeg(p)
	return out, err
}

type pngExporter struct{}

func (e *pngExporter) Name() string { return "png" }
func (e *pngExporter) Export(img *vips.ImageRef, quality int, stripMetadata bool) ([]byte, error) {
	p := vips.NewPngExportParams()
	if stripMetadata {
		p.StripMetadata = true
	}
	out, _, err := img.ExportPng(p)
	return out, err
}

type webpExporter struct{}

func (e *webpExporter) Name() string { return "webp" }
func (e *webpExporter) Export(img *vips.ImageRef, quality int, stripMetadata bool) ([]byte, error) {
	p := vips.NewWebpExportParams()
	if quality > 0 && quality <= 100 {
		p.Quality = quality
	}
	if stripMetadata {
		p.StripMetadata = true
	}
	out, _, err := img.ExportWebp(p)
	return out, err
}

type gifExporter struct{}

func (e *gifExporter) Name() string { return "gif" }
func (e *gifExporter) Export(img *vips.ImageRef, quality int, stripMetadata bool) ([]byte, error) {
	p := vips.NewGifExportParams()
	if stripMetadata {
		p.StripMetadata = true
	}
	out, _, err := img.ExportGIF(p)
	return out, err
}

func init() {
	RegisterExporter(&jpegExporter{})
	RegisterExporter(&pngExporter{})
	RegisterExporter(&webpExporter{})
	RegisterExporter(&gifExporter{})
}
