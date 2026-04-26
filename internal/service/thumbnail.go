package service

import (
	"bytes"
	"log/slog"
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
		slog.Warn("thumbnail decode failed", "md5", md5Hash, "ext", extension, "error", err)
		return nil // best effort, don't fail upload
	}

	thumb := imaging.Fit(img, ThumbnailMaxSize, ThumbnailMaxSize, imaging.Lanczos)

	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		slog.Error("thumbnail mkdir failed", "dir", thumbnailDir, "error", err)
		return err
	}

	outPath := filepath.Join(thumbnailDir, md5Hash+".png")
	if err := imaging.Save(thumb, outPath); err != nil {
		slog.Error("thumbnail save failed", "path", outPath, "error", err)
		return err
	}

	slog.Info("thumbnail generated", "path", outPath, "md5", md5Hash)
	return nil
}
