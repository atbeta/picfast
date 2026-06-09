package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

func TestIsTrustedProxyIPv4MappedIPv6(t *testing.T) {
	SetTrustedProxies([]string{"192.0.2.0/24"})

	if !isTrustedProxy("192.0.2.1:1234") {
		t.Error("plain IPv4 should match")
	}
	if !isTrustedProxy("[::ffff:192.0.2.1]:1234") {
		t.Error("IPv4-mapped IPv6 should match IPv4 CIDR")
	}
	if !isTrustedProxy("[::ffff:c000:0201]:1234") {
		t.Error("hex-form IPv4-mapped IPv6 should match IPv4 CIDR")
	}
	if isTrustedProxy("198.51.100.1:1234") {
		t.Error("untrusted IPv4 should not match")
	}
}

func TestIsTrustedProxyIPv6CIDR(t *testing.T) {
	SetTrustedProxies([]string{"2001:db8::/32"})

	if !isTrustedProxy("[2001:db8::1]:443") {
		t.Error("IPv6 in trusted CIDR should match")
	}
	if isTrustedProxy("[2001:db9::1]:443") {
		t.Error("IPv6 outside trusted CIDR should not match")
	}
}

func TestIsTrustedProxyInvalidAddr(t *testing.T) {
	SetTrustedProxies([]string{"192.0.2.0/24"})

	if isTrustedProxy("not-an-ip:1234") {
		t.Error("invalid address should not match")
	}
}

func TestSetTrustedProxies(t *testing.T) {
	SetTrustedProxies([]string{})
	if isTrustedProxy("192.0.2.1:1234") {
		t.Error("empty proxy list should match nothing")
	}
}

func TestSetAccessTokenCookieFlags(t *testing.T) {
	SetTrustedProxies([]string{"192.0.2.0/24", "127.0.0.0/8"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.Header.Set("X-Forwarded-Proto", "https")

	writeAuthTokens(rec, req, &domain.AuthTokens{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
		TokenType:    "Bearer",
	})

	cookies := rec.Result().Cookies()
	var authCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == accessTokenCookieName {
			authCookie = cookie
			break
		}
	}
	if authCookie == nil {
		t.Fatal("missing picfast_token cookie")
	}
	if authCookie.Value != "access-token" {
		t.Fatalf("cookie value = %q", authCookie.Value)
	}
	if !authCookie.HttpOnly {
		t.Fatal("expected HttpOnly cookie")
	}
	if !authCookie.Secure {
		t.Fatal("expected Secure cookie over TLS")
	}
	if authCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("SameSite = %v, want Lax", authCookie.SameSite)
	}
}

func TestClearAccessTokenCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	clearAccessTokenCookie(rec, req)

	cookies := rec.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == accessTokenCookieName && cookie.MaxAge != -1 {
			t.Fatalf("expected cleared cookie, got MaxAge=%d", cookie.MaxAge)
		}
	}
}
