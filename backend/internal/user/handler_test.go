package user

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/guest"
)

func newTestHandler() (*Handler, *mockUserRepo, *mockGuestRepo) {
	userRepo := &mockUserRepo{}
	guestRepo := &mockGuestRepo{}
	svc := NewService(userRepo, guestRepo)
	return NewHandler(svc), userRepo, guestRepo
}

func TestHandlerRegisterSuccess(t *testing.T) {
	h, userRepo, guestRepo := newTestHandler()
	guestRepo.getByPhone = func(ctx context.Context, phone string) (*guest.Guest, error) {
		return sampleGuest(), nil
	}
	userRepo.getByGuestID = func(ctx context.Context, guestID int64) (*User, error) {
		return nil, nil
	}
	userRepo.getByURACF = func(ctx context.Context, uracf string) (*User, error) {
		return nil, nil
	}
	userRepo.createFn = func(ctx context.Context, u *User) (*User, error) {
		return &User{
			ID:        1,
			GuestID:   u.GuestID,
			Role:      u.Role,
			URACF:     u.URACF,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}

	body := `{"phone":"11999999999","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRegister(w, req)

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
	h, userRepo, guestRepo := newTestHandler()
	guestRepo.getByPhone = func(ctx context.Context, phone string) (*guest.Guest, error) {
		return sampleGuest(), nil
	}
	userRepo.getByGuestID = func(ctx context.Context, guestID int64) (*User, error) {
		return sampleUser(), nil
	}

	body := `{"phone":"11999999999","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRegister(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterGuestNotFound(t *testing.T) {
	h, _, guestRepo := newTestHandler()
	guestRepo.getByPhone = func(ctx context.Context, phone string) (*guest.Guest, error) {
		return nil, nil
	}

	body := `{"phone":"11999999999","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRegister(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterInvalidJSON(t *testing.T) {
	h, _, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterInvalidPhone(t *testing.T) {
	h, _, _ := newTestHandler()

	body := `{"phone":"abc","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerRegisterURACFTaken(t *testing.T) {
	h, userRepo, guestRepo := newTestHandler()
	guestRepo.getByPhone = func(ctx context.Context, phone string) (*guest.Guest, error) {
		return sampleGuest(), nil
	}
	userRepo.getByGuestID = func(ctx context.Context, guestID int64) (*User, error) {
		return nil, nil
	}
	userRepo.getByURACF = func(ctx context.Context, uracf string) (*User, error) {
		return sampleUser(), nil
	}

	body := `{"phone":"11999999999","uracf":"USR01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleRegister(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCheckExists(t *testing.T) {
	h, userRepo, guestRepo := newTestHandler()
	guestRepo.getByPhone = func(ctx context.Context, phone string) (*guest.Guest, error) {
		return sampleGuest(), nil
	}
	userRepo.getByGuestID = func(ctx context.Context, guestID int64) (*User, error) {
		return sampleUser(), nil
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

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
	if resp.Role != "guest" {
		t.Fatalf("expected role=guest, got %q", resp.Role)
	}
}

func TestHandlerCheckNotExists(t *testing.T) {
	h, _, guestRepo := newTestHandler()
	guestRepo.getByPhone = func(ctx context.Context, phone string) (*guest.Guest, error) {
		return nil, nil
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

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
	h, _, _ := newTestHandler()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/users/check", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCheckInvalidPhone(t *testing.T) {
	h, _, _ := newTestHandler()

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/users/check?phone=invalid", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
