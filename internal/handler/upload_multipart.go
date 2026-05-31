package handler

import (
	"errors"
	"fmt"
	"net/http"
)

func multipartParseErrorMessage(err error, maxUploadBytes int64) string {
	if err == nil {
		return "failed to parse multipart form"
	}
	var maxErr *http.MaxBytesError
	if errors.As(err, &maxErr) && maxUploadBytes > 0 {
		limitMiB := float64(maxUploadBytes) / (1024 * 1024)
		return fmt.Sprintf("upload exceeds size limit (max %.1f MiB)", limitMiB)
	}
	return "failed to parse multipart form"
}
