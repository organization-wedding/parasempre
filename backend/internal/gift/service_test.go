package gift

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type mockRepository struct {
	listFn             func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error)
	getByIDFn          func(ctx context.Context, id int64) (*Gift, error)
	createFn           func(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error)
	updateFn           func(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error)
	deleteFn           func(ctx context.Context, id int64, userRACF string) error
	findByDedupeKeysFn func(ctx context.Context, keys []string) (map[string]bool, error)
}

func (m *mockRepository) List(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
	return m.listFn(ctx, limit, offset, statusFilter)
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*Gift, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepository) Create(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error) {
	return m.createFn(ctx, input, dedupeKey, userRACF)
}

func (m *mockRepository) Update(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error) {
	return m.updateFn(ctx, id, input, dedupeKey, userRACF)
}

func (m *mockRepository) Delete(ctx context.Context, id int64, userRACF string) error {
	return m.deleteFn(ctx, id, userRACF)
}

func (m *mockRepository) FindByDedupeKeys(ctx context.Context, keys []string) (map[string]bool, error) {
	if m.findByDedupeKeysFn != nil {
		return m.findByDedupeKeysFn(ctx, keys)
	}
	return map[string]bool{}, nil
}

func (m *mockRepository) WithTx(_ pgx.Tx) Repository {
	return m
}

func (m *mockRepository) BulkCreate(ctx context.Context, inputs []CreateGiftInput, dedupeKeys []string, userRACF string) ([]Gift, error) {
	if len(inputs) != len(dedupeKeys) {
		return nil, errors.New("inputs/keys length mismatch")
	}
	created := make([]Gift, 0, len(inputs))
	for i, input := range inputs {
		if m.createFn == nil {
			continue
		}
		g, err := m.createFn(ctx, input, dedupeKeys[i], userRACF)
		if err != nil {
			return nil, err
		}
		created = append(created, *g)
	}
	return created, nil
}

type mockTxRunner struct{}

func (m *mockTxRunner) RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return fn(nil)
}

type mockScraper struct {
	scrapeProductFn func(ctx context.Context, url string) (*ScrapedProduct, error)
}

func (m *mockScraper) ScrapeProduct(ctx context.Context, url string) (*ScrapedProduct, error) {
	if m.scrapeProductFn == nil {
		return &ScrapedProduct{}, nil
	}
	return m.scrapeProductFn(ctx, url)
}

func sampleGift() Gift {
	return Gift{
		ID:         1,
		Name:       "Panela Inox",
		PriceCents: 19990,
		Status:     "active",
		DedupeKey:  "panela inox",
		CreatedBy:  "TST01",
		UpdatedBy:  "TST01",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

func assertAppError(t *testing.T, err error, wantCode int, wantMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantMsg)
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != wantCode {
		t.Fatalf("expected code %d, got %d", wantCode, ae.Code)
	}
	if !strings.Contains(ae.Message, wantMsg) {
		t.Fatalf("expected message containing %q, got %q", wantMsg, ae.Message)
	}
}

func strPtr(s string) *string { return &s }

func int64Ptr(v int64) *int64 { return &v }

func TestServiceList(t *testing.T) {
	tests := []struct {
		name      string
		mockFn    func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error)
		page      int
		limit     int
		wantLen   int
		wantTotal int
		wantErr   bool
	}{
		{
			name: "returns gifts with pagination",
			mockFn: func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
				return []Gift{sampleGift()}, 5, nil
			},
			page: 1, limit: 20,
			wantLen: 1, wantTotal: 5,
		},
		{
			name: "defaults for invalid page/limit",
			mockFn: func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
				if limit != 20 || offset != 0 {
					return nil, 0, errors.New("expected default limit=20, offset=0")
				}
				return []Gift{}, 0, nil
			},
			page: 0, limit: 0,
		},
		{
			name: "caps limit at 100",
			mockFn: func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
				if limit != 100 {
					return nil, 0, errors.New("expected limit capped at 100")
				}
				return []Gift{}, 0, nil
			},
			page: 1, limit: 500,
		},
		{
			name: "propagates error",
			mockFn: func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
				return nil, 0, errors.New("db error")
			},
			page: 1, limit: 20, wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{listFn: tt.mockFn}, &mockTxRunner{}, nil)
			result, err := svc.List(context.Background(), tt.page, tt.limit, nil)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(result.Data) != tt.wantLen {
				t.Fatalf("expected %d gifts, got %d", tt.wantLen, len(result.Data))
			}
			if result.Total != tt.wantTotal {
				t.Fatalf("expected total %d, got %d", tt.wantTotal, result.Total)
			}
		})
	}
}

