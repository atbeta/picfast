package service

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/disintegration/imaging"
)

type ProcessedImage struct {
	Data   []byte
	Width  int
	Height int
}

func ProcessImage(data []byte, saveFormat string, quality int) (*ProcessedImage, error) {
	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Determine output format
	format := imaging.JPEG
	if saveFormat == "png" {
		format = imaging.PNG
	} else if saveFormat == "gif" {
		format = imaging.GIF
	} else if saveFormat == "webp" {
		// imaging doesn't support webp, keep original
		return &ProcessedImage{Data: data, Width: width, Height: height}, nil
	}

	// Re-encode if format change or quality adjustment needed
	buf := new(bytes.Buffer)
	if format == imaging.JPEG && quality > 0 && quality < 100 {
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return nil, err
		}
	} else if format == imaging.PNG {
		if err := png.Encode(buf, img); err != nil {
			return nil, err
		}
	} else {
		if err := imaging.Encode(buf, img, format); err != nil {
			return nil, err
		}
	}

	return &ProcessedImage{Data: buf.Bytes(), Width: width, Height: height}, nil
}

func DecodeImageDimensions(r io.Reader) (int, int, error) {
	img, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, err
	}
	return img.Width, img.Height, nil
}
