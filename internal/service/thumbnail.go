package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	picmetrics "github.com/atbeta/picfast/internal/metrics"
	"github.com/atbeta/picfast/internal/service/storage"
)

const ThumbnailMaxSize = 400

// thumbnailPrefix is the object path prefix for thumbnails stored inside an
// image's storage strategy. It is prefixed so lifecycle tooling can tell
// PicFast-managed objects apart from user uploads.
const thumbnailPrefix = "_picfast/thumbnails/"

// ThumbnailObjectPath returns the object path of a thumbnail inside the
// image's storage strategy.
func ThumbnailObjectPath(md5 string) string {
	return thumbnailPrefix + md5 + ".png"
}

// ErrThumbnailSkipped marks image types that never get a thumbnail.
var ErrThumbnailSkipped = errors.New("thumbnail skipped for this image type")

// RenderThumbnail renders a PNG thumbnail for data. It performs no storage IO:
// SVG and ICO inputs return ErrThumbnailSkipped. The pixel work lives in the
// per-backend renderThumbnailCore (libvips or pure Go).
func RenderThumbnail(data []byte, extension string) ([]byte, error) {
	start := time.Now()
	if extension == "svg" || extension == "ico" {
		picmetrics.ObserveThumbnailGeneration("success", "skipped", time.Since(start))
		return nil, ErrThumbnailSkipped
	}
	return renderThumbnailCore(data)
}

// StoreThumbnail renders a thumbnail for data and writes it into the image's
// storage strategy. Skipped types (SVG, ICO) return nil; other render failures
// are returned wrapped so callers can mark the thumbnail as failed instead of
// silently recording it as complete. Storage write failures are returned too.
func StoreThumbnail(ctx context.Context, store storage.Storage, data []byte, extension, md5 string) error {
	png, err := RenderThumbnail(data, extension)
	if err != nil {
		if errors.Is(err, ErrThumbnailSkipped) {
			return nil
		}
		return fmt.Errorf("render thumbnail: %w", err)
	}

	path := ThumbnailObjectPath(md5)
	start := time.Now()
	if err := store.Write(ctx, path, png, "image/png"); err != nil {
		slog.Error("thumbnail write failed", "path", path, "error", err)
		picmetrics.ObserveThumbnailGeneration("error", "write_failed", time.Since(start))
		return err
	}
	slog.Info("thumbnail generated", "path", path, "md5", md5)
	return nil
}

// isNotFoundError reports whether a storage error means the object simply does
// not exist. Backends do not share a typed not-found error, so this matches the
// common shapes.
func isNotFoundError(err error) bool {
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "nosuchkey") ||
		strings.Contains(msg, "no such key") ||
		strings.Contains(msg, "404")
}
