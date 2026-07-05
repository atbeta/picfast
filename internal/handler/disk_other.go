//go:build windows

package handler

import "github.com/atbeta/picfast/internal/config"

func diskInfo(cfg *config.Config) map[string]any {
	return map[string]any{"healthy": true, "path": cfg.Storage.LocalRoot, "note": "disk stats not available on this platform"}
}
