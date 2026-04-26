package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/atbeta/picfast/internal/handler"
)

type RateLimiter struct {
	mu       sync.Mutex
	windows  map[string]*slidingWindow
	max      int
	interval time.Duration
}

type slidingWindow struct {
	count    int
	resetAt  time.Time
}

func NewRateLimiter(max int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		windows:  make(map[string]*slidingWindow),
		max:      max,
		interval: interval,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	w, exists := rl.windows[key]
	if !exists || now.After(w.resetAt) {
		rl.windows[key] = &slidingWindow{count: 1, resetAt: now.Add(rl.interval)}
		return true
	}

	w.count++
	return w.count <= rl.max
}

func RateLimit(limiter *RateLimiter, keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			if !limiter.Allow(key) {
				handler.Fail(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
