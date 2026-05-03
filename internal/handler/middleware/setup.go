package middleware

import (
	"net/http"
	"strings"

	"github.com/atbeta/picfast/internal/handler/httputil"
	"github.com/atbeta/picfast/internal/sqlc"
)

func RequireSetupCompleteForWrites(db *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/api/v1/") || isReadOnlyMethod(r.Method) || r.URL.Path == "/api/v1/setup/admin" {
				next.ServeHTTP(w, r)
				return
			}

			count, err := db.CountUsers(r.Context())
			if err != nil {
				httputil.Fail(w, http.StatusServiceUnavailable, "failed to read setup status")
				return
			}
			if count == 0 {
				httputil.Fail(w, http.StatusConflict, "setup required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isReadOnlyMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
