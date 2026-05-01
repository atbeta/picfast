package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/handler"
	"github.com/atbeta/picfast/internal/handler/httputil"
)

// AuthInfo carries the authenticated user's identity.
type AuthInfo struct {
	UserID  int64
	Role    domain.UserRole
	GroupID int64
}

// Authenticator extracts and validates credentials from a request.
type Authenticator interface {
	Authenticate(r *http.Request) (*AuthInfo, error)
}

// JWTAuthenticator validates Bearer JWT tokens.
type JWTAuthenticator struct {
	jwtSvc *handler.JWTService
}

func NewJWTAuthenticator(jwtSvc *handler.JWTService) *JWTAuthenticator {
	return &JWTAuthenticator{jwtSvc: jwtSvc}
}

func (a *JWTAuthenticator) Authenticate(r *http.Request) (*AuthInfo, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		if cookie, err := r.Cookie("picfast_token"); err == nil {
			authHeader = "Bearer " + cookie.Value
		} else if token := r.URL.Query().Get("token"); token != "" {
			authHeader = "Bearer " + token
		} else {
			return nil, errMissingCredentials
		}
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return nil, errInvalidScheme
	}

	claims, err := a.jwtSvc.ValidateAccessToken(parts[1])
	if err != nil {
		return nil, err
	}

	return &AuthInfo{
		UserID:  claims.UserID,
		Role:    claims.Role,
		GroupID: claims.GroupID,
	}, nil
}

type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }

var (
	errMissingCredentials = &authError{"missing credentials"}
	errInvalidScheme      = &authError{"invalid auth scheme"}
)

func enrichContext(ctx context.Context, info *AuthInfo) context.Context {
	ctx = context.WithValue(ctx, domain.ContextKeyUserID, info.UserID)
	ctx = context.WithValue(ctx, domain.ContextKeyRole, info.Role)
	ctx = context.WithValue(ctx, domain.ContextKeyGroupID, info.GroupID)
	return ctx
}

func Auth(authn Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			info, err := authn.Authenticate(r)
			if err != nil {
				httputil.Fail(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			next.ServeHTTP(w, r.WithContext(enrichContext(r.Context(), info)))
		})
	}
}

func OptionalAuth(authn Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if info, err := authn.Authenticate(r); err == nil {
				r = r.WithContext(enrichContext(r.Context(), info))
			}
			next.ServeHTTP(w, r)
		})
	}
}
