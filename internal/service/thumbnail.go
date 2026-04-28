package service

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/davidbyttow/govips/v2/vips"
)

const ThumbnailMaxSize = 400

func GenerateThumbnail(data []byte, extension, thumbnailDir, md5Hash string) error {
	// SVG and ICO: skip thumbnail generation
	if extension == "svg" || extension == "ico" {
		return nil
	}

	thumb, err := vips.NewThumbnailFromBuffer(data, ThumbnailMaxSize, ThumbnailMaxSize, vips.InterestingNone)
	if err != nil {
		slog.Warn("thumbnail decode/resize failed", "md5", md5Hash, "ext", extension, "error", err)
		return nil // best effort, don't fail upload
	}
	defer thumb.Close()

	if err := os.MkdirAll(thumbnailDir, 0755); err != nil {
		slog.Error("thumbnail mkdir failed", "dir", thumbnailDir, "error", err)
		return err
	}

	outPath := filepath.Join(thumbnailDir, md5Hash+".png")
	p := vips.NewPngExportParams()
	out, _, err := thumb.ExportPng(p)
	if err != nil {
		slog.Error("thumbnail export failed", "path", outPath, "error", err)
		return err
	}

	if err := os.WriteFile(outPath, out, 0644); err != nil {
		slog.Error("thumbnail write failed", "path", outPath, "error", err)
		return err
	}

	slog.Info("thumbnail generated", "path", outPath, "md5", md5Hash)
	return nil
}
