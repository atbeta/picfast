package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/handler"
	"github.com/atbeta/picfast/internal/sqlc"
)

func APITokenAuth(db *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Support X-API-Token header or Authorization: Bearer <token>
			token := r.Header.Get("X-API-Token")
			if token == "" {
				authHeader := r.Header.Get("Authorization")
				if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
					token = authHeader[7:]
				}
			}

			if token == "" {
				handler.Fail(w, http.StatusUnauthorized, "missing API token")
				return
			}

			hash := sha256.Sum256([]byte(token))
			tokenHash := hex.EncodeToString(hash[:])

			ctx := r.Context()
			row, err := db.GetAPITokenByHash(ctx, tokenHash)
			if err != nil {
				handler.Fail(w, http.StatusUnauthorized, "invalid API token")
				return
			}

			// Check expiration
			if row.ExpiresAt.Valid && row.ExpiresAt.Time.Before(time.Now()) {
				handler.Fail(w, http.StatusUnauthorized, "API token expired")
				return
			}

			// Check user status
			if row.Status != int16(domain.UserStatusActive) {
				handler.Fail(w, http.StatusForbidden, "account is frozen")
				return
			}

			// Update last used (best effort)
			go db.UpdateAPITokenLastUsed(context.Background(), row.ID)

			ctx = context.WithValue(ctx, domain.ContextKeyUserID, row.UserID)
			ctx = context.WithValue(ctx, domain.ContextKeyRole, domain.UserRole(row.Role))
			ctx = context.WithValue(ctx, domain.ContextKeyGroupID, domain.PgInt8Val(row.GroupID))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
