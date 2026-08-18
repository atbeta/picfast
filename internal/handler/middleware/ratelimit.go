package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/atbeta/picfast/internal/handler/httputil"
	picmetrics "github.com/atbeta/picfast/internal/metrics"
)

// RateLimiter decides whether a request identified by key should be allowed.
type RateLimiter interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}

// MemoryRateLimiter is an in-memory sliding-window rate limiter.
type MemoryRateLimiter struct {
	mu         sync.Mutex
	windows    map[string]*slidingWindow
	max        int
	interval   time.Duration
	lastPruned time.Time
	pruneEvery time.Duration
}

type slidingWindow struct {
	count   int
	resetAt time.Time
}

func NewRateLimiter(max int, interval time.Duration) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		windows:    make(map[string]*slidingWindow),
		max:        max,
		interval:   interval,
		pruneEvery: time.Minute,
	}
}

func (rl *MemoryRateLimiter) Allow(key string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	rl.pruneExpired(now)
	w, exists := rl.windows[key]
	if !exists || now.After(w.resetAt) {
		rl.windows[key] = &slidingWindow{count: 1, resetAt: now.Add(rl.interval)}
		return true, 0
	}

	w.count++
	if w.count <= rl.max {
		return true, 0
	}
	retryAfter := time.Until(w.resetAt)
	if retryAfter < 0 {
		retryAfter = 0
	}
	return false, retryAfter
}

func (rl *MemoryRateLimiter) pruneExpired(now time.Time) {
	if !rl.lastPruned.IsZero() && now.Sub(rl.lastPruned) < rl.pruneEvery {
		return
	}
	for key, w := range rl.windows {
		if now.After(w.resetAt) {
			delete(rl.windows, key)
		}
	}
	rl.lastPruned = now
}

func RateLimit(area string, limiter RateLimiter, keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			allowed, retryAfter := limiter.Allow(key)
			if !allowed {
				picmetrics.ObserveRateLimited(area)
				retryAfterSeconds := int(math.Ceil(retryAfter.Seconds()))
				if retryAfterSeconds < 1 {
					retryAfterSeconds = 1
				}
				w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
				httputil.Fail(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
