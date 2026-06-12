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
	svc := newTestService(repo, defaultUserBridge())
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
	repo.listFn = func(ctx context.Context, limit, offset int) ([]Guest, int, error) {
		return []Guest{sampleGuest()}, 1, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/guests?page=1&limit=20", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result PagedResponse
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("expected 1 guest, got %d", len(result.Data))
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
}

func TestHandlerListGuestsError(t *testing.T) {
	h, repo := newTestHandler()
	repo.listFn = func(ctx context.Context, limit, offset int) ([]Guest, int, error) {
		return nil, 0, errors.New("db error")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/guests", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	h.HandleList(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandlerGetGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
		g := sampleGuest()
		return &g, nil
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/guests/1", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlerGetGuestNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
		return nil, apperror.NotFound("guest not found")
	}

	mux := http.NewServeMux()
	registerTestRoutes(mux, h)

	req := httptest.NewRequest(http.MethodGet, "/api/guests/999", nil)
	req = withTestClaims(req, "TST01")
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

	body := `{"attending":true}`
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

func TestHandlerConfirmGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
		g := sampleGuest()
		return &g, nil
	}
	repo.setAttendingFn = func(ctx context.Context, id int64, attending bool, userRACF string) (*Guest, error) {
		g := sampleGuest()
		g.Attending = boolPtr(true)
		return &g, nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/guests/{id}/confirm", h.HandleConfirm)

	req := httptest.NewRequest(http.MethodPatch, "/api/guests/1/confirm", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var guest Guest
	if err := json.NewDecoder(w.Body).Decode(&guest); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if guest.Attending == nil || !*guest.Attending {
		t.Fatal("expected attending to be true")
	}
}

func TestHandlerConfirmGuestNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
		return nil, apperror.NotFound("guest not found")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/guests/{id}/confirm", h.HandleConfirm)

	req := httptest.NewRequest(http.MethodPatch, "/api/guests/999/confirm", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandlerCancelGuest(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
		g := sampleGuest()
		g.Attending = boolPtr(true)
		return &g, nil
	}
	repo.setAttendingFn = func(ctx context.Context, id int64, attending bool, userRACF string) (*Guest, error) {
		g := sampleGuest()
		g.Attending = boolPtr(false)
		return &g, nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/guests/{id}/cancel", h.HandleCancel)

	req := httptest.NewRequest(http.MethodPatch, "/api/guests/1/cancel", nil)
	req = withTestClaims(req, "TST01")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlerCancelGuestNotFound(t *testing.T) {
	h, repo := newTestHandler()
	repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
		return nil, apperror.NotFound("guest not found")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("PATCH /api/guests/{id}/cancel", h.HandleCancel)

	req := httptest.NewRequest(http.MethodPatch, "/api/guests/999/cancel", nil)
	req = withTestClaims(req, "TST01")
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

	var resp ImportResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.SuccessCount != 2 || resp.ErrorCount != 0 {
		t.Fatalf("expected 2 success, 0 errors, got %+v", resp)
	}
}

func TestHandlerImportCSVAllErrors(t *testing.T) {
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

	var resp ImportResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.ErrorCount != 1 || resp.Errors[0].Row != 2 {
		t.Fatalf("expected 1 error on row 2, got %+v", resp)
	}
}

func TestHandlerImportCSVPartialSuccess(t *testing.T) {
	h, repo := newTestHandler()
	callCount := 0
	repo.createFn = func(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
		callCount++
		if callCount == 2 {
			return nil, errors.New("duplicate name")
		}
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

	if w.Code != http.StatusMultiStatus {
		t.Fatalf("expected 207, got %d: %s", w.Code, w.Body.String())
	}

	var resp ImportResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.SuccessCount != 1 || resp.ErrorCount != 1 {
		t.Fatalf("expected 1 success + 1 error, got %+v", resp)
	}
}

func TestHandlerPhonePathValidation(t *testing.T) {
	tests := []struct {
		name       string
		phone      string
		wantStatus int
		callsSvc   bool
	}{
		{"valid phone", "11999990000", http.StatusOK, true},
		{"letters only", "abc", http.StatusBadRequest, false},
		{"too short", "123", http.StatusBadRequest, false},
		{"too long", "119999900001", http.StatusBadRequest, false},
		{"missing ninth digit", "11899990000", http.StatusBadRequest, false},
		{"empty", "", http.StatusTemporaryRedirect, false},
	}

	for _, tc := range tests {
		t.Run("confirm_by_phone/"+tc.name, func(t *testing.T) {
			h, repo := newTestHandler()
			svcCalled := false
			guestID := int64(1)
			repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			}
			repo.setAttendingFn = func(ctx context.Context, id int64, attending bool, userRACF string) (*Guest, error) {
				svcCalled = true
				g := sampleGuest()
				g.Attending = boolPtr(true)
				return &g, nil
			}
			h.svc.users.(*mockUserBridge).getGuestIDByPhoneFn = func(ctx context.Context, phone string) (*int64, error) {
				return &guestID, nil
			}

			mux := http.NewServeMux()
			mux.HandleFunc("PATCH /api/guests/phone/{phone}/confirm", h.HandleConfirmByPhone)

			req := httptest.NewRequest(http.MethodPatch, "/api/guests/phone/"+tc.phone+"/confirm", nil)
			req = withTestClaims(req, "TST01")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tc.wantStatus, w.Code, w.Body.String())
			}
			if !tc.callsSvc && svcCalled {
				t.Fatal("service should not have been called for invalid phone")
			}
		})

		t.Run("cancel_by_phone/"+tc.name, func(t *testing.T) {
			h, repo := newTestHandler()
			svcCalled := false
			guestID := int64(1)
			repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			}
			repo.setAttendingFn = func(ctx context.Context, id int64, attending bool, userRACF string) (*Guest, error) {
				svcCalled = true
				g := sampleGuest()
				g.Attending = boolPtr(false)
				return &g, nil
			}
			h.svc.users.(*mockUserBridge).getGuestIDByPhoneFn = func(ctx context.Context, phone string) (*int64, error) {
				return &guestID, nil
			}

			mux := http.NewServeMux()
			mux.HandleFunc("PATCH /api/guests/phone/{phone}/cancel", h.HandleCancelByPhone)

			req := httptest.NewRequest(http.MethodPatch, "/api/guests/phone/"+tc.phone+"/cancel", nil)
			req = withTestClaims(req, "TST01")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tc.wantStatus, w.Code, w.Body.String())
			}
			if !tc.callsSvc && svcCalled {
				t.Fatal("service should not have been called for invalid phone")
			}
		})

		t.Run("confirm_family_by_phone/"+tc.name, func(t *testing.T) {
			h, repo := newTestHandler()
			svcCalled := false
			repo.getFamilyGroupByPhoneFn = func(ctx context.Context, phone string) (*int64, error) {
				svcCalled = true
				fg := int64(1)
				return &fg, nil
			}
			repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			}
			repo.setAttendingByFamilyGroupFn = func(ctx context.Context, familyGroup int64, attending bool, userRACF string) ([]Guest, error) {
				g := sampleGuest()
				g.Attending = boolPtr(true)
				return []Guest{g}, nil
			}

			mux := http.NewServeMux()
			mux.HandleFunc("PATCH /api/guests/family/phone/{phone}/confirm", h.HandleConfirmFamilyByPhone)

			req := httptest.NewRequest(http.MethodPatch, "/api/guests/family/phone/"+tc.phone+"/confirm", nil)
			req = withTestClaims(req, "TST01")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tc.wantStatus, w.Code, w.Body.String())
			}
			if !tc.callsSvc && svcCalled {
				t.Fatal("service should not have been called for invalid phone")
			}
		})

		t.Run("cancel_family_by_phone/"+tc.name, func(t *testing.T) {
			h, repo := newTestHandler()
			svcCalled := false
			repo.getFamilyGroupByPhoneFn = func(ctx context.Context, phone string) (*int64, error) {
				svcCalled = true
				fg := int64(1)
				return &fg, nil
			}
			repo.getByIDAnyFn = func(ctx context.Context, id int64) (*Guest, error) {
				g := sampleGuest()
				return &g, nil
			}
			repo.setAttendingByFamilyGroupFn = func(ctx context.Context, familyGroup int64, attending bool, userRACF string) ([]Guest, error) {
				g := sampleGuest()
				g.Attending = boolPtr(false)
				return []Guest{g}, nil
			}

			mux := http.NewServeMux()
			mux.HandleFunc("PATCH /api/guests/family/phone/{phone}/cancel", h.HandleCancelFamilyByPhone)

			req := httptest.NewRequest(http.MethodPatch, "/api/guests/family/phone/"+tc.phone+"/cancel", nil)
			req = withTestClaims(req, "TST01")
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d: %s", tc.wantStatus, w.Code, w.Body.String())
			}
			if !tc.callsSvc && svcCalled {
				t.Fatal("service should not have been called for invalid phone")
			}
		})
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
