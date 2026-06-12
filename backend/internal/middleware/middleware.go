package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
)

var phoneSegmentRegex = regexp.MustCompile(`\d{10,11}`)

func maskPhonePath(path string) string {
	return phoneSegmentRegex.ReplaceAllString(path, "<phone>")
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("request",
			"method", r.Method,
			"path", maskPhonePath(r.URL.Path),
			"status", sw.status,
			"duration", time.Since(start),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
	wrote  bool
}

func (sw *statusWriter) WriteHeader(code int) {
	if !sw.wrote {
		sw.status = code
		sw.wrote = true
	}
	sw.ResponseWriter.WriteHeader(code)
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", fmt.Sprint(err),
					"stack", string(debug.Stack()),
				)
				httputil.WriteError(w, r, apperror.Internal("internal server error", fmt.Errorf("%v", err)))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders applies baseline hardening for an API server. CSP is
// intentionally absent here — this is an API, not HTML. The frontend server
// owns CSP for its own responses (see frontend/src/index.ts).
//
// HSTS only set in production: dev runs over plain HTTP and the header
// would brick local browsers for the API hostname.
func SecurityHeaders(appEnv string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("X-Frame-Options", "DENY")
			if appEnv == "production" {
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func CORS(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, user-racf")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
