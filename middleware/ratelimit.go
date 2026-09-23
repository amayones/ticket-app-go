package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type visitor struct {
	count int
	reset time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(rl.window)
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.After(v.reset) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		rl.mu.Lock()
		v, ok := rl.visitors[ip]
		now := time.Now()
		if !ok || now.After(v.reset) {
			rl.visitors[ip] = &visitor{count: 1, reset: now.Add(rl.window)}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		count := v.count
		rl.mu.Unlock()
		if count > rl.limit {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
