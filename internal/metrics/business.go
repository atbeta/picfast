package metrics

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const ReasonNone = "none"

var (
	uploadsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_uploads_total",
			Help: "Total number of PicFast upload attempts",
		},
		[]string{"source", "result", "reason"},
	)

	uploadBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_upload_bytes_total",
			Help: "Total number of bytes received by PicFast uploads",
		},
		[]string{"source"},
	)

	uploadDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "picfast_upload_duration_seconds",
			Help:    "PicFast upload request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"source", "result"},
	)

	imageServesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_image_serves_total",
			Help: "Total number of PicFast image and thumbnail serve attempts",
		},
		[]string{"kind", "result"},
	)

	imageServeDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "picfast_image_serve_duration_seconds",
			Help:    "PicFast image and thumbnail serve duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"kind", "result"},
	)

	thumbnailGenerationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_thumbnail_generations_total",
			Help: "Total number of PicFast thumbnail generation attempts",
		},
		[]string{"result", "reason"},
	)

	thumbnailGenerationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "picfast_thumbnail_generation_duration_seconds",
			Help:    "PicFast thumbnail generation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)

	storageOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_storage_operations_total",
			Help: "Total number of PicFast storage backend operations",
		},
		[]string{"operation", "backend", "result", "reason"},
	)

	storageOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "picfast_storage_operation_duration_seconds",
			Help:    "PicFast storage backend operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "backend", "result"},
	)

	cleanupRunsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_cleanup_runs_total",
			Help: "Total number of PicFast cleanup task runs",
		},
		[]string{"result"},
	)

	cleanupDeletedImagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_cleanup_deleted_images_total",
			Help: "Total number of images removed by PicFast cleanup tasks",
		},
		[]string{"reason"},
	)

	cleanupDeletedBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_cleanup_deleted_bytes_total",
			Help: "Total bytes reclaimed by PicFast cleanup tasks",
		},
		[]string{"reason"},
	)

	cleanupDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "picfast_cleanup_duration_seconds",
			Help:    "PicFast cleanup task run duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"result"},
	)

	authAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_auth_attempts_total",
			Help: "Total number of PicFast authentication attempts",
		},
		[]string{"kind", "result", "reason"},
	)

	rateLimitedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_rate_limited_total",
			Help: "Total number of PicFast requests rejected by rate limiting",
		},
		[]string{"area"},
	)

	moderationActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "picfast_moderation_actions_total",
			Help: "Total number of PicFast moderation actions",
		},
		[]string{"mode", "action", "result"},
	)
)

func ObserveUpload(source, result, reason string, bytes int64, duration time.Duration) {
	reason = normalizeReason(reason)
	uploadsTotal.WithLabelValues(source, result, reason).Inc()
	uploadDuration.WithLabelValues(source, result).Observe(duration.Seconds())
	if result == "success" && bytes > 0 {
		uploadBytesTotal.WithLabelValues(source).Add(float64(bytes))
	}
}

func ObserveImageServe(kind, result string, duration time.Duration) {
	imageServesTotal.WithLabelValues(kind, result).Inc()
	imageServeDuration.WithLabelValues(kind, result).Observe(duration.Seconds())
}

func ObserveThumbnailGeneration(result, reason string, duration time.Duration) {
	reason = normalizeReason(reason)
	thumbnailGenerationsTotal.WithLabelValues(result, reason).Inc()
	thumbnailGenerationDuration.WithLabelValues(result).Observe(duration.Seconds())
}

func ObserveStorageOperation(operation, backend, result, reason string, duration time.Duration) {
	reason = normalizeReason(reason)
	storageOperationsTotal.WithLabelValues(operation, backend, result, reason).Inc()
	storageOperationDuration.WithLabelValues(operation, backend, result).Observe(duration.Seconds())
}

func ObserveCleanupRun(result string, duration time.Duration) {
	cleanupRunsTotal.WithLabelValues(result).Inc()
	cleanupDuration.WithLabelValues(result).Observe(duration.Seconds())
}

func ObserveCleanupDeleted(reason string, images int, bytes int64) {
	if images > 0 {
		cleanupDeletedImagesTotal.WithLabelValues(reason).Add(float64(images))
	}
	if bytes > 0 {
		cleanupDeletedBytesTotal.WithLabelValues(reason).Add(float64(bytes))
	}
}

func ObserveAuthAttempt(kind, result, reason string) {
	reason = normalizeReason(reason)
	authAttemptsTotal.WithLabelValues(kind, result, reason).Inc()
}

func ObserveRateLimited(area string) {
	rateLimitedTotal.WithLabelValues(area).Inc()
}

func ObserveModerationAction(mode, action, result string) {
	moderationActionsTotal.WithLabelValues(mode, action, result).Inc()
}

// ClassifyStorageError 把存储后端错误归入低基数枚举，禁止把原始错误文本放进 label。
func ClassifyStorageError(err error) string {
	switch {
	case err == nil:
		return ReasonNone
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		return "timeout"
	case errors.Is(err, os.ErrNotExist):
		return "not_found"
	default:
		return "unknown"
	}
}

func ClassifyUploadError(err error) string {
	if err == nil {
		return ReasonNone
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "capacity exceeded"):
		return "quota_exceeded"
	case strings.Contains(msg, "rate limit"):
		return "rate_limited"
	case strings.Contains(msg, "guest upload is disabled"):
		return "denied"
	case strings.Contains(msg, "storage"), strings.Contains(msg, "write file"), strings.Contains(msg, "init storage"):
		return "storage_failed"
	case strings.Contains(msg, "database"), strings.Contains(msg, "db"), strings.Contains(msg, "save image record"):
		return "db_failed"
	case strings.Contains(msg, "invalid"), strings.Contains(msg, "file"):
		return "invalid_file"
	default:
		return "unknown"
	}
}

func normalizeReason(reason string) string {
	if reason == "" {
		return ReasonNone
	}
	return reason
}
