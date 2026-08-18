package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)

	if allowed, _ := rl.Allow("key1"); !allowed {
		t.Fatal("expected first request to be allowed")
	}
	if allowed, _ := rl.Allow("key1"); !allowed {
		t.Fatal("expected second request to be allowed")
	}
	if allowed, retryAfter := rl.Allow("key1"); allowed {
		t.Fatal("expected third request to be denied")
	} else if retryAfter <= 0 {
		t.Fatal("expected retry-after to be positive for denied request")
	}

	// Different key should be allowed
	if allowed, _ := rl.Allow("key2"); !allowed {
		t.Fatal("expected different key to be allowed")
	}
}

func TestRateLimiter_WindowReset(t *testing.T) {
	rl := NewRateLimiter(1, 50*time.Millisecond)

	if allowed, _ := rl.Allow("key"); !allowed {
		t.Fatal("expected first request to be allowed")
	}
	if allowed, _ := rl.Allow("key"); allowed {
		t.Fatal("expected second request to be denied")
	}

	// Wait for window to reset
	time.Sleep(60 * time.Millisecond)
	if allowed, _ := rl.Allow("key"); !allowed {
		t.Fatal("expected request after reset to be allowed")
	}
}

func TestRateLimiter_PrunesExpiredWindows(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	rl.windows["expired"] = &slidingWindow{count: 1, resetAt: time.Now().Add(-time.Minute)}

	if allowed, _ := rl.Allow("fresh"); !allowed {
		t.Fatal("expected fresh request to be allowed")
	}
	if _, exists := rl.windows["expired"]; exists {
		t.Fatal("expected expired window to be pruned")
	}
}

func TestRateLimit_Middleware(t *testing.T) {
	rl := NewRateLimiter(2, time.Second)
	mw := RateLimit("test", rl, func(r *http.Request) string { return r.RemoteAddr })

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
	if rec3.Header().Get("Retry-After") == "" {
		t.Fatal("third request: expected Retry-After header")
	}
}
