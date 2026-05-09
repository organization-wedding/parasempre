package giftmessage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

// --- mocks ---

type mockRepo struct {
	createFn       func(ctx context.Context, in CreateRow) (*GiftMessage, error)
	getByIDFn     func(ctx context.Context, id int64) (*GiftMessage, error)
	getByTxIDFn   func(ctx context.Context, txID int64) (*GiftMessage, error)
	listByGiftFn  func(ctx context.Context, giftID int64, limit, offset int) ([]GiftMessage, int, error)
	listAllFn     func(ctx context.Context, limit, offset int) ([]GiftMessage, int, error)
	softDeleteFn  func(ctx context.Context, id, byUserID int64) error
}

func (m *mockRepo) Create(ctx context.Context, in CreateRow) (*GiftMessage, error) {
	return m.createFn(ctx, in)
}
func (m *mockRepo) GetByID(ctx context.Context, id int64) (*GiftMessage, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepo) GetByTransactionID(ctx context.Context, txID int64) (*GiftMessage, error) {
	if m.getByTxIDFn == nil {
		return nil, apperror.NotFound("not found")
	}
	return m.getByTxIDFn(ctx, txID)
}
func (m *mockRepo) ListByGift(ctx context.Context, giftID int64, limit, offset int) ([]GiftMessage, int, error) {
	return m.listByGiftFn(ctx, giftID, limit, offset)
}
func (m *mockRepo) ListAll(ctx context.Context, limit, offset int) ([]GiftMessage, int, error) {
	return m.listAllFn(ctx, limit, offset)
}
func (m *mockRepo) SoftDelete(ctx context.Context, id, byUserID int64) error {
	return m.softDeleteFn(ctx, id, byUserID)
}
func (m *mockRepo) WithTx(_ pgx.Tx) Repository { return m }

type mockTxFinder struct {
	getFn func(ctx context.Context, id int64) (*TransactionSnapshot, error)
}

func (m *mockTxFinder) GetByID(ctx context.Context, id int64) (*TransactionSnapshot, error) {
	return m.getFn(ctx, id)
}

type mockStorage struct {
	uploadFn   func(ctx context.Context, key, mime string, r io.Reader, size int64) error
	signFn     func(ctx context.Context, keys []string, ttl time.Duration) (map[string]string, error)
	deleteFn   func(ctx context.Context, key string) error
	deleteCalls int
}

func (m *mockStorage) Upload(ctx context.Context, key, mime string, r io.Reader, size int64) error {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, key, mime, r, size)
	}
	_, _ = io.Copy(io.Discard, r)
	return nil
}
func (m *mockStorage) SignURLs(ctx context.Context, keys []string, ttl time.Duration) (map[string]string, error) {
	if m.signFn != nil {
		return m.signFn(ctx, keys, ttl)
	}
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		out[k] = "https://signed.example/" + k
	}
	return out, nil
}
func (m *mockStorage) Delete(ctx context.Context, key string) error {
	m.deleteCalls++
	if m.deleteFn != nil {
		return m.deleteFn(ctx, key)
	}
	return nil
}

type mockAudit struct {
	calls []auditCall
}
type auditCall struct {
	userID  int64
	action  string
	details map[string]any
}

func (m *mockAudit) LogAction(_ context.Context, userID int64, action string, details map[string]any) error {
	m.calls = append(m.calls, auditCall{userID: userID, action: action, details: details})
	return nil
}

// --- helpers ---

func approvedTx() *TransactionSnapshot {
	return &TransactionSnapshot{ID: 100, GiftID: 1, UserID: 42, Status: "approved"}
}

func validInput() CreateInput {
	return CreateInput{AuthorName: "Maria", Content: "Felicidades!"}
}

// jpegMedia gera um Media com bytes que http.DetectContentType identifica
// como image/jpeg.
func jpegMedia(size int64) *Media {
	hdr := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0, 1}
	body := make([]byte, size-int64(len(hdr)))
	return &Media{
		DeclaredMime: "image/jpeg",
		Size:         size,
		Reader:       io.MultiReader(bytes.NewReader(hdr), bytes.NewReader(body)),
	}
}

