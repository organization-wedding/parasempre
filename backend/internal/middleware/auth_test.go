package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/auth"
)

func newTestJWT() *auth.JWTService {
	return auth.NewJWTService("test-secret-for-middleware", 1*time.Hour)
}

func TestRequireAuthValid(t *testing.T) {
	jwtSvc := newTestJWT()
	token, _ := jwtSvc.Generate(1, "USR01", "guest")

	handler := RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil {
			t.Fatal("expected claims in context")
		}
		if claims.URACF != "USR01" {
			t.Fatalf("expected URACF USR01, got %q", claims.URACF)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireAuthMissing(t *testing.T) {
	jwtSvc := newTestJWT()
	handler := RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthInvalidFormat(t *testing.T) {
	jwtSvc := newTestJWT()
	handler := RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	jwtSvc := newTestJWT()
	handler := RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireRoleAllowed(t *testing.T) {
	jwtSvc := newTestJWT()
	token, _ := jwtSvc.Generate(1, "GRM01", "groom")

	handler := RequireAuth(jwtSvc)(RequireRole("groom", "bride")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireRoleForbidden(t *testing.T) {
	jwtSvc := newTestJWT()
	token, _ := jwtSvc.Generate(1, "USR01", "guest")

	handler := RequireAuth(jwtSvc)(RequireRole("groom", "bride")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	})))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestDevOnlyBlocked(t *testing.T) {
	handler := DevOnly("production")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach handler")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}

	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["error"] != "this endpoint is not available in production" {
		t.Fatalf("unexpected error: %q", body["error"])
	}
}

func TestDevOnlyAllowed(t *testing.T) {
	handler := DevOnly("test")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserRACFFromContext(t *testing.T) {
	jwtSvc := newTestJWT()
	token, _ := jwtSvc.Generate(1, "TST01", "guest")

	handler := RequireAuth(jwtSvc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uracf := UserRACFFromContext(r.Context())
		if uracf != "TST01" {
			t.Fatalf("expected TST01, got %q", uracf)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUserRACFFromContextEmpty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	uracf := UserRACFFromContext(req.Context())
	if uracf != "" {
		t.Fatalf("expected empty, got %q", uracf)
	}
}
