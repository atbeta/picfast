package clientip

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

func TestSetTrustedProxies_Empty(t *testing.T) {
	SetTrustedProxies([]string{})
	mu.RLock()
	defer mu.RUnlock()
	if len(cidrs) != 0 {
		t.Fatalf("expected 0 cidrs, got %d", len(cidrs))
	}
}

func TestSetTrustedProxies_IPv4(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	mu.RLock()
	defer mu.RUnlock()
	if len(cidrs) != 3 {
		t.Fatalf("expected 3 cidrs, got %d", len(cidrs))
	}
}

func TestSetTrustedProxies_NoCIDRSlash(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.1"})
	mu.RLock()
	defer mu.RUnlock()
	if len(cidrs) != 1 {
		t.Fatalf("expected 1 cidr, got %d", len(cidrs))
	}
}

func TestIsTrustedProxy(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12"})

	tests := []struct {
		addr   string
		expect bool
	}{
		{"10.1.2.3:12345", true},
		{"172.20.0.1:8080", true},
		{"192.168.1.1:443", true},  // private IP, auto-trusted
		{"127.0.0.1:9999", true},   // loopback, auto-trusted
		{"169.254.1.1:8080", true}, // link-local, auto-trusted
		{"8.8.8.8:443", false},     // public IP, not in trusted CIDRs
		{"invalid", false},
	}
	for _, tt := range tests {
		if got := IsTrustedProxy(tt.addr); got != tt.expect {
			t.Errorf("IsTrustedProxy(%q) = %v, want %v", tt.addr, got, tt.expect)
		}
	}
}

func TestRemoteHost(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"1.2.3.4:12345", "1.2.3.4"},
		{"[::1]:8080", "::1"},
		{"1.2.3.4", "1.2.3.4"},
	}
	for _, tt := range tests {
		if got := remoteHost(tt.addr); got != tt.want {
			t.Errorf("remoteHost(%q) = %q, want %q", tt.addr, got, tt.want)
		}
	}
}

func TestFromRequest_Direct(t *testing.T) {
	SetTrustedProxies([]string{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	// No original_addr in context → originalAddr falls back to RemoteAddr
	// RemoteAddr 1.2.3.4 is NOT trusted → returns remoteHost(r.RemoteAddr)

	if got := FromRequest(req); got != "1.2.3.4" {
		t.Errorf("FromRequest(direct) = %q, want %q", got, "1.2.3.4")
	}
}

func TestFromRequest_TrustedProxy_XForwardedFor(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:443"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.5.0.1")
	ctx := context.WithValue(req.Context(), domain.ContextKeyOriginalAddr, "10.0.0.1:443")
	req = req.WithContext(ctx)

	if got := FromRequest(req); got != "1.2.3.4" {
		t.Errorf("FromRequest(x-forwarded-for) = %q, want %q", got, "1.2.3.4")
	}
}

func TestFromRequest_TrustedProxy_XRealIP(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:443"
	req.Header.Set("X-Real-IP", "5.6.7.8")
	ctx := context.WithValue(req.Context(), domain.ContextKeyOriginalAddr, "10.0.0.1:443")
	req = req.WithContext(ctx)

	if got := FromRequest(req); got != "5.6.7.8" {
		t.Errorf("FromRequest(x-real-ip) = %q, want %q", got, "5.6.7.8")
	}
}

func TestFromRequest_TrustedProxy_CFConnectingIP(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:443"
	req.Header.Set("CF-Connecting-IP", "9.9.9.9")
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	ctx := context.WithValue(req.Context(), domain.ContextKeyOriginalAddr, "10.0.0.1:443")
	req = req.WithContext(ctx)

	if got := FromRequest(req); got != "9.9.9.9" {
		t.Errorf("FromRequest(cf-connecting-ip) = %q, want %q", got, "9.9.9.9")
	}
}

func TestFromRequest_UntrustedProxy_IgnoresHeaders(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "8.8.8.8:443"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	req.Header.Set("X-Real-IP", "5.6.7.8")
	ctx := context.WithValue(req.Context(), domain.ContextKeyOriginalAddr, "8.8.8.8:443")
	req = req.WithContext(ctx)

	if got := FromRequest(req); got != "8.8.8.8" {
		t.Errorf("FromRequest(untrusted) = %q, want %q", got, "8.8.8.8")
	}
}

func TestFromRequest_PrivateNetworkProxy_ParsesHeaders(t *testing.T) {
	// Docker bridges and reverse proxies on private networks should be
	// auto-trusted so that X-Forwarded-For is respected.
	SetTrustedProxies([]string{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "172.21.0.1:443"
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 10.0.0.5")
	ctx := context.WithValue(req.Context(), domain.ContextKeyOriginalAddr, "172.21.0.1:443")
	req = req.WithContext(ctx)

	if got := FromRequest(req); got != "1.2.3.4" {
		t.Errorf("FromRequest(private-network-proxy) = %q, want %q", got, "1.2.3.4")
	}
}

func TestFromRequest_PrivateNetworkProxy_XRealIP(t *testing.T) {
	SetTrustedProxies([]string{})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:443"
	req.Header.Set("X-Real-IP", "5.6.7.8")
	ctx := context.WithValue(req.Context(), domain.ContextKeyOriginalAddr, "192.168.1.1:443")
	req = req.WithContext(ctx)

	if got := FromRequest(req); got != "5.6.7.8" {
		t.Errorf("FromRequest(private-network-x-real-ip) = %q, want %q", got, "5.6.7.8")
	}
}

func TestFromRequest_FallbackToRemoteAddr(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:443"
	ctx := context.WithValue(req.Context(), domain.ContextKeyOriginalAddr, "10.0.0.1:443")
	req = req.WithContext(ctx)

	if got := FromRequest(req); got != "10.0.0.1" {
		t.Errorf("FromRequest(fallback) = %q, want %q", got, "10.0.0.1")
	}
}

func TestRightmostUntrustedIP(t *testing.T) {
	SetTrustedProxies([]string{"10.0.0.0/8"})
	tests := []struct {
		forwarded string
		want      string
	}{
		{"1.2.3.4, 10.0.0.5", "1.2.3.4"},
		{"10.0.0.5, 10.0.0.6", ""},
		{"1.2.3.4", "1.2.3.4"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := rightmostUntrustedIP(tt.forwarded); got != tt.want {
			t.Errorf("rightmostUntrustedIP(%q) = %q, want %q", tt.forwarded, got, tt.want)
		}
	}
}