func assertAppError(t *testing.T, err error, wantCode int, wantMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected AppError containing %q, got nil", wantMsg)
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != wantCode {
		t.Fatalf("expected code %d, got %d (msg=%q)", wantCode, ae.Code, ae.Message)
	}
	if !strings.Contains(ae.Message, wantMsg) {
		t.Fatalf("expected message containing %q, got %q", wantMsg, ae.Message)
	}
}

// --- tests ---

func TestCreate_RejectsUnauthenticated(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{}, nil, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 1, 0, validInput(), nil)
	assertAppError(t, err, http.StatusUnauthorized, "autenticação")
}

func TestCreate_TransactionNotFound(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) {
			return nil, apperror.NotFound("transaction not found")
		},
	}, nil, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 999, 42, validInput(), nil)
	assertAppError(t, err, http.StatusNotFound, "transaction not found")
}

func TestCreate_TransactionFromAnotherUser(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) {
			tx := approvedTx()
			tx.UserID = 999
			return tx, nil
		},
	}, nil, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 100, 42, validInput(), nil)
	assertAppError(t, err, http.StatusForbidden, "não pertence")
}

func TestCreate_TransactionPending(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) {
			tx := approvedTx()
			tx.Status = "pending"
			return tx, nil
		},
	}, nil, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 100, 42, validInput(), nil)
	assertAppError(t, err, http.StatusConflict, "aguarde")
}

func TestCreate_RejectsContentTooLong(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, nil, &mockAudit{}, time.Minute)
	in := validInput()
	in.Content = strings.Repeat("a", 501)
	_, err := svc.Create(context.Background(), 100, 42, in, nil)
	assertAppError(t, err, http.StatusBadRequest, "Content")
}

func TestCreate_PreCheckFindsExistingMessage(t *testing.T) {
	repo := &mockRepo{
		getByTxIDFn: func(_ context.Context, _ int64) (*GiftMessage, error) {
			return &GiftMessage{ID: 7}, nil
		},
	}
	svc := NewService(repo, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, nil, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 100, 42, validInput(), nil)
	assertAppError(t, err, http.StatusConflict, "já existe")
}

