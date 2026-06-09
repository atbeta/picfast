package clientip

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/atbeta/picfast/internal/domain"
)

var (
	mu    sync.RWMutex
	cidrs []*net.IPNet
)

func SetTrustedProxies(proxies []string) {
	mu.Lock()
	defer mu.Unlock()
	cidrs = nil
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
		cidrs = append(cidrs, cidr)
	}
}

// FromRequest extracts the real client IP from an HTTP request.
//
// It uses domain.ContextKeyOriginalAddr (set before RealIP middleware) as the
// connection-address for trusted-proxy checks; this ensures the trust decision
// is based on the actual TCP peer, not on header-supplied values.
//
// Resolution order (only when the original address belongs to a trusted proxy):
//  1. CF-Connecting-IP
//  2. X-Forwarded-For (rightmost untrusted IP)
//  3. X-Real-IP
//  4. r.RemoteAddr (post-RealIP)
func FromRequest(r *http.Request) string {
	addr := originalAddr(r)
	if !isTrustedProxy(addr) {
		return remoteHost(r.RemoteAddr)
	}
	if cfIP := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); cfIP != "" {
		if ip := net.ParseIP(cfIP); ip != nil {
			return ip.String()
		}
	}
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		if ip := rightmostUntrustedIP(forwarded); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		if ip := net.ParseIP(realIP); ip != nil {
			return ip.String()
		}
	}
	return remoteHost(r.RemoteAddr)
}

func IsTrustedProxy(addr string) bool {
	return isTrustedProxy(addr)
}

func originalAddr(r *http.Request) string {
	if v, ok := r.Context().Value(domain.ContextKeyOriginalAddr).(string); ok && v != "" {
		return v
	}
	return r.RemoteAddr
}

func isTrustedProxy(addr string) bool {
	mu.RLock()
	defer mu.RUnlock()
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
	for _, cidr := range cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

func remoteHost(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err == nil && host != "" {
		return host
	}
	return addr
}

func rightmostUntrustedIP(forwarded string) string {
	parts := strings.Split(forwarded, ",")
	for i := len(parts) - 1; i >= 0; i-- {
		ip := net.ParseIP(strings.TrimSpace(parts[i]))
		if ip == nil {
			continue
		}
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}
		if !ipInTrustedCIDRs(ip) {
			return ip.String()
		}
	}
	return ""
}

func ipInTrustedCIDRs(ip net.IP) bool {
	mu.RLock()
	defer mu.RUnlock()
	for _, cidr := range cidrs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
