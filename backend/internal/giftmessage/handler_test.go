package giftmessage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
)

func newTestHandler(repo TxAwareRepository, txns TransactionFinder, storage Storage) *Handler {
	svc := NewService(repo, txns, storage, &mockAudit{}, time.Minute)
	return NewHandler(svc)
}

func authedRequest(method, target string, body io.Reader, userID int64, contentType string, pathID string) *http.Request {
	r := httptest.NewRequest(method, target, body)
	r.SetPathValue("id", pathID)
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	claims := &auth.Claims{UserID: userID, URACF: "ABC12", Role: "guest"}
	r = r.WithContext(middleware.WithClaims(context.Background(), claims))
	return r
}

func buildMultipart(t *testing.T, fields map[string]string, file *multipartFile) (string, *bytes.Buffer) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("WriteField(%s): %v", k, err)
		}
	}
	if file != nil {
		header := make(map[string][]string)
		header["Content-Disposition"] = []string{`form-data; name="media"; filename="` + file.filename + `"`}
		header["Content-Type"] = []string{file.contentType}
		fw, err := w.CreatePart(header)
		if err != nil {
			t.Fatalf("CreatePart: %v", err)
		}
		if _, err := fw.Write(file.body); err != nil {
			t.Fatalf("write file: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return w.FormDataContentType(), body
}

type multipartFile struct {
	filename, contentType string
	body                  []byte
}

func TestHandleCreate_TextOnly_Returns201(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, in CreateRow) (*GiftMessage, error) {
			return &GiftMessage{
				ID: 1, GiftID: in.GiftID, GiftTransactionID: in.GiftTransactionID,
				UserID: in.UserID, AuthorName: in.AuthorName, Content: in.Content,
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			}, nil
		},
	}
	txns := &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}
	h := newTestHandler(repo, txns, nil)

	contentType, body := buildMultipart(t, map[string]string{
		"author_name": "Maria",
		"content":     "Felicidades!",
	}, nil)
	r := authedRequest(http.MethodPost, "/api/transactions/100/message", body, 42, contentType, "100")
	w := httptest.NewRecorder()
	h.HandleCreate(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", w.Code, w.Body.String())
	}
	var resp PublicMessage
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.AuthorName != "Maria" {
		t.Errorf("expected author Maria, got %q", resp.AuthorName)
	}
	if resp.MediaURL != nil {
		t.Errorf("expected nil MediaURL, got %v", resp.MediaURL)
	}
}

func TestHandleCreate_WithMedia_Returns201AndUploads(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, in CreateRow) (*GiftMessage, error) {
			return &GiftMessage{
				ID: 1, GiftID: in.GiftID, GiftTransactionID: in.GiftTransactionID,
				UserID: in.UserID, AuthorName: in.AuthorName, Content: in.Content,
				MediaObjectKey: in.MediaObjectKey, MediaKind: in.MediaKind,
				MediaSizeBytes: in.MediaSizeBytes, MediaMimeType: in.MediaMimeType,
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			}, nil
		},
	}
	txns := &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}
	storage := &mockStorage{}
	h := newTestHandler(repo, txns, storage)

	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0, 1, 1, 0, 0, 1, 0, 1, 0, 0}
	contentType, body := buildMultipart(t, map[string]string{
		"author_name": "Maria",
		"content":     "Felicidades!",
	}, &multipartFile{filename: "foto.jpg", contentType: "image/jpeg", body: jpegBytes})
	r := authedRequest(http.MethodPost, "/api/transactions/100/message", body, 42, contentType, "100")
	w := httptest.NewRecorder()
	h.HandleCreate(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%s)", w.Code, w.Body.String())
	}
	var resp PublicMessage
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.MediaKind == nil || *resp.MediaKind != MediaKindImage {
		t.Errorf("expected image kind, got %v", resp.MediaKind)
	}
	if resp.MediaURL == nil {
		t.Errorf("expected signed media URL on response")
	}
}

