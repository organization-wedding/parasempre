package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"golang.org/x/time/rate"

	"github.com/ferjunior7/parasempre/backend/internal/httputil"
)

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     rate.Limit
	burst    int
	stopOnce sync.Once
	stop     chan struct{}
}

// KeyFunc extracts the rate-limit bucket key from a request. Default is IP
// (see IPKey); the purchase route uses a user-id key so guests on the same
// network (Wi-Fi de família/escritório) don't share a bucket.
type KeyFunc func(r *http.Request) string

func NewRateLimiter(r rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		rate:    r,
		burst:   burst,
		stop:    make(chan struct{}),
	}

	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.buckets[key]
	if !exists {
		v = &bucket{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.buckets[key] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func IPKey(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			for key, v := range rl.buckets {
				if time.Since(v.lastSeen) > 10*time.Minute {
					delete(rl.buckets, key)
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
	return rl.MiddlewareWithKey(IPKey)
}

func (rl *RateLimiter) MiddlewareWithKey(keyFn KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.getLimiter(keyFn(r)).Allow() {
				httputil.WriteError(w, r, apperror.TooManyRequests("too many requests, wait 1 minute and try again"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
