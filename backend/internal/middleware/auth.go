package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
)

type contextKey string

const claimsKey contextKey = "claims"

func RequireAuth(jwtSvc *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				httputil.WriteError(w, http.StatusUnauthorized, "authorization header required")
				return
			}

			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if tokenStr == header {
				httputil.WriteError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			claims, err := jwtSvc.Parse(tokenStr)
			if err != nil {
				httputil.HandleError(w, err)
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				httputil.WriteError(w, http.StatusUnauthorized, "authentication required")
				return
			}

			if !allowed[claims.Role] {
				httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func DevOnly(appEnv string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if appEnv == "production" {
				httputil.WriteError(w, http.StatusForbidden, "this endpoint is not available in production")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func ClaimsFromContext(ctx context.Context) *auth.Claims {
	claims, _ := ctx.Value(claimsKey).(*auth.Claims)
	return claims
}

func WithClaims(ctx context.Context, claims *auth.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

func UserRACFFromContext(ctx context.Context) string {
	claims := ClaimsFromContext(ctx)
	if claims == nil {
		return ""
	}
	return claims.URACF
}
