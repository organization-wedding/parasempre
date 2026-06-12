package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func newTestDevLoginHandler(finder UserFinder) *DevLoginHandler {
	jwtSvc := NewJWTService("test-secret", 1*time.Hour)
	loginRecorder := &mockLoginRecorder{}
	return NewDevLoginHandler(jwtSvc, finder, loginRecorder)
}

func TestDevLogin(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		finderFn  func(ctx context.Context, uracf string) (int64, string, string, error)
		wantCode  int
		wantURACF string
		wantRole  string
		wantToken bool
	}{
		{
			name:  "happy path returns 200 with token",
			query: "?uracf=USR99",
			finderFn: func(ctx context.Context, uracf string) (int64, string, string, error) {
				if uracf != "USR99" {
					t.Fatalf("expected uracf USR99, got %q", uracf)
				}
				return 42, "USR99", "groom", nil
			},
			wantCode:  http.StatusOK,
			wantURACF: "USR99",
			wantRole:  "groom",
			wantToken: true,
		},
		{
			name:  "lowercase uracf is normalized to uppercase before lookup",
			query: "?uracf=%20usr99%20",
			finderFn: func(ctx context.Context, uracf string) (int64, string, string, error) {
				if uracf != "USR99" {
					t.Fatalf("expected normalized uracf USR99, got %q", uracf)
				}
				return 42, "USR99", "groom", nil
			},
			wantCode:  http.StatusOK,
			wantURACF: "USR99",
			wantRole:  "groom",
			wantToken: true,
		},
		{
			name:     "empty uracf query param returns 400",
			query:    "?uracf=",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "whitespace-only uracf returns 400",
			query:    "?uracf=%20%20",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing uracf query param returns 400",
			query:    "",
			wantCode: http.StatusBadRequest,
		},
		{
			name:  "uracf not found returns 404",
			query: "?uracf=ZZZZZ",
			finderFn: func(ctx context.Context, uracf string) (int64, string, string, error) {
				return 0, "", "", apperror.NotFound("no user found with this URACF")
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := &mockUserFinder{
				findByURACFFn: tt.finderFn,
			}
			h := newTestDevLoginHandler(finder)

			req := httptest.NewRequest(http.MethodPost, "/api/auth/dev-login"+tt.query, nil)
			w := httptest.NewRecorder()
			h.Handle(w, req)

			if w.Code != tt.wantCode {
				t.Fatalf("expected status %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			if tt.wantCode != http.StatusOK {
				return
			}

			var resp TokenResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if tt.wantToken && resp.Token == "" {
				t.Fatal("expected non-empty token")
			}
			if resp.URACF != tt.wantURACF {
				t.Fatalf("expected URACF %q, got %q", tt.wantURACF, resp.URACF)
			}
			if resp.Role != tt.wantRole {
				t.Fatalf("expected role %q, got %q", tt.wantRole, resp.Role)
			}
		})
	}
}
