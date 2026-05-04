package metrics

import (
	"errors"
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
