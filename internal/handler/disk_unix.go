//go:build !windows

package handler

import (
	"syscall"

	"github.com/atbeta/picfast/internal/config"
)

func diskInfo(cfg *config.Config) map[string]any {
	root := cfg.Storage.LocalRoot
	if root == "" {
		root = "."
	}
	var stat syscall.Statfs_t
	if err := syscall.Statfs(root, &stat); err != nil {
		return map[string]any{"healthy": false, "error": err.Error()}
	}
	freeBytes := stat.Bavail * uint64(stat.Bsize)
	totalBytes := stat.Blocks * uint64(stat.Bsize)
	return map[string]any{
		"healthy":     true,
		"path":        root,
		"total_bytes": totalBytes,
		"free_bytes":  freeBytes,
	}
}
