//go:build !novips

package service

import "github.com/davidbyttow/govips/v2/vips"

// StartupImaging initializes the libvips backend (shared by server startup
// and maintenance commands).
func StartupImaging() {
	vips.Startup(&vips.Config{
		ConcurrencyLevel: 2,
		MaxCacheFiles:    0,
		MaxCacheMem:      100 * 1024 * 1024,
		MaxCacheSize:     500,
	})
}

// ShutdownImaging releases libvips resources.
func ShutdownImaging() {
	vips.Shutdown()
}

// Capabilities reports the libvips backend's feature set.
func Capabilities() ImageCapabilities {
	return ImageCapabilities{Backend: "vips", WebPOutput: true, Watermark: true}
}
