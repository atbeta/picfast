package handler

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/atbeta/picfast/internal/domain"
)

const accessTokenCookieName = "picfast_token"

var (
	trustedProxyMu    sync.RWMutex
	trustedProxyCIDRs []*net.IPNet
)

func SetTrustedProxies(proxies []string) {
	trustedProxyMu.Lock()
	defer trustedProxyMu.Unlock()
	trustedProxyCIDRs = nil
	for _, p := range proxies {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.Contains(p, "/") {
			if strings.Contains(p, ":") {
				p += "/128"
			} else {
				p += "/32"
			}
		}
		_, cidr, err := net.ParseCIDR(p)
		if err != nil {
			slog.Warn("ignoring invalid trusted_proxies entry", "entry", p, "error", err)
			continue
		}
		trustedProxyCIDRs = append(trustedProxyCIDRs, cidr)
	}
}

func isTrustedProxy(addr string) bool {
	trustedProxyMu.RLock()
	defer trustedProxyMu.RUnlock()
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		ip = v4
	}
	for _, cidr := range trustedProxyCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
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