func TestHandleCreate_RejectsAnotherUserTransaction(t *testing.T) {
	txns := &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) {
			tx := approvedTx()
			tx.UserID = 999 // outro usuário
			return tx, nil
		},
	}
	h := newTestHandler(&mockRepo{}, txns, nil)

	contentType, body := buildMultipart(t, map[string]string{
		"author_name": "X", "content": "tentativa",
	}, nil)
	r := authedRequest(http.MethodPost, "/api/transactions/100/message", body, 42, contentType, "100")
	w := httptest.NewRecorder()
	h.HandleCreate(w, r)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestHandleCreate_RejectsPendingTransaction(t *testing.T) {
	txns := &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) {
			tx := approvedTx()
			tx.Status = "pending"
			return tx, nil
		},
	}
	h := newTestHandler(&mockRepo{}, txns, nil)

	contentType, body := buildMultipart(t, map[string]string{
		"author_name": "X", "content": "tentativa",
	}, nil)
	r := authedRequest(http.MethodPost, "/api/transactions/100/message", body, 42, contentType, "100")
	w := httptest.NewRecorder()
	h.HandleCreate(w, r)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestHandleCreate_RejectsUnauthenticated(t *testing.T) {
	h := newTestHandler(&mockRepo{}, &mockTxFinder{}, nil)
	contentType, body := buildMultipart(t, map[string]string{
		"author_name": "X", "content": "tentativa",
	}, nil)
	r := httptest.NewRequest(http.MethodPost, "/api/transactions/100/message", body)
	r.Header.Set("Content-Type", contentType)
	r.SetPathValue("id", "100")
	w := httptest.NewRecorder()
	h.HandleCreate(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleGetMine_Returns404WhenAbsent(t *testing.T) {
	repo := &mockRepo{
		getByTxIDFn: func(_ context.Context, _ int64) (*GiftMessage, error) {
			return nil, apperror.NotFound("mensagem não encontrada")
		},
	}
	txns := &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}
	h := newTestHandler(repo, txns, nil)
	r := authedRequest(http.MethodGet, "/api/transactions/100/message", nil, 42, "", "100")
	w := httptest.NewRecorder()
	h.HandleGetMine(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestHandleCreate_BodyAt50MB_Passes(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, in CreateRow) (*GiftMessage, error) {
			return &GiftMessage{
				ID: 1, GiftID: in.GiftID, GiftTransactionID: in.GiftTransactionID,
				UserID: in.UserID, AuthorName: in.AuthorName, Content: in.Content,
				MediaObjectKey: in.MediaObjectKey, MediaKind: in.MediaKind,
				MediaSizeBytes: in.MediaSizeBytes, MediaMimeType: in.MediaMimeType,
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			}, nil
		},
	}
	txns := &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}
	storage := &mockStorage{}
	h := newTestHandler(repo, txns, storage)

	mp4Magic := []byte{0, 0, 0, 20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm',
		0, 0, 0, 0, 'i', 's', 'o', 'm'}
	videoSize := int64(50 * 1024 * 1024)
	padding := make([]byte, videoSize-int64(len(mp4Magic)))
	videoBytes := append(mp4Magic, padding...)

	ct, body := buildMultipart(t,
		map[string]string{"author_name": "Maria", "content": "video test"},
		&multipartFile{filename: "clip.mp4", contentType: "video/mp4", body: videoBytes},
	)
	r := authedRequest(http.MethodPost, "/api/transactions/100/message", body, 42, ct, "100")
	w := httptest.NewRecorder()
	h.HandleCreate(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("50 MB video should be accepted, got %d (body=%s)", w.Code, w.Body.String())
	}
}

func TestHandleCreate_BodyAboveLimit_Returns400WithSizeError(t *testing.T) {
	h := newTestHandler(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, &mockStorage{})

	oversizeBytes := make([]byte, 56*1024*1024)
	ct, body := buildMultipart(t,
		map[string]string{"author_name": "X", "content": "y"},
		&multipartFile{filename: "big.mp4", contentType: "video/mp4", body: oversizeBytes},
	)
	r := authedRequest(http.MethodPost, "/api/transactions/100/message", body, 42, ct, "100")
	w := httptest.NewRecorder()
	h.HandleCreate(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("oversized body should return 400, got %d (body=%s)", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	errMsg, ok := resp["error"]
	if !ok {
		t.Fatalf("expected 'error' field in response, got %v", resp)
	}
	if strings.Contains(errMsg, "multipart") {
		t.Errorf("error message should not mention 'multipart' for a size rejection, got %q", errMsg)
	}
}

func TestHandleAdminDelete_Returns204(t *testing.T) {
	repo := &mockRepo{
		softDeleteFn: func(_ context.Context, _, _ int64) error { return nil },
	}
	h := newTestHandler(repo, &mockTxFinder{}, nil)
	r := authedRequest(http.MethodDelete, "/api/admin/gift-messages/7", nil, 99, "", "7")
	w := httptest.NewRecorder()
	h.HandleAdminDelete(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
