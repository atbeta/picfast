//go:build !novips

package service

import (
	"log/slog"

	"github.com/davidbyttow/govips/v2/vips"
)

// StartupImaging initializes the libvips backend (shared by server startup
// and maintenance commands).
func StartupImaging() {
	if err := vips.Startup(&vips.Config{
		ConcurrencyLevel: 2,
		MaxCacheFiles:    0,
		MaxCacheMem:      100 * 1024 * 1024,
		MaxCacheSize:     500,
	}); err != nil {
		slog.Error("failed to start vips", "error", err)
		panic(err)
	}
}

// ShutdownImaging releases libvips resources.
func ShutdownImaging() {
	vips.Shutdown()
}

// Capabilities reports the libvips backend's feature set.
func Capabilities() ImageCapabilities {
	return ImageCapabilities{Backend: "vips", WebPOutput: true, Watermark: true}
}