func TestServiceListPassesStatusFilter(t *testing.T) {
	var gotStatus *string
	repo := &mockRepository{
		listFn: func(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
			gotStatus = statusFilter
			return []Gift{}, 0, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, nil)

	active := "active"
	_, err := svc.List(context.Background(), 1, 20, &active)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotStatus == nil || *gotStatus != "active" {
		t.Fatalf("expected status filter 'active', got %v", gotStatus)
	}
}

func TestServiceGetByID(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, id int64) (*Gift, error)
		wantErr bool
	}{
		{
			name: "returns gift",
			mockFn: func(ctx context.Context, id int64) (*Gift, error) {
				g := sampleGift()
				return &g, nil
			},
		},
		{
			name: "not found",
			mockFn: func(ctx context.Context, id int64) (*Gift, error) {
				return nil, apperror.NotFound("gift not found")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{getByIDFn: tt.mockFn}, &mockTxRunner{}, nil)
			g, err := svc.GetByID(context.Background(), 1)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if g == nil {
				t.Fatal("expected gift, got nil")
			}
		})
	}
}

func TestServiceCreate(t *testing.T) {
	tests := []struct {
		name        string
		input       CreateGiftInput
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
	}{
		{
			name:  "valid input",
			input: CreateGiftInput{Name: "Panela", PriceCents: 19990},
		},
		{
			name:        "missing name",
			input:       CreateGiftInput{PriceCents: 19990},
			wantErr:     true,
			wantErrMsg:  "name is required",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "zero price",
			input:       CreateGiftInput{Name: "Panela", PriceCents: 0},
			wantErr:     true,
			wantErrMsg:  "price_cents is required",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "negative price",
			input:       CreateGiftInput{Name: "Panela", PriceCents: -100},
			wantErr:     true,
			wantErrMsg:  "price_cents must be greater than 0",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "non-https image url",
			input:       CreateGiftInput{Name: "Panela", PriceCents: 19990, ImageURL: strPtr("http://img.example.com/p.jpg")},
			wantErr:     true,
			wantErrMsg:  "image_url must start with https://",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "invalid status",
			input:       CreateGiftInput{Name: "Panela", PriceCents: 19990, Status: strPtr("disabled")},
			wantErr:     true,
			wantErrMsg:  "status must be 'active' or 'inactive'",
			wantErrCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				createFn: func(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error) {
					g := sampleGift()
					return &g, nil
				},
			}
			svc := NewService(repo, &mockTxRunner{}, nil)
			_, err := svc.Create(context.Background(), tt.input, "TST01")
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceCreateNormalizesDedupeKey(t *testing.T) {
	var gotKey string
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error) {
			gotKey = dedupeKey
			g := sampleGift()
			return &g, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, nil)

	_, err := svc.Create(context.Background(), CreateGiftInput{
		Name:       "  Máquina  de  Café  ",
		PriceCents: 100000,
	}, "TST01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotKey != "maquina de cafe" {
		t.Fatalf("expected normalized dedupe_key 'maquina de cafe', got %q", gotKey)
	}
}

func TestServiceCreateDuplicate(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error) {
			return nil, apperror.Conflict("Já existe um presente com esse nome.")
		},
	}
	svc := NewService(repo, &mockTxRunner{}, nil)

	_, err := svc.Create(context.Background(), CreateGiftInput{
		Name:       "Panela",
		PriceCents: 19990,
	}, "TST01")
	assertAppError(t, err, http.StatusConflict, "Já existe um presente com esse nome.")
}

