package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppErrorMessage(t *testing.T) {
	err := Validation("campo obrigatorio")
	if err.Error() != "campo obrigatorio" {
		t.Fatalf("expected %q, got %q", "campo obrigatorio", err.Error())
	}
}

func TestAppErrorWithWrapped(t *testing.T) {
	inner := errors.New("db timeout")
	err := Internal("falha interna", inner)
	if err.Error() != "falha interna" {
		t.Fatalf("unexpected message: %q", err.Error())
	}
	if !errors.Is(err, inner) {
		t.Fatal("expected Unwrap to return inner error")
	}
}

func TestAppErrorCodes(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		code int
	}{
		{"NotFound", NotFound("not found"), http.StatusNotFound},
		{"Validation", Validation("bad input"), http.StatusBadRequest},
		{"Conflict", Conflict("duplicate"), http.StatusConflict},
		{"Unauthorized", Unauthorized("no token"), http.StatusUnauthorized},
		{"Forbidden", Forbidden("no access"), http.StatusForbidden},
		{"Internal", Internal("server error", nil), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Fatalf("expected code %d, got %d", tt.code, tt.err.Code)
			}
		})
	}
}

func TestWrapIfNotApp(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		if err := WrapIfNotApp("msg", nil); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})
	t.Run("AppError passes through", func(t *testing.T) {
		original := NotFound("not found")
		result := WrapIfNotApp("wrapped", original)
		ae, ok := IsAppError(result)
		if !ok {
			t.Fatal("expected AppError")
		}
		if ae.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", ae.Code)
		}
		if ae.Message != "not found" {
			t.Fatalf("expected original message, got %q", ae.Message)
		}
	})
	t.Run("plain error wraps as Internal", func(t *testing.T) {
		plain := errors.New("db timeout")
		result := WrapIfNotApp("operation failed", plain)
		ae, ok := IsAppError(result)
		if !ok {
			t.Fatal("expected AppError")
		}
		if ae.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", ae.Code)
		}
		if ae.Message != "operation failed" {
			t.Fatalf("expected wrapped message, got %q", ae.Message)
		}
		if !errors.Is(result, plain) {
			t.Fatal("expected original error in chain")
		}
	})
}

func TestIsAppError(t *testing.T) {
	ae := Validation("test")
	if got, ok := IsAppError(ae); !ok || got != ae {
		t.Fatal("expected IsAppError to return true for AppError")
	}

	plain := errors.New("plain error")
	if _, ok := IsAppError(plain); ok {
		t.Fatal("expected IsAppError to return false for plain error")
	}
}
