package gift

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type ScrapedProduct struct {
	Name        string
	Description string
	PriceBRL    string
	ImageURL    string
}

type FirecrawlClient struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewFirecrawlClient(apiKey, baseURL string) *FirecrawlClient {
	if baseURL == "" {
		baseURL = "https://api.firecrawl.dev"
	}
	return &FirecrawlClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 40 * time.Second},
	}
}

var firecrawlExtractionSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"name": map[string]any{
			"type":        "string",
			"description": "Nome completo do produto principal da página",
		},
		"price_brl": map[string]any{
			"type":        "string",
			"description": "Preço atual do produto em reais (BRL), por exemplo 'R$1.234,56'",
		},
		"description": map[string]any{
			"type":        "string",
			"description": "Descrição curta do produto",
		},
		"image_url": map[string]any{
			"type":        "string",
			"description": "URL absoluta (https) da imagem principal do produto",
		},
	},
	"required": []string{"name", "price_brl"},
}

type firecrawlRequest struct {
	URL     string                  `json:"url"`
	Formats []string                `json:"formats"`
	Extract firecrawlExtractOptions `json:"extract"`
	Proxy   string                  `json:"proxy,omitempty"`
	Timeout int                     `json:"timeout,omitempty"`
}

type firecrawlExtractOptions struct {
	Schema map[string]any `json:"schema"`
}

type firecrawlResponse struct {
	Success bool                  `json:"success"`
	Data    firecrawlResponseData `json:"data"`
	Error   string                `json:"error,omitempty"`
}

type firecrawlResponseData struct {
	Extract       *firecrawlExtraction `json:"extract,omitempty"`
	JSON          *firecrawlExtraction `json:"json,omitempty"`
	LLMExtraction *firecrawlExtraction `json:"llm_extraction,omitempty"`
	Warning       string               `json:"warning,omitempty"`
}

type firecrawlExtraction struct {
	Name        string `json:"name"`
	PriceBRL    string `json:"price_brl"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
}

func (c *FirecrawlClient) ScrapeProduct(ctx context.Context, url string) (*ScrapedProduct, error) {
	payload := firecrawlRequest{
		URL:     url,
		Formats: []string{"extract"},
		Extract: firecrawlExtractOptions{
			Schema: firecrawlExtractionSchema,
		},
		Proxy:   "auto",
		Timeout: 30000,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, apperror.Internal("failed to marshal Firecrawl payload", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/scrape", bytes.NewReader(body))
	if err != nil {
		return nil, apperror.Internal("failed to create Firecrawl request", err)
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		slog.Error("firecrawl: request failed", "url", url, "error", err)
		return nil, apperror.Internal("Falha ao consultar serviço de scraping", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("firecrawl: failed to read response body", "url", url, "error", err)
		return nil, apperror.Internal("Falha ao ler resposta do serviço de scraping", err)
	}

	if resp.StatusCode >= 400 {
		slog.Error("firecrawl: bad response", "url", url, "status", resp.StatusCode, "body", truncate(string(respBody), 500))
		return nil, apperror.Validation("Não conseguimos buscar dados desta URL.")
	}

	var parsed firecrawlResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		slog.Error("firecrawl: failed to parse response", "url", url, "error", err, "body", truncate(string(respBody), 500))
		return nil, apperror.Internal("Resposta inválida do serviço de scraping", err)
	}

	if !parsed.Success && parsed.Error != "" {
		slog.Error("firecrawl: api returned error", "url", url, "error", parsed.Error)
		return nil, apperror.Validation("Não conseguimos buscar dados desta URL.")
	}

	extraction := parsed.Data.Extract
	if extraction == nil {
		extraction = parsed.Data.JSON
	}
	if extraction == nil {
		extraction = parsed.Data.LLMExtraction
	}
	if extraction == nil {
		slog.Warn("firecrawl: no extraction in response",
			"url", url,
			"warning", parsed.Data.Warning,
			"raw_body", truncate(string(respBody), 2000),
		)
		return &ScrapedProduct{}, nil
	}

	return &ScrapedProduct{
		Name:        extraction.Name,
		Description: extraction.Description,
		PriceBRL:    extraction.PriceBRL,
		ImageURL:    extraction.ImageURL,
	}, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
