package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// MaxVisitors bounds memory when attackers spoof X-Forwarded-For.
	MaxVisitors = 10000
	// purgeEvery requests between opportunistic purges of expired entries.
	// No background goroutine: cleanup piggybacks on traffic (leak-free,
	// no Stop needed, safe for tests that build many routers).
	purgeEvery = 1000
)

type visitor struct {
	count int
	reset time.Time
}

// RateLimiter is a fixed-window limiter with anti-spoofing controls.
type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitor
	limit      int
	window     time.Duration
	trustProxy bool
	calls      uint64
}

// NewRateLimiterWithOptions creates a limiter. Set trustProxy=true only
// behind a trusted reverse proxy that sanitizes X-Forwarded-For; otherwise
// the client IP comes from RemoteAddr (spoof-proof).
func NewRateLimiterWithOptions(limit int, window time.Duration, trustProxy bool) *RateLimiter {
	return &RateLimiter{
		visitors:   make(map[string]*visitor),
		limit:      limit,
		window:     window,
		trustProxy: trustProxy,
	}
}

// purgeLocked drops expired entries. Caller must hold mu.
func (rl *RateLimiter) purgeLocked(now time.Time) {
	for ip, v := range rl.visitors {
		if now.After(v.reset) {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.clientIP(r)
		rl.mu.Lock()
		now := time.Now()
		rl.calls++
		if rl.calls%purgeEvery == 0 {
			rl.purgeLocked(now)
		}
		v, ok := rl.visitors[ip]
		if !ok || now.After(v.reset) {
			if !ok && len(rl.visitors) >= MaxVisitors {
				rl.purgeLocked(now)
				if len(rl.visitors) >= MaxVisitors {
					rl.mu.Unlock()
					w.Header().Set("Retry-After", "60")
					http.Error(w, "Too many requests", http.StatusTooManyRequests)
					return
				}
			}
			rl.visitors[ip] = &visitor{count: 1, reset: now.Add(rl.window)}
			rl.mu.Unlock()
			next.ServeHTTP(w, r)
			return
		}
		v.count++
		count := v.count
		reset := v.reset
		rl.mu.Unlock()
		if count > rl.limit {
			// Retry-After must be integer seconds (RFC 9111), not "42s".
			w.Header().Set("Retry-After", strconv.Itoa(int(time.Until(reset).Seconds())+1))
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) clientIP(r *http.Request) string {
	if rl.trustProxy {
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			// Take the leftmost (original client), trimmed and validated.
			first := strings.TrimSpace(strings.Split(xff, ",")[0])
			if first != "" {
				return first
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
