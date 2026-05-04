package metrics

import (
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
