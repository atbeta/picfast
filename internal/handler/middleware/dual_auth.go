package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/handler/httputil"
	"github.com/atbeta/picfast/internal/sqlc"
)

func apiTokenAuth(r *http.Request, db *sqlc.Queries) (context.Context, []string, bool) {
	token := r.Header.Get("X-API-Token")
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	if token == "" {
		return nil, nil, false
	}

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	row, err := db.GetAPITokenByHash(r.Context(), tokenHash)
	if err != nil {
		return nil, nil, false
	}

	if row.ExpiresAt.Valid && row.ExpiresAt.Time.Before(time.Now()) {
		return nil, nil, false
	}

	if row.Status != int16(domain.UserStatusActive) {
		return nil, nil, false
	}

	go db.UpdateAPITokenLastUsed(context.Background(), row.ID)

	var scopes []string
	if err := json.Unmarshal(row.Scopes, &scopes); err != nil || len(scopes) == 0 {
		scopes = []string{"read", "write"}
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, domain.ContextKeyUserID, row.UserID)
	ctx = context.WithValue(ctx, domain.ContextKeyRole, domain.UserRole(row.Role))
	ctx = context.WithValue(ctx, domain.ContextKeyGroupID, domain.PgInt8Val(row.GroupID))
	ctx = context.WithValue(ctx, domain.ContextKeyScopes, scopes)
	return ctx, scopes, true
}

func enrichWithScopes(ctx context.Context, info *AuthInfo, scopes []string) context.Context {
	ctx = enrichContext(ctx, info)
	ctx = context.WithValue(ctx, domain.ContextKeyScopes, scopes)
	return ctx
}

// DualAuth accepts either a JWT access token or an API token (img_...).
// It tries JWT first; if that fails it falls back to API token lookup.
func DualAuth(jwtAuth *JWTAuthenticator, db *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if info, err := jwtAuth.Authenticate(r); err == nil {
				next.ServeHTTP(w, r.WithContext(enrichWithScopes(r.Context(), info, nil)))
				return
			}

			if ctx, _, ok := apiTokenAuth(r, db); ok {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			httputil.Fail(w, http.StatusUnauthorized, "unauthorized")
		})
	}
}

// OptionalDualAuth behaves like DualAuth but also allows unauthenticated requests.
func OptionalDualAuth(jwtAuth *JWTAuthenticator, db *sqlc.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if info, err := jwtAuth.Authenticate(r); err == nil {
				r = r.WithContext(enrichWithScopes(r.Context(), info, nil))
			} else if ctx, _, ok := apiTokenAuth(r, db); ok {
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireScope is middleware that checks for a specific scope in context.
// JWT-authenticated requests (scopes == nil) always pass through.
// API token requests must have the required scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scopes := r.Context().Value(domain.ContextKeyScopes)
			if scopes == nil {
				// JWT sessions have no scope restriction.
				next.ServeHTTP(w, r)
				return
			}
			scopeList, ok := scopes.([]string)
			if !ok || !hasScope(scopeList, scope) {
				httputil.Fail(w, http.StatusForbidden, "insufficient scope: "+scope)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasScope(scopes []string, target string) bool {
	for _, s := range scopes {
		if s == target {
			return true
		}
	}
	return false
}