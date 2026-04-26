package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/handler"
)

func Auth(jwtSvc *handler.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				handler.Fail(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				handler.Fail(w, http.StatusUnauthorized, "invalid authorization header format")
				return
			}

			claims, err := jwtSvc.ValidateAccessToken(parts[1])
			if err != nil {
				handler.Fail(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, domain.ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, domain.ContextKeyRole, claims.Role)
			ctx = context.WithValue(ctx, domain.ContextKeyGroupID, claims.GroupID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OptionalAuth(jwtSvc *handler.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				if claims, err := jwtSvc.ValidateAccessToken(parts[1]); err == nil {
					ctx := r.Context()
					ctx = context.WithValue(ctx, domain.ContextKeyUserID, claims.UserID)
					ctx = context.WithValue(ctx, domain.ContextKeyRole, claims.Role)
					ctx = context.WithValue(ctx, domain.ContextKeyGroupID, claims.GroupID)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
