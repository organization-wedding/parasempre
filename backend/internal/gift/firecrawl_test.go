package gift

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

func TestFirecrawlClientScrapeProduct_Success_ExtractField(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedBody)

		_, _ = io.WriteString(w, `{
			"success": true,
			"data": {
				"extract": {
					"name": "1 AP. DE JANTAR/CHA 30PCS",
					"price_brl": "R$506,98",
					"description": "Coleção Brisa",
					"image_url": "https://m.media-amazon.com/images/I/61WfjowoKRL.jpg"
				}
			}
		}`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("k", server.URL)
	got, err := c.ScrapeProduct(context.Background(), "https://www.amazon.com.br/dp/B0CFFQBY89")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "1 AP. DE JANTAR/CHA 30PCS" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.PriceBRL != "R$506,98" {
		t.Errorf("price: got %q", got.PriceBRL)
	}
	if got.ImageURL != "https://m.media-amazon.com/images/I/61WfjowoKRL.jpg" {
		t.Errorf("image_url: got %q", got.ImageURL)
	}

	formats, _ := capturedBody["formats"].([]any)
	if len(formats) != 1 || formats[0] != "extract" {
		t.Errorf("expected formats=[extract], got %v", formats)
	}
	extract, ok := capturedBody["extract"].(map[string]any)
	if !ok {
		t.Fatalf("expected extract object in payload, got %T", capturedBody["extract"])
	}
	if _, hasSchema := extract["schema"]; !hasSchema {
		t.Errorf("expected extract.schema in payload, got %v", extract)
	}
	if _, hasJSONOpts := capturedBody["jsonOptions"]; hasJSONOpts {
		t.Errorf("legacy jsonOptions should not be in payload")
	}
	if proxy, _ := capturedBody["proxy"].(string); proxy != "auto" {
		t.Errorf("expected proxy=auto to bypass anti-bot pages, got %q", proxy)
	}
}

func TestFirecrawlClientScrapeProduct_Success_JSONField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("expected Authorization header 'Bearer test-key', got %q", got)
		}
		if r.URL.Path != "/v1/scrape" {
			t.Errorf("expected /v1/scrape, got %s", r.URL.Path)
		}

		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("invalid request body: %v", err)
		}
		if req["url"] != "https://shop.example.com/p/123" {
			t.Errorf("expected url forwarded, got %v", req["url"])
		}

		_, _ = io.WriteString(w, `{
			"success": true,
			"data": {
				"json": {
					"name": "Jogo de Panelas",
					"price_brl": "459,90",
					"description": "Tramontina inox",
					"image_url": "https://shop.example.com/img.jpg"
				}
			}
		}`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("test-key", server.URL)
	got, err := c.ScrapeProduct(context.Background(), "https://shop.example.com/p/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Jogo de Panelas" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.PriceBRL != "459,90" {
		t.Errorf("price: got %q", got.PriceBRL)
	}
	if got.ImageURL != "https://shop.example.com/img.jpg" {
		t.Errorf("image_url: got %q", got.ImageURL)
	}
}

func TestFirecrawlClientScrapeProduct_Success_LLMExtractionFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{
			"success": true,
			"data": {
				"llm_extraction": {
					"name": "Liquidificador",
					"price_brl": "199.00"
				}
			}
		}`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("k", server.URL)
	got, err := c.ScrapeProduct(context.Background(), "https://x.com/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Liquidificador" {
		t.Errorf("name: got %q", got.Name)
	}
	if got.PriceBRL != "199.00" {
		t.Errorf("price: got %q", got.PriceBRL)
	}
}

func TestFirecrawlClientScrapeProduct_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error": "invalid api key"}`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("k", server.URL)
	_, err := c.ScrapeProduct(context.Background(), "https://x.com/p")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", ae.Code)
	}
	if !strings.Contains(ae.Message, "Não conseguimos buscar dados") {
		t.Errorf("unexpected message: %q", ae.Message)
	}
}

func TestFirecrawlClientScrapeProduct_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `not json {`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("k", server.URL)
	_, err := c.ScrapeProduct(context.Background(), "https://x.com/p")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", ae.Code)
	}
}

func TestFirecrawlClientScrapeProduct_APIErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success": false, "error": "URL not reachable"}`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("k", server.URL)
	_, err := c.ScrapeProduct(context.Background(), "https://x.com/p")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", ae.Code)
	}
}

func TestFirecrawlClientScrapeProduct_EmptyExtraction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"success": true, "data": {}}`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("k", server.URL)
	got, err := c.ScrapeProduct(context.Background(), "https://x.com/p")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("expected empty struct, got nil")
	}
	if got.Name != "" || got.PriceBRL != "" || got.ImageURL != "" {
		t.Errorf("expected zero values, got %+v", got)
	}
}

func TestFirecrawlClientScrapeProduct_NoAuthHeaderWhenKeyEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "" {
			t.Errorf("expected no Authorization header, got %q", got)
		}
		_, _ = io.WriteString(w, `{"success": true, "data": {"json": {"name": "X"}}}`)
	}))
	defer server.Close()

	c := NewFirecrawlClient("", server.URL)
	if _, err := c.ScrapeProduct(context.Background(), "https://x.com/p"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewFirecrawlClientTrimsTrailingSlash(t *testing.T) {
	c := NewFirecrawlClient("k", "https://example.com/")
	if c.baseURL != "https://example.com" {
		t.Errorf("expected trailing slash trimmed, got %q", c.baseURL)
	}
}
