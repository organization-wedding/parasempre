package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func TestRateLimiterAllows(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(10), 10)
	defer rl.Close()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRateLimiterBlocks(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(1), 1)
	defer rl.Close()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.RemoteAddr = "1.2.3.4:12345"

	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request: expected 200, got %d", w1.Code)
	}

	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request: expected 429, got %d", w2.Code)
	}
}

func TestRateLimiterPerIP(t *testing.T) {
	rl := NewRateLimiter(rate.Limit(1), 1)
	defer rl.Close()

	handler := rl.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req1 := httptest.NewRequest(http.MethodPost, "/", nil)
	req1.RemoteAddr = "1.2.3.4:12345"
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("IP1 first: expected 200, got %d", w1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/", nil)
	req2.RemoteAddr = "5.6.7.8:12345"
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("IP2 first: expected 200, got %d (different IP should not be limited)", w2.Code)
	}
}
