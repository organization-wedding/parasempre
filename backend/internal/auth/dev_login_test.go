package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func newTestDevLoginHandler(finder UserFinder) *DevLoginHandler {
	jwtSvc := NewJWTService("test-secret", 1*time.Hour)
	loginRecorder := &mockLoginRecorder{}
	return NewDevLoginHandler(jwtSvc, finder, loginRecorder, "11999999999")
}

func TestDevLoginHappyPath(t *testing.T) {
	finder := &mockUserFinder{
		findFn: func(ctx context.Context, phone string) (int64, string, string, error) {
			if phone != "11999999999" {
				t.Fatalf("expected phone 11999999999, got %q", phone)
			}
			return 42, "USR99", "groom", nil
		},
	}
	h := newTestDevLoginHandler(finder)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/dev-login", nil)
	w := httptest.NewRecorder()
	h.Handle(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp TokenResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.URACF != "USR99" {
		t.Fatalf("expected URACF USR99, got %q", resp.URACF)
	}
	if resp.Role != "groom" {
		t.Fatalf("expected role groom, got %q", resp.Role)
	}
}

func TestDevLoginPhoneNotSeeded(t *testing.T) {
	finder := &mockUserFinder{
		findFn: func(ctx context.Context, phone string) (int64, string, string, error) {
			return 0, "", "", apperror.NotFound("no user found with this phone", errors.New("not found"))
		},
	}
	h := newTestDevLoginHandler(finder)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/dev-login", nil)
	w := httptest.NewRecorder()
	h.Handle(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
