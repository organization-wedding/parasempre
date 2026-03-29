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

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
)

func registerTestRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /api/guests", h.HandleList)
	mux.HandleFunc("POST /api/guests", h.HandleCreate)
	mux.HandleFunc("GET /api/guests/{id}", h.HandleGet)
	mux.HandleFunc("PUT /api/guests/{id}", h.HandleUpdate)
	mux.HandleFunc("DELETE /api/guests/{id}", h.HandleDelete)
	mux.HandleFunc("POST /api/guests/import", h.HandleImport)
}

func newTestHandler() (*Handler, *mockRepository) {
	repo := &mockRepository{}
	svc := newTestService(repo, alwaysExistsChecker(), noopUserCreator())
	return NewHandler(svc), repo
}

func withTestClaims(req *http.Request, uracf string) *http.Request {
	claims := &auth.Claims{
		UserID: 1,
		URACF:  uracf,
		Role:   "groom",
	}
	ctx := middleware.WithClaims(req.Context(), claims)
	return req.WithContext(ctx)
}

func TestHandlerListGuests(t *testing.T) {
	h, repo := newTestHandler()
	repo.listFn = func(ctx context.Context) ([]Guest, error) {
		return []Guest{sampleGuest()}, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/guests", nil)
	w := httptest.NewRecorder()
	h.HandleList(w, req)

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
	h.HandleList(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandlerGetGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByID = func(ctx context.Context, id int64) (*Guest, error) {
		g := sampleGuest()
		return &g, nil
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/guests/1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlerGetGuestNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByID = func(ctx context.Context, id int64) (*Guest, error) {
		return nil, apperror.NotFound("guest not found")
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/guests/999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlerGetGuestInvalidID(t *testing.T) {
	h, _ := newTestHandler()

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/guests/abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlerCreateGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.createFn = func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
		return &Guest{
			ID:           1,
			FirstName:    input.FirstName,
			LastName:     input.LastName,
			Relationship: input.Relationship,
			FamilyGroup:  *input.FamilyGroup,
			CreatedBy:    userRACF,
			UpdatedBy:    userRACF,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}, nil
	}

	body := `{"first_name":"Maria","last_name":"Santos","relationship":"R","family_group":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/guests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCreateGuestValidationError(t *testing.T) {
	h, _ := newTestHandler()

	body := `{"first_name":"","last_name":"Santos","relationship":"R","family_group":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/guests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCreateGuestMissingFamilyGroup(t *testing.T) {
	h, repo := newTestHandler()
	repo.createFn = func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
		g := sampleGuest()
		if input.FamilyGroup == nil {
			t.Fatal("expected auto-assigned family_group, got nil")
		}
		g.FamilyGroup = *input.FamilyGroup
		return &g, nil
	}

	body := `{"first_name":"Maria","last_name":"Santos","relationship":"R"}`
	req := httptest.NewRequest(http.MethodPost, "/api/guests", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerUpdateGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.updateFn = func(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
		g := sampleGuest()
		return &g, nil
	}

	body := `{"confirmed":true}`
	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodPut, "/api/guests/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerDeleteGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.deleteFn = func(ctx context.Context, id int64) error {
		return nil
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodDelete, "/api/guests/1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestHandlerDeleteGuestNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.deleteFn = func(ctx context.Context, id int64) error {
		return apperror.NotFound("guest not found")
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodDelete, "/api/guests/999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlerImportCSV(t *testing.T) {
	h, repo := newTestHandler()
	var created int
	repo.createFn = func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
		created++
		g := sampleGuest()
		return &g, nil
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "guests.csv")
	part.Write([]byte("first_name,last_name,relationship,family_group\nJoão,Silva,P,1\nMaria,Santos,R,2\n"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/guests/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleImport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if created != 2 {
		t.Fatalf("expected 2 guests created, got %d", created)
	}
}

func TestHandlerImportCSVWithErrors(t *testing.T) {
	h, repo := newTestHandler()
	repo.createFn = func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
		return nil, errors.New("duplicate name")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, _ := writer.CreateFormFile("file", "guests.csv")
	part.Write([]byte("first_name,last_name,relationship,family_group\nJoão,Silva,P,1\n"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/guests/import", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleImport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerImportNoFile(t *testing.T) {
	h, _ := newTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/guests/import", nil)
	req.Header.Set("Content-Type", "multipart/form-data")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleImport(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
