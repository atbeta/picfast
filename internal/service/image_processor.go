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

	var out []byte
	var exportErr error

	switch saveFormat {
	case "png":
		p := vips.NewPngExportParams()
		if stripExif {
			p.StripMetadata = true
		}
		out, _, exportErr = img.ExportPng(p)
	case "webp":
		p := vips.NewWebpExportParams()
		if quality > 0 && quality <= 100 {
			p.Quality = quality
		}
		if stripExif {
			p.StripMetadata = true
		}
		out, _, exportErr = img.ExportWebp(p)
	case "gif":
		p := vips.NewGifExportParams()
		if stripExif {
			p.StripMetadata = true
		}
		out, _, exportErr = img.ExportGIF(p)
	default: // jpeg
		p := vips.NewJpegExportParams()
		if quality > 0 && quality < 100 {
			p.Quality = quality
		}
		if stripExif {
			p.StripMetadata = true
		}
		out, _, exportErr = img.ExportJpeg(p)
	}

	if exportErr != nil {
		return nil, exportErr
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
