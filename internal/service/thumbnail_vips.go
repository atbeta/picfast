//go:build !novips

package service

import (
	"time"

	picmetrics "github.com/atbeta/picfast/internal/metrics"
	"github.com/davidbyttow/govips/v2/vips"
)

// renderThumbnailCore renders the thumbnail PNG via libvips.
func renderThumbnailCore(data []byte) ([]byte, error) {
	start := time.Now()

	thumb, err := vips.NewThumbnailFromBuffer(data, ThumbnailMaxSize, ThumbnailMaxSize, vips.InterestingNone)
	if err != nil {
		picmetrics.ObserveThumbnailGeneration("error", "decode_failed", time.Since(start))
		return nil, err
	}
	defer thumb.Close()

	out, _, err := thumb.ExportPng(vips.NewPngExportParams())
	if err != nil {
		picmetrics.ObserveThumbnailGeneration("error", "encode_failed", time.Since(start))
		return nil, err
	}

	picmetrics.ObserveThumbnailGeneration("success", picmetrics.ReasonNone, time.Since(start))
	return out, nil
}
