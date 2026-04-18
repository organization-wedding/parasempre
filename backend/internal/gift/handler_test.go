package gift

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
)

func newTestHandler() (*Handler, *mockRepository) {
	repo := &mockRepository{}
	svc := NewService(repo)
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

func registerTestRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /api/gifts", h.HandleList)
	mux.HandleFunc("POST /api/gifts", h.HandleCreate)
	mux.HandleFunc("GET /api/gifts/{id}", h.HandleGet)
	mux.HandleFunc("PUT /api/gifts/{id}", h.HandleUpdate)
	mux.HandleFunc("DELETE /api/gifts/{id}", h.HandleDelete)
}

func TestHandlerListGifts(t *testing.T) {
	h, repo := newTestHandler()
	var gotStatus *string
	repo.listFn = func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
		gotStatus = statusFilter
		return []Gift{sampleGift()}, 1, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/gifts?page=1&limit=20", nil)
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result PublicPagedResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 gift, got %d", len(result.Data))
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
	if gotStatus == nil || *gotStatus != "active" {
		t.Fatalf("expected forced 'active' filter on public list, got %v", gotStatus)
	}
}

func TestHandlerListGiftsDoesNotLeakInternalFields(t *testing.T) {
	h, repo := newTestHandler()
	repo.listFn = func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
		return []Gift{sampleGift()}, 1, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/gifts", nil)
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	body := w.Body.String()
	for _, forbidden := range []string{"created_by", "updated_by", "dedupe_key"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("public list response leaks internal field %q: %s", forbidden, body)
		}
	}
}

func TestHandlerGetGiftDoesNotLeakInternalFields(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDFn = func(ctx context.Context, id int64) (*Gift, error) {
		g := sampleGift()
		return &g, nil
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/gifts/1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	body := w.Body.String()
	for _, forbidden := range []string{"created_by", "updated_by", "dedupe_key"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("public get response leaks internal field %q: %s", forbidden, body)
		}
	}
}

func TestHandlerListGiftsIgnoresUserProvidedStatus(t *testing.T) {
	h, repo := newTestHandler()
	var gotStatus *string
	repo.listFn = func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
		gotStatus = statusFilter
		return []Gift{}, 0, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/gifts?status=inactive", nil)
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	if gotStatus == nil || *gotStatus != "active" {
		t.Fatalf("expected forced 'active' filter even when client requests 'inactive', got %v", gotStatus)
	}
}

func TestHandlerGetGift(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDFn = func(ctx context.Context, id int64) (*Gift, error) {
		g := sampleGift()
		return &g, nil
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/gifts/1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlerGetGiftInactiveReturns404(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDFn = func(ctx context.Context, id int64) (*Gift, error) {
		g := sampleGift()
		g.Status = "inactive"
		return &g, nil
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/gifts/1", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for inactive gift, got %d", w.Code)
	}
}

func TestHandlerGetGiftNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDFn = func(ctx context.Context, id int64) (*Gift, error) {
		return nil, apperror.NotFound("gift not found")
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/gifts/999", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlerGetGiftInvalidID(t *testing.T) {
	h, _ := newTestHandler()

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/gifts/abc", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlerCreateGift(t *testing.T) {
	h, repo := newTestHandler()
	repo.createFn = func(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error) {
		g := sampleGift()
		g.Name = input.Name
		g.PriceCents = input.PriceCents
		g.DedupeKey = dedupeKey
		g.CreatedBy = userRACF
		g.UpdatedBy = userRACF
		return &g, nil
	}

	body := `{"name":"Panela Inox","price_cents":19990}`
	req := httptest.NewRequest(http.MethodPost, "/api/gifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCreateGiftInvalidJSON(t *testing.T) {
	h, _ := newTestHandler()

	body := `{"name":`
	req := httptest.NewRequest(http.MethodPost, "/api/gifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCreateGiftValidationError(t *testing.T) {
	h, _ := newTestHandler()

	body := `{"name":"","price_cents":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/gifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCreateGiftOversizedBody(t *testing.T) {
	h, _ := newTestHandler()

	huge := strings.Repeat("a", (1<<20)+1024)
	body := `{"name":"Panela","price_cents":19990,"description":"` + huge + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/gifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when body exceeds limit, got %d", w.Code)
	}
}

func TestHandlerCreateGiftConflict(t *testing.T) {
	h, repo := newTestHandler()
	repo.createFn = func(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error) {
		return nil, apperror.Conflict("a gift with this name already exists")
	}

	body := `{"name":"Panela","price_cents":19990}`
	req := httptest.NewRequest(http.MethodPost, "/api/gifts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleCreate(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestHandlerUpdateGift(t *testing.T) {
	h, repo := newTestHandler()
	repo.updateFn = func(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error) {
		g := sampleGift()
		return &g, nil
	}

	body := `{"price_cents":25000}`
	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodPut, "/api/gifts/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerUpdateGiftNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.updateFn = func(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error) {
		return nil, apperror.NotFound("gift not found")
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	body := `{"price_cents":25000}`
	req := httptest.NewRequest(http.MethodPut, "/api/gifts/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlerDeleteGift(t *testing.T) {
	h, repo := newTestHandler()
	repo.deleteFn = func(ctx context.Context, id int64, userRACF string) error {
		return nil
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodDelete, "/api/gifts/1", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestHandlerDeleteGiftNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.deleteFn = func(ctx context.Context, id int64, userRACF string) error {
		return apperror.NotFound("gift not found")
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodDelete, "/api/gifts/999", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