func TestServiceUpdate(t *testing.T) {
	name := "Novo Nome"
	neg := int64(-1)
	invalidStatus := "disabled"

	tests := []struct {
		name        string
		input       UpdateGiftInput
		wantErr     bool
		wantErrMsg  string
		wantErrCode int
	}{
		{name: "valid partial update", input: UpdateGiftInput{Name: &name}},
		{name: "empty update is valid", input: UpdateGiftInput{}},
		{
			name:        "negative price",
			input:       UpdateGiftInput{PriceCents: &neg},
			wantErr:     true,
			wantErrMsg:  "price_cents must be greater than 0",
			wantErrCode: http.StatusBadRequest,
		},
		{
			name:        "invalid status",
			input:       UpdateGiftInput{Status: &invalidStatus},
			wantErr:     true,
			wantErrMsg:  "status must be 'active' or 'inactive'",
			wantErrCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepository{
				updateFn: func(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error) {
					g := sampleGift()
					return &g, nil
				},
			}
			svc := NewService(repo, &mockTxRunner{}, nil)
			_, err := svc.Update(context.Background(), 1, tt.input, "TST01")
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceUpdateRecalculatesDedupeKeyOnlyWhenNameChanges(t *testing.T) {
	tests := []struct {
		name       string
		input      UpdateGiftInput
		wantKey    *string
		wantKeySet bool
	}{
		{
			name:       "name changed -> dedupe_key recomputed",
			input:      UpdateGiftInput{Name: strPtr("Máquina de Café")},
			wantKey:    strPtr("maquina de cafe"),
			wantKeySet: true,
		},
		{
			name:       "name nil -> dedupe_key nil",
			input:      UpdateGiftInput{PriceCents: int64Ptr(50000)},
			wantKeySet: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotKey *string
			repo := &mockRepository{
				updateFn: func(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error) {
					gotKey = dedupeKey
					g := sampleGift()
					return &g, nil
				},
			}
			svc := NewService(repo, &mockTxRunner{}, nil)
			_, err := svc.Update(context.Background(), 1, tt.input, "TST01")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantKeySet {
				if gotKey == nil || *gotKey != *tt.wantKey {
					t.Fatalf("expected dedupe_key %q, got %v", *tt.wantKey, gotKey)
				}
			} else if gotKey != nil {
				t.Fatalf("expected dedupe_key nil, got %q", *gotKey)
			}
		})
	}
}

func TestServiceDelete(t *testing.T) {
	tests := []struct {
		name    string
		mockFn  func(ctx context.Context, id int64, userRACF string) error
		wantErr bool
	}{
		{name: "success", mockFn: func(ctx context.Context, id int64, userRACF string) error { return nil }},
		{
			name:    "not found",
			mockFn:  func(ctx context.Context, id int64, userRACF string) error { return apperror.NotFound("gift not found") },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{deleteFn: tt.mockFn}, &mockTxRunner{}, nil)
			err := svc.Delete(context.Background(), 1, "TST01")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestServiceScrapePreview(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		scrapeFn       func(ctx context.Context, url string) (*ScrapedProduct, error)
		wantName       string
		wantDesc       string
		wantImage      string
		wantStore      string
		wantPriceCents int64
		wantErr        bool
		wantErrCode    int
		wantErrMsg     string
	}{
		{
			name: "success all fields",
			url:  "https://shop.example.com/p/1",
			scrapeFn: func(ctx context.Context, url string) (*ScrapedProduct, error) {
				return &ScrapedProduct{
					Name:        "Jogo de Panelas",
					Description: "Tramontina inox",
					PriceBRL:    "459,90",
					ImageURL:    "https://shop.example.com/img.jpg",
				}, nil
			},
			wantName:       "Jogo de Panelas",
			wantDesc:       "Tramontina inox",
			wantImage:      "https://shop.example.com/img.jpg",
			wantStore:      "https://shop.example.com/p/1",
			wantPriceCents: 45990,
		},
		{
			name: "price with BR thousands separator",
			url:  "https://shop.example.com/p/2",
			scrapeFn: func(ctx context.Context, url string) (*ScrapedProduct, error) {
				return &ScrapedProduct{Name: "Geladeira", PriceBRL: "R$ 1.234,56"}, nil
			},
			wantName:       "Geladeira",
			wantStore:      "https://shop.example.com/p/2",
			wantPriceCents: 123456,
		},
		{
			name: "invalid price falls back to zero",
			url:  "https://shop.example.com/p/3",
			scrapeFn: func(ctx context.Context, url string) (*ScrapedProduct, error) {
				return &ScrapedProduct{Name: "Mixer", PriceBRL: "sob consulta"}, nil
			},
			wantName:       "Mixer",
			wantStore:      "https://shop.example.com/p/3",
			wantPriceCents: 0,
		},
		{
			name: "non-https image url is discarded",
			url:  "https://shop.example.com/p/4",
			scrapeFn: func(ctx context.Context, url string) (*ScrapedProduct, error) {
				return &ScrapedProduct{Name: "Cafeteira", ImageURL: "http://shop.example.com/img.jpg"}, nil
			},
			wantName:  "Cafeteira",
			wantImage: "",
			wantStore: "https://shop.example.com/p/4",
		},
		{
			name: "empty extraction returns 400",
			url:  "https://shop.example.com/p/5",
			scrapeFn: func(ctx context.Context, url string) (*ScrapedProduct, error) {
				return &ScrapedProduct{}, nil
			},
			wantErr:     true,
			wantErrCode: http.StatusBadRequest,
			wantErrMsg:  "Não conseguimos identificar o produto",
		},
		{
			name:        "invalid input url",
			url:         "not-a-url",
			scrapeFn:    func(ctx context.Context, url string) (*ScrapedProduct, error) { return nil, nil },
			wantErr:     true,
			wantErrCode: http.StatusBadRequest,
			wantErrMsg:  "URL inválida",
		},
		{
			name: "scraper returns error",
			url:  "https://shop.example.com/p/6",
			scrapeFn: func(ctx context.Context, url string) (*ScrapedProduct, error) {
				return nil, apperror.Validation("Não conseguimos buscar dados desta URL.")
			},
			wantErr:     true,
			wantErrCode: http.StatusBadRequest,
			wantErrMsg:  "Não conseguimos buscar dados",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(&mockRepository{}, &mockTxRunner{}, &mockScraper{scrapeProductFn: tt.scrapeFn})
			got, err := svc.ScrapePreview(context.Background(), tt.url, "TST01")
			if tt.wantErr {
				assertAppError(t, err, tt.wantErrCode, tt.wantErrMsg)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name: want %q, got %q", tt.wantName, got.Name)
			}
			if got.Description != tt.wantDesc {
				t.Errorf("Description: want %q, got %q", tt.wantDesc, got.Description)
			}
			if got.ImageURL != tt.wantImage {
				t.Errorf("ImageURL: want %q, got %q", tt.wantImage, got.ImageURL)
			}
			if got.StoreURL != tt.wantStore {
				t.Errorf("StoreURL: want %q, got %q", tt.wantStore, got.StoreURL)
			}
			if got.PriceCents != tt.wantPriceCents {
				t.Errorf("PriceCents: want %d, got %d", tt.wantPriceCents, got.PriceCents)
			}
		})
	}
}

func TestServiceScrapePreviewReturns503WhenScraperNil(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockTxRunner{}, nil)
	_, err := svc.ScrapePreview(context.Background(), "https://x.com/p", "TST01")
	assertAppError(t, err, http.StatusServiceUnavailable, "Busca por link não está configurada")
}

func TestServiceScrapePreviewTrimsName(t *testing.T) {
	longName := strings.Repeat("a", 250)
	svc := NewService(&mockRepository{}, &mockTxRunner{}, &mockScraper{
		scrapeProductFn: func(ctx context.Context, url string) (*ScrapedProduct, error) {
			return &ScrapedProduct{Name: longName}, nil
		},
	})
	got, err := svc.ScrapePreview(context.Background(), "https://x.com/p", "TST01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len([]rune(got.Name)) != scrapeMaxNameLen {
		t.Errorf("expected name truncated to %d runes, got %d", scrapeMaxNameLen, len([]rune(got.Name)))
	}
}

func TestServiceDeletePassesUserRACFToRepo(t *testing.T) {
	var gotRACF string
	repo := &mockRepository{
		deleteFn: func(ctx context.Context, id int64, userRACF string) error {
			gotRACF = userRACF
			return nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, nil)
	if err := svc.Delete(context.Background(), 1, "ABC12"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotRACF != "ABC12" {
		t.Fatalf("expected userRACF 'ABC12' passed to repo, got %q", gotRACF)
	}
}
