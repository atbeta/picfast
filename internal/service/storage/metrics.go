package storage

import (
	"context"
	"time"

	picmetrics "github.com/atbeta/picfast/internal/metrics"
)

// metricsStorage 在存储后端边界记录操作次数与耗时。
// label 只含 operation/backend/result/枚举 reason，不含 key、bucket、错误原文。
type metricsStorage struct {
	backend string
	next    Storage
}

// WithMetrics wraps a Storage so every operation is observed in Prometheus.
func WithMetrics(backend string, s Storage) Storage {
	if s == nil {
		return nil
	}
	return &metricsStorage{backend: backend, next: s}
}

func (m *metricsStorage) observe(operation string, start time.Time, err error) {
	if err != nil {
		picmetrics.ObserveStorageOperation(operation, m.backend, "error", picmetrics.ClassifyStorageError(err), time.Since(start))
		return
	}
	picmetrics.ObserveStorageOperation(operation, m.backend, "success", picmetrics.ReasonNone, time.Since(start))
}

func (m *metricsStorage) Write(ctx context.Context, path string, data []byte, contentType string) error {
	start := time.Now()
	err := m.next.Write(ctx, path, data, contentType)
	m.observe("put", start, err)
	return err
}

func (m *metricsStorage) Read(ctx context.Context, path string) ([]byte, error) {
	start := time.Now()
	data, err := m.next.Read(ctx, path)
	m.observe("get", start, err)
	return data, err
}

func (m *metricsStorage) Delete(ctx context.Context, path string) error {
	start := time.Now()
	err := m.next.Delete(ctx, path)
	m.observe("delete", start, err)
	return err
}

func (m *metricsStorage) URL(pathname string) string {
	return m.next.URL(pathname)
}

func (m *metricsStorage) HealthCheck(ctx context.Context) HealthResult {
	start := time.Now()
	res := m.next.HealthCheck(ctx)
	var err error
	if !res.Healthy {
		err = errUnhealthy
	}
	m.observe("healthcheck", start, err)
	return res
}

func (m *metricsStorage) Close() error {
	return m.next.Close()
}

type storageError string

func (e storageError) Error() string { return string(e) }

const errUnhealthy = storageError("health check failed")
