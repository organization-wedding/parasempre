package guest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestHandler() (*Handler, *mockRepository) {
	repo := &mockRepository{}
	svc := NewService(repo)
	return NewHandler(svc), repo
}

func TestHandlerListGuests(t *testing.T) {
	h, repo := newTestHandler()
	repo.listFn = func(ctx context.Context) ([]Guest, error) {
		return []Guest{sampleGuest()}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/guests", nil)
	w := httptest.NewRecorder()
	h.handleList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var guests []Guest
	if err := json.NewDecoder(w.Body).Decode(&guests); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(guests) != 1 {
		t.Fatalf("expected 1 guest, got %d", len(guests))
	}
}

func TestHandlerListGuestsError(t *testing.T) {
	h, repo := newTestHandler()
	repo.listFn = func(ctx context.Context) ([]Guest, error) {
		return nil, errors.New("db error")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/guests", nil)
	w := httptest.NewRecorder()
	h.handleList(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandlerGetGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByID = func(ctx context.Context, id string) (*Guest, error) {
		g := sampleGuest()
		return &g, nil
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/guests/abc-123", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlerGetGuestNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByID = func(ctx context.Context, id string) (*Guest, error) {
		return nil, ErrNotFound
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/guests/not-exist", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlerCreateGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.createFn = func(ctx context.Context, input CreateGuestInput) (*Guest, error) {
		return &Guest{
			ID:             "new-id",
			Nome:           input.Nome,
			Sobrenome:      input.Sobrenome,
			Telefone:       input.Telefone,
			Relacionamento: input.Relacionamento,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}, nil
	}

	body := `{"nome":"Maria","sobrenome":"Santos","telefone":"11888888888","relacionamento":"noiva"}`
	req := httptest.NewRequest(http.MethodPost, "/api/guests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCreateGuestValidationError(t *testing.T) {
	h, _ := newTestHandler()

	body := `{"nome":"","sobrenome":"Santos","telefone":"11888888888","relacionamento":"noiva"}`
	req := httptest.NewRequest(http.MethodPost, "/api/guests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.handleCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerUpdateGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.updateFn = func(ctx context.Context, id string, input UpdateGuestInput) (*Guest, error) {
		g := sampleGuest()
		return &g, nil
	}

	body := `{"confirmacao":true}`
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPut, "/api/guests/abc-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerDeleteGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.deleteFn = func(ctx context.Context, id string) error {
		return nil
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/guests/abc-123", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestHandlerDeleteGuestNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.deleteFn = func(ctx context.Context, id string) error {
		return ErrNotFound
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodDelete, "/api/guests/not-exist", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlerImportCSV(t *testing.T) {
	h, repo := newTestHandler()
	var created int
	repo.createFn = func(ctx context.Context, input CreateGuestInput) (*Guest, error) {
		created++
		g := sampleGuest()
		return &g, nil
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "guests.csv")
	part.Write([]byte("nome,sobrenome,telefone,relacionamento\nJoão,Silva,11999999999,noivo\nMaria,Santos,11888888888,noiva\n"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/guests/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	h.handleImport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if created != 2 {
		t.Fatalf("expected 2 guests created, got %d", created)
	}
}

func TestHandlerImportNoFile(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/guests/import", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	w := httptest.NewRecorder()
	h.handleImport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
