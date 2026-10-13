//go:build novips

package service

// Capabilities reports the feature set of the pure-Go image backend built
// with -tags novips: no WebP encoding and no text watermark compositing.
func Capabilities() ImageCapabilities {
	return ImageCapabilities{Backend: "go", WebPOutput: false, Watermark: false}
}

// StartupImaging is a no-op for the pure-Go backend: no native library needs
// initialization.
func StartupImaging() {}

// ShutdownImaging is a no-op for the pure-Go backend.
func ShutdownImaging() {}
