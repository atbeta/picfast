package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

func TestSetAccessTokenCookieFlags(t *testing.T) {
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
