package middleware

import (
	"net/http"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/handler"
)

func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, ok := r.Context().Value(domain.ContextKeyRole).(domain.UserRole)
		if !ok || role != domain.RoleAdmin {
			handler.Fail(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
