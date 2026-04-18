package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func registerUserTestRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("POST /api/users", h.HandleRegister)
	mux.HandleFunc("GET /api/users/check", h.HandleCheck)
	mux.HandleFunc("GET /api/users", h.HandleList)
	mux.HandleFunc("GET /api/users/me", h.HandleMe)
}

func newTestHandler() (*Handler, *mockUserRepo) {
	userRepo := &mockUserRepo{}
	guestRepo := &mockGuestRepo{}
	svc := NewService(userRepo, guestRepo)
	return NewHandler(svc, "test"), userRepo
}

func TestHandlerRegisterSuccess(t *testing.T) {
	h, userRepo := newTestHandler()
	userRepo.getByPhone = func(ctx context.Context, phone string) (*User, error) {
		return nil, nil
	}
	userRepo.getByURACF = func(ctx context.Context, uracf string) (*User, error) {
		return nil, nil
	}
	userRepo.createFn = func(ctx context.Context, u *User) (*User, error) {
		return &User{
			ID:        1,
			Role:      u.Role,
			URACF:     u.URACF,
			Phone:     u.Phone,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	body := `{"phone":"11999999999","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var u User
	if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if u.ID == 0 {
		t.Fatal("expected user ID in response")
	}
}

func TestHandlerRegisterAlreadyRegistered(t *testing.T) {
	h, userRepo := newTestHandler()
	userRepo.getByPhone = func(ctx context.Context, phone string) (*User, error) {
		return sampleUser(), nil
	}

	body := `{"phone":"11999999999","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterInvalidJSON(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterInvalidPhone(t *testing.T) {
	h, _ := newTestHandler()

	body := `{"phone":"abc","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterURACFTaken(t *testing.T) {
	h, userRepo := newTestHandler()
	userRepo.getByPhone = func(ctx context.Context, phone string) (*User, error) {
		return nil, nil
	}
	userRepo.getByURACF = func(ctx context.Context, uracf string) (*User, error) {
		return sampleUser(), nil
	}

	body := `{"phone":"11999999999","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleRegister(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCheckExists(t *testing.T) {
	h, userRepo := newTestHandler()
	userRepo.getByPhone = func(ctx context.Context, phone string) (*User, error) {
		return sampleUser(), nil
	}

	mux := http.NewServeMux()
	registerUserTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/users/check?phone=11999999999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CheckResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !resp.Exists {
		t.Fatal("expected exists=true")
	}
}

func TestHandlerCheckNotExists(t *testing.T) {
	h, userRepo := newTestHandler()
	userRepo.getByPhone = func(ctx context.Context, phone string) (*User, error) {
		return nil, nil
	}

	mux := http.NewServeMux()
	registerUserTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/users/check?phone=11999999999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp CheckResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Exists {
		t.Fatal("expected exists=false")
	}
}

func TestHandlerCheckMissingPhone(t *testing.T) {
	h, _ := newTestHandler()

	mux := http.NewServeMux()
	registerUserTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/users/check", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCheckInvalidPhone(t *testing.T) {
	h, _ := newTestHandler()

	mux := http.NewServeMux()
	registerUserTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/users/check?phone=invalid", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
