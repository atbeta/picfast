package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminMaintenanceHandler struct {
	db           *sqlc.Queries
	pool         *pgxpool.Pool
	cfg          *config.Config
	obsHandler   *AdminObservabilityHandler
}

func NewAdminMaintenanceHandler(db *sqlc.Queries, pool *pgxpool.Pool, cfg *config.Config, obsHandler *AdminObservabilityHandler) *AdminMaintenanceHandler {
	return &AdminMaintenanceHandler{
		db:         db,
		pool:       pool,
		cfg:        cfg,
		obsHandler: obsHandler,
	}
}

func (h *AdminMaintenanceHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	usage, err := h.obsHandler.loadUsage(ctx)
	if err != nil {
		slog.Warn("failed to load usage for maintenance", "error", err)
		usage = map[string]any{}
	}

	strategies, err := h.obsHandler.loadStorageStrategyHealth(ctx)
	if err != nil {
		slog.Warn("failed to load strategies for maintenance", "error", err)
		strategies = []map[string]any{}
	}

	risks := h.collectRisks(ctx, usage, strategies)
	diskInfo := h.diskInfo()
	backupInfo := h.backupInfo()
	dbTables := h.tableStats(ctx)
	phashCoverage := h.phashStats(ctx)

	Success(w, map[string]any{
		"generated_at":   time.Now().UTC(),
		"risks":          risks,
		"storage": map[string]any{
			"disk":       diskInfo,
			"strategies": strategies,
		},
		"usage":          usage,
		"backup":         backupInfo,
		"database":       dbTables,
		"phash_coverage": phashCoverage,
	})
}

func (h *AdminMaintenanceHandler) collectRisks(ctx context.Context, usage map[string]any, strategies []map[string]any) []map[string]any {
	var risks []map[string]any

	// Check JWT secret
	if h.cfg.UsesDefaultJWTSecret() {
		risks = append(risks, map[string]any{
			"level":   "warn",
			"code":    "default_jwt_secret",
			"message": "JWT secret is using the default value. Set a strong secret for production.",
		})
	}

	// Check storage strategies health
	for _, st := range strategies {
		healthy, _ := st["healthy"].(bool)
		if !healthy {
			risks = append(risks, map[string]any{
				"level":   "error",
				"code":    "strategy_unhealthy",
				"message": "Storage strategy \"" + getString(st, "name") + "\" is unhealthy: " + getString(st, "error"),
			})
		}
	}

	// Check pending moderation
	if pending, ok := usage["pending_moderation"].(int64); ok && pending > 0 {
		risks = append(risks, map[string]any{
			"level":   "info",
			"code":    "pending_moderation",
			"message": "There are pending images waiting for moderation review.",
			"count":   pending,
		})
	}

	// Check expired images
	expired, err := h.db.CountExpiredImages(ctx)
	if err == nil && expired > 0 {
		risks = append(risks, map[string]any{
			"level":   "info",
			"code":    "expired_images",
			"message": "There are expired images that can be cleaned up.",
			"count":   expired,
		})
	}

	// Check disk space
	disk := h.diskInfo()
	if freeBytes, ok := disk["free_bytes"].(uint64); ok && freeBytes < 1024*1024*1024 { // < 1 GB
		risks = append(risks, map[string]any{
			"level":   "error",
			"code":    "low_disk_space",
			"message": "Disk space is critically low.",
		})
	}

	return risks
}

func (h *AdminMaintenanceHandler) diskInfo() map[string]any {
	root := h.cfg.Storage.LocalRoot
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

func (h *AdminMaintenanceHandler) backupInfo() map[string]any {
	dataDir := h.cfg.Storage.LocalRoot
	if dataDir == "" {
		return map[string]any{"status": "no_storage"}
	}

	backupDir := dataDir + "/backups"
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return map[string]any{"status": "no_backups", "path": backupDir}
	}

	var latest os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if latest == nil || info.ModTime().After(latest.ModTime()) {
			latest = info
		}
	}

	if latest == nil {
		return map[string]any{"status": "no_backups", "path": backupDir}
	}

	return map[string]any{
		"status":    "ok",
		"file":      latest.Name(),
		"size":      latest.Size(),
		"timestamp": latest.ModTime().UTC().Format(time.RFC3339),
	}
}

func getString(m map[string]any, key string) string {
	v, _ := m[key].(string)
	if v == "" {
		v, _ = m[key].(string)
	}
	return v
}

func (h *AdminMaintenanceHandler) tableStats(ctx context.Context) []map[string]any {
	rows, err := h.pool.Query(ctx, `
		SELECT schemaname, relname, n_live_tup
		FROM pg_stat_user_tables
		ORDER BY n_live_tup DESC
	`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var tables []map[string]any
	for rows.Next() {
		var schema, name string
		var rowCount int64
		if err := rows.Scan(&schema, &name, &rowCount); err != nil {
			continue
		}
		tables = append(tables, map[string]any{
			"table": name,
			"rows":  rowCount,
		})
	}
	return tables
}

func (h *AdminMaintenanceHandler) phashStats(ctx context.Context) map[string]any {
	var total, withPhash int64
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM images`).Scan(&total); err != nil {
		return nil
	}
	if err := h.pool.QueryRow(ctx, `SELECT COUNT(*) FROM images WHERE phash IS NOT NULL`).Scan(&withPhash); err != nil {
		return nil
	}
	return map[string]any{
		"total":      total,
		"with_phash": withPhash,
	}
}

func (h *AdminMaintenanceHandler) CleanupExpired(w http.ResponseWriter, r *http.Request) {
	count, err := h.db.CountExpiredImages(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to count expired images")
		return
	}

	result, err := h.db.DeleteExpiredImages(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete expired images")
		return
	}

	SuccessMessage(w, fmt.Sprintf("cleaned %d of %d expired images", len(result), count))
}
