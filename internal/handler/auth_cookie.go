package handler

import (
	"net/http"
	"strings"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/domain"
)

const accessTokenCookieName = "picfast_token"

func SetTrustedProxies(proxies []string) {
	clientip.SetTrustedProxies(proxies)
}

func isTrustedProxy(addr string) bool {
	return clientip.IsTrustedProxy(addr)
}

func originalAddr(r *http.Request) string {
	if v, ok := r.Context().Value(domain.ContextKeyOriginalAddr).(string); ok && v != "" {
		return v
	}
	return r.RemoteAddr
}

func requestIsSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if !isTrustedProxy(originalAddr(r)) {
		return false
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func setAccessTokenCookie(w http.ResponseWriter, r *http.Request, token string, maxAge int64) {
	if token == "" || maxAge <= 0 {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(maxAge),
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearAccessTokenCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   requestIsSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func writeAuthTokens(w http.ResponseWriter, r *http.Request, tokens *domain.AuthTokens) {
	setAccessTokenCookie(w, r, tokens.AccessToken, tokens.ExpiresIn)
	Success(w, tokens)
}
