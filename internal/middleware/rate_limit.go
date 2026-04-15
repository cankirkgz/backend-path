package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientRecord struct {
	count     int
	resetTime time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]*clientRecord
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]*clientRecord),
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)

		if !rl.allow(clientIP) {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte("rate limit exceeded"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	record, exists := rl.clients[clientIP]
	if !exists || now.After(record.resetTime) {
		rl.clients[clientIP] = &clientRecord{
			count:     1,
			resetTime: now.Add(rl.window),
		}
		return true
	}

	if record.count >= rl.limit {
		return false
	}

	record.count++
	return true
}

func getClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
