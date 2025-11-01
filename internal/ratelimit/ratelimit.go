package ratelimit

import (
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for HTTP requests
type RateLimiter struct {
	limiters sync.Map
	rpm      int
	burst    int
}

// New creates a new rate limiter
func New(requestsPerMinute, burstSize int) *RateLimiter {
	return &RateLimiter{
		rpm:   requestsPerMinute,
		burst: burstSize,
	}
}

// getLimiter gets or creates a limiter for a specific client key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	if limiter, exists := rl.limiters.Load(key); exists {
		return limiter.(*rate.Limiter)
	}

	// Create new limiter: convert RPM to per-second rate
	limiter := rate.NewLimiter(
		rate.Limit(float64(rl.rpm)/60.0),
		rl.burst,
	)
	rl.limiters.Store(key, limiter)
	return limiter
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	limiter := rl.getLimiter(key)
	return limiter.Allow()
}

// Middleware creates HTTP middleware for rate limiting
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Determine client key: prefer API key, fallback to IP
		clientKey := r.Header.Get("X-API-Key")
		if clientKey == "" {
			if auth := r.Header.Get("Authorization"); auth != "" {
				clientKey = auth
			}
		}
		if clientKey == "" {
			clientKey = r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientKey = forwarded
			}
		}

		if !rl.Allow(clientKey) {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
