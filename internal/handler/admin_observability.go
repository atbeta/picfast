package handler

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AdminObservabilityHandler struct {
	db        *sqlc.Queries
	pool      *pgxpool.Pool
	cfg       *config.Config
	mailReady bool
	startedAt time.Time
}

func NewAdminObservabilityHandler(db *sqlc.Queries, pool *pgxpool.Pool, cfg *config.Config, mailReady bool) *AdminObservabilityHandler {
	return &AdminObservabilityHandler{
		db:        db,
		pool:      pool,
		cfg:       cfg,
		mailReady: mailReady,
		startedAt: time.Now(),
	}
}

func (h *AdminObservabilityHandler) Summary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	server, app := h.cfg.RuntimeSnapshot()

	usage, err := h.loadUsage(ctx)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to load observability summary")
		return
	}

	strategies, err := h.loadStorageStrategyHealth(ctx)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to load storage health")
		return
	}

	Success(w, map[string]any{
		"generated_at":   time.Now().UTC(),
		"uptime_seconds": int64(time.Since(h.startedAt).Seconds()),
		"health": map[string]any{
			"database":   h.databaseHealth(ctx),
			"uploads":    directoryHealth(h.cfg.Storage.LocalRoot),
			"thumbnails": directoryHealth(h.cfg.Storage.ThumbnailDir),
			"mail": map[string]any{
				"healthy": !app.RequireEmailVerification || h.mailReady,
				"ready":   h.mailReady,
			},
		},
		"runtime":            runtimeSummary(),
		"database":           databasePoolSummary(h.pool),
		"usage":              usage,
		"storage_strategies": strategies,
		"config": map[string]any{
			"metrics_enabled":   true,
			"pprof_enabled":     server.EnablePprof,
			"moderation_mode":   app.ModerationMode,
			"audit_upload_logs": app.AuditUploadLogs,
		},
	})
}

func (h *AdminObservabilityHandler) databaseHealth(ctx context.Context) map[string]any {
	if err := h.pool.Ping(ctx); err != nil {
		return map[string]any{"healthy": false, "error": err.Error()}
	}
	return map[string]any{"healthy": true}
}

func directoryHealth(path string) map[string]any {
	if path == "" {
		return map[string]any{"healthy": false, "error": "not configured"}
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return map[string]any{"healthy": false, "path": path, "error": err.Error()}
	}
	if info, err := os.Stat(path); err != nil {
		return map[string]any{"healthy": false, "path": path, "error": err.Error()}
	} else if !info.IsDir() {
		return map[string]any{"healthy": false, "path": path, "error": "not a directory"}
	}
	return map[string]any{"healthy": true, "path": path}
}

func runtimeSummary() map[string]any {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return map[string]any{
		"go_version":         runtime.Version(),
		"goos":               runtime.GOOS,
		"goarch":             runtime.GOARCH,
		"num_cpu":            runtime.NumCPU(),
		"goroutines":         runtime.NumGoroutine(),
		"memory_alloc_bytes": mem.Alloc,
		"memory_sys_bytes":   mem.Sys,
	}
}

func databasePoolSummary(pool *pgxpool.Pool) map[string]any {
	stat := pool.Stat()
	return map[string]any{
		"total_connections":    stat.TotalConns(),
		"acquired_connections": stat.AcquiredConns(),
		"idle_connections":     stat.IdleConns(),
		"max_connections":      stat.MaxConns(),
	}
}

func (h *AdminObservabilityHandler) loadUsage(ctx context.Context) (map[string]any, error) {
	users, err := h.db.CountUsers(ctx)
	if err != nil {
		return nil, err
	}
	images, err := h.db.CountAllImages(ctx)
	if err != nil {
		return nil, err
	}
	pending, err := h.db.CountPendingImages(ctx)
	if err != nil {
		return nil, err
	}

	var storageBytes, uploads24h, auditLogs24h int64
	if err := h.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(size_bytes), 0)::bigint,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours')::bigint
		FROM images
	`).Scan(&storageBytes, &uploads24h); err != nil {
		return nil, err
	}
	if err := h.pool.QueryRow(ctx, `
		SELECT COUNT(*)::bigint
		FROM audit_logs
		WHERE created_at >= NOW() - INTERVAL '24 hours'
	`).Scan(&auditLogs24h); err != nil {
		return nil, err
	}

	return map[string]any{
		"users_total":        users,
		"images_total":       images,
		"storage_bytes":      storageBytes,
		"uploads_24h":        uploads24h,
		"pending_moderation": pending,
		"audit_logs_24h":     auditLogs24h,
	}, nil
}

func (h *AdminObservabilityHandler) loadStorageStrategyHealth(ctx context.Context) ([]map[string]any, error) {
	strategies, err := h.db.ListStrategies(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]map[string]any, 0, len(strategies))
	for _, st := range strategies {
		item := map[string]any{
			"id":      st.ID,
			"name":    st.Name,
			"type":    st.StrategyType,
			"healthy": false,
		}
		store, err := service.GetStorageForStrategy(st)
		if err != nil {
			item["error"] = "failed to init: " + err.Error()
			results = append(results, item)
			continue
		}

		health := store.HealthCheck(ctx)
		if err := store.Close(); err != nil {
			slog.Warn("failed to close storage after health check", "strategy_id", st.ID, "error", err)
		}
		item["healthy"] = health.Healthy
		if health.Error != "" {
			item["error"] = health.Error
		}
		if health.Warning != "" {
			item["warning"] = health.Warning
		}
		results = append(results, item)
	}
	return results, nil
}
