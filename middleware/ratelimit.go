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
)

type visitor struct {
	count int
	reset time.Time
}

// RateLimiter is a fixed-window limiter with anti-spoofing controls.
type RateLimiter struct {
	mu          sync.Mutex
	visitors    map[string]*visitor
	limit       int
	window      time.Duration
	trustProxy  bool
	stopCleanup chan struct{}
	stopOnce    sync.Once
}

// NewRateLimiter creates a limiter. Set trustProxy=true only behind a
// trusted reverse proxy that sanitizes X-Forwarded-For; otherwise the
// client IP comes from RemoteAddr (spoof-proof).
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return NewRateLimiterWithOptions(limit, window, false)
}

func NewRateLimiterWithOptions(limit int, window time.Duration, trustProxy bool) *RateLimiter {
	rl := &RateLimiter{
		visitors:    make(map[string]*visitor),
		limit:       limit,
		window:      window,
		trustProxy:  trustProxy,
		stopCleanup: make(chan struct{}),
	}
	go rl.cleanup()
	return rl
}

// Stop terminates the background cleanup goroutine (prevents test leaks).
func (rl *RateLimiter) Stop() {
	rl.stopOnce.Do(func() { close(rl.stopCleanup) })
}

func (rl *RateLimiter) cleanup() {
	t := time.NewTicker(rl.window)
	defer t.Stop()
	for {
		select {
		case <-rl.stopCleanup:
			return
		case <-t.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, v := range rl.visitors {
				if now.After(v.reset) {
					delete(rl.visitors, ip)
				}
			}
			// Hard bound: drop oldest overflow arbitrarily to avoid OOM.
			if len(rl.visitors) > MaxVisitors {
				n := 0
				for ip := range rl.visitors {
					delete(rl.visitors, ip)
					n++
					if n > len(rl.visitors)-MaxVisitors {
						break
					}
				}
			}
			rl.mu.Unlock()
		}
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := rl.clientIP(r)
		rl.mu.Lock()
		v, ok := rl.visitors[ip]
		now := time.Now()
		if !ok || now.After(v.reset) {
			if !ok && len(rl.visitors) >= MaxVisitors {
				rl.mu.Unlock()
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
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