func TestCreate_HappyPathTextOnly(t *testing.T) {
	var inserted CreateRow
	repo := &mockRepo{
		createFn: func(_ context.Context, in CreateRow) (*GiftMessage, error) {
			inserted = in
			return &GiftMessage{
				ID: 1, GiftTransactionID: in.GiftTransactionID, GiftID: in.GiftID,
				UserID: in.UserID, AuthorName: in.AuthorName, Content: in.Content,
				CreatedAt: time.Now(), UpdatedAt: time.Now(),
			}, nil
		},
	}
	audit := &mockAudit{}
	svc := NewService(repo, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, nil, audit, time.Minute)
	msg, err := svc.Create(context.Background(), 100, 42, validInput(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != 1 {
		t.Errorf("expected ID 1, got %d", msg.ID)
	}
	if inserted.MediaObjectKey != nil {
		t.Errorf("expected no media for text-only, got key=%v", *inserted.MediaObjectKey)
	}
	if len(audit.calls) != 1 || audit.calls[0].action != auditMessageCreated {
		t.Fatalf("expected audit %q, got %v", auditMessageCreated, audit.calls)
	}
}

func TestCreate_HappyPathWithMedia(t *testing.T) {
	var uploadedKey, uploadedMime string
	storage := &mockStorage{
		uploadFn: func(_ context.Context, key, mime string, r io.Reader, _ int64) error {
			uploadedKey = key
			uploadedMime = mime
			_, _ = io.Copy(io.Discard, r)
			return nil
		},
	}
	repo := &mockRepo{
		createFn: func(_ context.Context, in CreateRow) (*GiftMessage, error) {
			return &GiftMessage{
				ID: 1, GiftTransactionID: in.GiftTransactionID, GiftID: in.GiftID,
				UserID: in.UserID, AuthorName: in.AuthorName, Content: in.Content,
				MediaObjectKey: in.MediaObjectKey, MediaKind: in.MediaKind,
				MediaSizeBytes: in.MediaSizeBytes, MediaMimeType: in.MediaMimeType,
			}, nil
		},
	}
	svc := NewService(repo, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, storage, &mockAudit{}, time.Minute)
	msg, err := svc.Create(context.Background(), 100, 42, validInput(), jpegMedia(2048))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.MediaKind == nil || *msg.MediaKind != MediaKindImage {
		t.Errorf("expected image kind, got %v", msg.MediaKind)
	}
	if !strings.HasPrefix(uploadedKey, "messages/100/") {
		t.Errorf("uploaded key should be in messages/100/, got %q", uploadedKey)
	}
	if uploadedMime != "image/jpeg" {
		t.Errorf("expected uploaded mime image/jpeg, got %q", uploadedMime)
	}
}

func TestCreate_RejectsUnsupportedMime(t *testing.T) {
	storage := &mockStorage{}
	svc := NewService(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, storage, &mockAudit{}, time.Minute)
	textBody := "this is plain text content not a real media file blah blah"
	media := &Media{
		DeclaredMime: "image/jpeg", // mente
		Size:         int64(len(textBody)),
		Reader:       strings.NewReader(textBody),
	}
	_, err := svc.Create(context.Background(), 100, 42, validInput(), media)
	assertAppError(t, err, http.StatusBadRequest, "tipo de mídia")
}

func TestCreate_DeletesObjectIfRepoCreateFails(t *testing.T) {
	storage := &mockStorage{}
	repo := &mockRepo{
		createFn: func(_ context.Context, _ CreateRow) (*GiftMessage, error) {
			return nil, apperror.Conflict("já existe uma mensagem para essa transação")
		},
	}
	svc := NewService(repo, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, storage, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 100, 42, validInput(), jpegMedia(2048))
	assertAppError(t, err, http.StatusConflict, "já existe")
	if storage.deleteCalls != 1 {
		t.Errorf("expected storage.Delete to be called once for orphan cleanup, got %d", storage.deleteCalls)
	}
}

func TestCreate_RejectsUploadFailure(t *testing.T) {
	storage := &mockStorage{
		uploadFn: func(_ context.Context, _, _ string, _ io.Reader, _ int64) error {
			return errors.New("supabase down")
		},
	}
	createCalled := false
	repo := &mockRepo{
		createFn: func(_ context.Context, _ CreateRow) (*GiftMessage, error) {
			createCalled = true
			return nil, nil
		},
	}
	svc := NewService(repo, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, storage, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 100, 42, validInput(), jpegMedia(2048))
	if err == nil {
		t.Fatal("expected error from upload failure")
	}
	if createCalled {
		t.Error("expected repo.Create NOT to be called when upload fails")
	}
}

func TestCreate_StorageDisabledRejectsMedia(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, nil, &mockAudit{}, time.Minute)
	_, err := svc.Create(context.Background(), 100, 42, validInput(), jpegMedia(2048))
	assertAppError(t, err, http.StatusServiceUnavailable, "indisponíveis")
}

func TestGetMine_ReturnsNilWhenNoMessage(t *testing.T) {
	repo := &mockRepo{
		getByTxIDFn: func(_ context.Context, _ int64) (*GiftMessage, error) {
			return nil, apperror.NotFound("mensagem não encontrada")
		},
	}
	svc := NewService(repo, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) { return approvedTx(), nil },
	}, nil, &mockAudit{}, time.Minute)
	msg, err := svc.GetMine(context.Background(), 100, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg != nil {
		t.Errorf("expected nil for no message, got %+v", msg)
	}
}

