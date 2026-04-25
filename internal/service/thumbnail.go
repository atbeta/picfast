package service

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
)

const ThumbnailMaxSize = 400

func GenerateThumbnail(data []byte, extension, thumbnailDir, md5Hash string) error {
	// SVG and ICO: skip thumbnail generation
	if extension == "svg" || extension == "ico" {
		return nil
	}

	img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
	if err != nil {
		return nil // best effort, don't fail upload
	}

	thumb := imaging.Fit(img, ThumbnailMaxSize, ThumbnailMaxSize, imaging.Lanczos)

	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		return err
	}

	outPath := filepath.Join(thumbnailDir, md5Hash+".png")
	return imaging.Save(thumb, outPath)
}
