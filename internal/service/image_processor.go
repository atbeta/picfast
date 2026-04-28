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

	format := saveFormat
	if format == "" {
		format = "jpeg"
	}

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
