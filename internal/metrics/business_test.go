package metrics

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestClassifyUploadError(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{errors.New("guest storage capacity exceeded"), "quota_exceeded"},
		{errors.New("rate limit exceeded"), "rate_limited"},
		{errors.New("guest upload is disabled"), "denied"},
		{errors.New("failed to write file: nope"), "storage_failed"},
		{errors.New("failed to save image record: nope"), "db_failed"},
		{errors.New("invalid file extension"), "invalid_file"},
		{errors.New("something else"), "unknown"},
	}

	for _, tc := range cases {
		if got := ClassifyUploadError(tc.err); got != tc.want {
			t.Fatalf("ClassifyUploadError(%q) = %q, want %q", tc.err, got, tc.want)
		}
	}
}

func TestClassifyStorageError(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{nil, ReasonNone},
		{context.DeadlineExceeded, "timeout"},
		{fmt.Errorf("wrap: %w", context.Canceled), "timeout"},
		{fmt.Errorf("read: %w", os.ErrNotExist), "not_found"},
		{errors.New("connection refused"), "unknown"},
	}

	for _, tc := range cases {
		if got := ClassifyStorageError(tc.err); got != tc.want {
			t.Fatalf("ClassifyStorageError(%v) = %q, want %q", tc.err, got, tc.want)
		}
	}
}

// 新增指标的 Observe 函数不允许 panic（重复注册、label 数不符都会在此暴露）。
func TestObserveFunctionsDoNotPanic(t *testing.T) {
	ObserveStorageOperation("put", "local", "success", ReasonNone, 0)
	ObserveStorageOperation("get", "s3", "error", "timeout", 0)
	ObserveCleanupRun("success", 0)
	ObserveCleanupDeleted("expired", 3, 1024)
	ObserveAuthAttempt("login", "error", "invalid_credentials")
	ObserveAuthAttempt("oauth", "success", "")
	ObserveRateLimited("login")
	ObserveModerationAction("manual", "approve", "success")
}
