package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/pbeta/imgapi/internal/domain"
)

// StructuredLogger replaces chi's default Logger with structured slog output.
func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(ww, r)

		duration := time.Since(start)

		userID, _ := r.Context().Value(domain.ContextKeyUserID).(int64)
		role, _ := r.Context().Value(domain.ContextKeyRole).(domain.UserRole)

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", ww.Status()),
			slog.Int("bytes", ww.BytesWritten()),
			slog.Duration("duration", duration),
			slog.String("request_id", middleware.GetReqID(r.Context())),
			slog.String("client_ip", r.RemoteAddr),
		}

		if userID != 0 {
			attrs = append(attrs, slog.Int64("user_id", userID))
			attrs = append(attrs, slog.String("role", string(role)))
		}

		slog.LogAttrs(r.Context(), slog.LevelInfo, "request", attrs...)
	})
}
