package handler

import (
	"errors"
	"net/http"
	"testing"
)

func TestMultipartParseErrorMessage(t *testing.T) {
	msg := multipartParseErrorMessage(errors.New("boom"), 50<<20)
	if msg != "failed to parse multipart form" {
		t.Fatalf("unexpected message: %q", msg)
	}

	maxErr := &http.MaxBytesError{Limit: 50 << 20}
	msg = multipartParseErrorMessage(maxErr, 50<<20)
	if msg != "upload exceeds size limit (max 50.0 MiB)" {
		t.Fatalf("unexpected message: %q", msg)
	}
}
