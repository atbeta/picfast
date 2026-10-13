package service

import (
	"errors"
	"image"
	_ "image/gif"  // register decoder for pure-Go paths and DecodeConfig probes
	_ "image/jpeg" // same
	_ "image/png"  // same
	"io"

	_ "golang.org/x/image/bmp"  // register decoder
	_ "golang.org/x/image/webp" // register decoder
)

type ProcessedImage struct {
	Data   []byte
	Width  int
	Height int
}

// ErrFormatUnsupported marks output formats the active image backend cannot
// encode (e.g. WebP output under the pure-Go build).
var ErrFormatUnsupported = errors.New("image format unsupported by this build")

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
