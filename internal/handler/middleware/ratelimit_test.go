package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	if !rl.Allow("key1") {
		t.Fatal("expected first request to be allowed")
	}
	if !rl.Allow("key1") {
		t.Fatal("expected second request to be allowed")
	}
	if rl.Allow("key1") {
		t.Fatal("expected third request to be denied")
	}

	// Different key should be allowed
	if !rl.Allow("key2") {
		t.Fatal("expected different key to be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	if !rl.Allow("key") {
		t.Fatal("expected first request to be allowed")
	}
	if rl.Allow("key") {
		t.Fatal("expected second request to be denied")
	}

	// Wait for window to reset
	time.Sleep(60 * time.Millisecond)
	if !rl.Allow("key") {
		t.Fatal("expected request after reset to be allowed")
	}
}

func TestRateLimit_Middleware(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)
	mw := RateLimit(rl, func(r *http.Request) string { return r.RemoteAddr })

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// First request
	rec1 := httptest.NewRecorder()
	handler.ServeHTTP(rec1, req)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first request: status = %d, want 200", rec1.Code)
	}

	// Second request
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req)
	if rec2.Code != http.StatusOK {
		t.Fatalf("second request: status = %d, want 200", rec2.Code)
	}

	// Third request should be rate limited
	rec3 := httptest.NewRecorder()
	handler.ServeHTTP(rec3, req)
	if rec3.Code != http.StatusTooManyRequests {
		t.Fatalf("third request: status = %d, want 429", rec3.Code)
	}
}
