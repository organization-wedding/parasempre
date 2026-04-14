package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"github.com/ferjunior7/parasempre/backend/internal/apperror"

	"github.com/ferjunior7/parasempre/backend/internal/httputil"
)

type ipLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*ipLimiter
	rate     rate.Limit
	burst    int
	stopOnce sync.Once
	stop     chan struct{}
}

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ipLimiter),
		rate:    r,
		burst:   burst,
		stop:    make(chan struct{}),
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.clients[ip]
	if !exists {
		v = &ipLimiter{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.clients[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for ip, v := range rl.clients {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(rl.clients, ip)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

func (rl *RateLimiter) Close() {
	rl.stopOnce.Do(func() { close(rl.stop) })
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			if !rl.getLimiter(ip).Allow() {
				httputil.WriteError(w, r, apperror.TooManyRequests("too many requests, wait 1 minute and try again"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
