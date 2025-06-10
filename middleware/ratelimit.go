package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple IP-based request counter with sliding window
type RateLimiter struct {
	mutex    sync.Mutex
	requests map[string][]time.Time
	limit    int           // Maximum number of requests
	window   time.Duration // Time window for counting
}

// NewRateLimiter creates a new rate limiter with specific limits
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Limit is a middleware that limits requests by IP
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the client's IP (or you can use a custom identifier)
		ip := r.RemoteAddr

		// Check if the IP is within the rate limits
		if !rl.isAllowed(ip) {
			w.Header().Set("Content-Type", "application/json")			w.Header().Set("Retry-After", "60")       // Suggest trying again after 60 seconds
			w.WriteHeader(http.StatusTooManyRequests) // 429 Too Many Requests
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Rate limit exceeded. Please try again later.",
			})
			return
		}

		// If within limits, process the request
		next.ServeHTTP(w, r)
	})
}

// isAllowed checks if the IP can make more requests
func (rl *RateLimiter) isAllowed(ip string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Remove old timestamps outside the time window
	if timestamps, exists := rl.requests[ip]; exists {
		var validTimestamps []time.Time
		for _, ts := range timestamps {
			if ts.After(windowStart) {
				validTimestamps = append(validTimestamps, ts)
			}
		}
		rl.requests[ip] = validTimestamps
		// Check if the limit has been reached
		if len(validTimestamps) >= rl.limit {
			return false
		}
	}

	// Add the new request to the counter
	rl.requests[ip] = append(rl.requests[ip], now)
	return true
}
