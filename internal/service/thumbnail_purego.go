//go:build novips

package service

import (
	"time"

	picmetrics "github.com/atbeta/picfast/internal/metrics"
)

// renderThumbnailCore renders the thumbnail PNG using the pure-Go backend.
func renderThumbnailCore(data []byte) ([]byte, error) {
	start := time.Now()

	// Bound concurrent decodes; held until the PNG encode below completes.
	decodeSem <- struct{}{}
	defer func() { <-decodeSem }()

	img, err := decodeAutoOriented(data)
	if err != nil {
		picmetrics.ObserveThumbnailGeneration("error", "decode_failed", time.Since(start))
		return nil, err
	}

	thumb := resizeFit(img, ThumbnailMaxSize, ThumbnailMaxSize)

	out, err := encodePNG(thumb)
	if err != nil {
		picmetrics.ObserveThumbnailGeneration("error", "encode_failed", time.Since(start))
		return nil, err
	}

	picmetrics.ObserveThumbnailGeneration("success", picmetrics.ReasonNone, time.Since(start))
	return out, nil
}
