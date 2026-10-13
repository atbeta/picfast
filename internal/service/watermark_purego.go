//go:build novips

package service

// ApplyWatermark is unsupported under the pure-Go backend: it has no font
// renderer, so it always returns ErrWatermarkUnsupported.
func ApplyWatermark(data []byte, cfg WatermarkConfig, formatHint string, quality int) ([]byte, error) {
	return nil, ErrWatermarkUnsupported
}