func TestGetMine_RejectsAccessFromAnotherUser(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{
		getFn: func(_ context.Context, _ int64) (*TransactionSnapshot, error) {
			tx := approvedTx()
			tx.UserID = 1
			return tx, nil
		},
	}, nil, &mockAudit{}, time.Minute)
	_, err := svc.GetMine(context.Background(), 100, 42)
	assertAppError(t, err, http.StatusForbidden, "não pertence")
}

func TestListByGift_SignsMediaURLs(t *testing.T) {
	key1 := "messages/1/abc.jpg"
	key2 := "messages/1/def.mp4"
	now := time.Now()
	repo := &mockRepo{
		listByGiftFn: func(_ context.Context, _ int64, _, _ int) ([]GiftMessage, int, error) {
			return []GiftMessage{
				{ID: 1, GiftID: 1, AuthorName: "A", Content: "hi", MediaObjectKey: &key1, MediaKind: ptr(MediaKindImage), CreatedAt: now},
				{ID: 2, GiftID: 1, AuthorName: "B", Content: "yo", MediaObjectKey: &key2, MediaKind: ptr(MediaKindVideo), CreatedAt: now},
				{ID: 3, GiftID: 1, AuthorName: "C", Content: "no media", CreatedAt: now},
			}, 3, nil
		},
	}
	var signedKeys []string
	storage := &mockStorage{
		signFn: func(_ context.Context, keys []string, _ time.Duration) (map[string]string, error) {
			signedKeys = keys
			out := make(map[string]string, len(keys))
			for _, k := range keys {
				out[k] = "https://signed/" + k
			}
			return out, nil
		},
	}
	svc := NewService(repo, &mockTxFinder{}, storage, &mockAudit{}, time.Minute)
	resp, err := svc.ListByGift(context.Background(), 1, 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(resp.Data))
	}
	if len(signedKeys) != 2 {
		t.Errorf("expected 2 keys signed in batch, got %d", len(signedKeys))
	}
	if resp.Data[0].MediaURL == nil || *resp.Data[0].MediaURL != "https://signed/"+key1 {
		t.Errorf("expected signed url for key1, got %v", resp.Data[0].MediaURL)
	}
	if resp.Data[2].MediaURL != nil {
		t.Errorf("expected nil URL for message without media, got %v", resp.Data[2].MediaURL)
	}
}

func TestRemove_RecordsAudit(t *testing.T) {
	repo := &mockRepo{
		softDeleteFn: func(_ context.Context, id, byUserID int64) error {
			if id != 7 || byUserID != 99 {
				t.Errorf("unexpected args id=%d byUserID=%d", id, byUserID)
			}
			return nil
		},
	}
	audit := &mockAudit{}
	svc := NewService(repo, &mockTxFinder{}, nil, audit, time.Minute)
	if err := svc.Remove(context.Background(), 7, 99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(audit.calls) != 1 || audit.calls[0].action != auditMessageRemoved {
		t.Fatalf("expected %q audit, got %v", auditMessageRemoved, audit.calls)
	}
}

func TestRemove_RejectsUnauthenticated(t *testing.T) {
	svc := NewService(&mockRepo{}, &mockTxFinder{}, nil, &mockAudit{}, time.Minute)
	err := svc.Remove(context.Background(), 1, 0)
	assertAppError(t, err, http.StatusUnauthorized, "autenticação")
}

func TestNormalizePaging(t *testing.T) {
	cases := []struct {
		page, limit, defLimit, maxLimit int
		wantPage, wantLimit             int
	}{
		{0, 0, 10, 50, 1, 10},
		{-3, -1, 10, 50, 1, 10},
		{2, 99, 10, 50, 2, 50},
		{5, 25, 10, 50, 5, 25},
	}
	for _, c := range cases {
		gotPage, gotLimit := normalizePaging(c.page, c.limit, c.defLimit, c.maxLimit)
		if gotPage != c.wantPage || gotLimit != c.wantLimit {
			t.Errorf("normalizePaging(%d, %d) = (%d, %d), want (%d, %d)",
				c.page, c.limit, gotPage, gotLimit, c.wantPage, c.wantLimit)
		}
	}
}

func ptr[T any](v T) *T { return &v }
