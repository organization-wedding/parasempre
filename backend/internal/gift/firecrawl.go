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
		http:    &http.Client{Timeout: 25 * time.Second},
	}
}

const firecrawlExtractionPrompt = "Extrair informações do produto principal desta página de e-commerce. " +
	"Retornar nome do produto, preço em BRL como string (ex: '1234,56'), descrição curta e URL absoluta da imagem principal. " +
	"Se algum campo não estiver disponível, retornar string vazia."

var firecrawlExtractionSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"name":        map[string]any{"type": "string"},
		"price_brl":   map[string]any{"type": "string"},
		"description": map[string]any{"type": "string"},
		"image_url":   map[string]any{"type": "string"},
	},
	"required": []string{"name"},
}

type firecrawlRequest struct {
	URL         string             `json:"url"`
	Formats     []string           `json:"formats"`
	JSONOptions firecrawlJSONOpts  `json:"jsonOptions"`
	Timeout     int                `json:"timeout,omitempty"`
}

type firecrawlJSONOpts struct {
	Schema map[string]any `json:"schema"`
	Prompt string         `json:"prompt"`
}

type firecrawlResponse struct {
	Success bool                  `json:"success"`
	Data    firecrawlResponseData `json:"data"`
	Error   string                `json:"error,omitempty"`
}

type firecrawlResponseData struct {
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
		Formats: []string{"json"},
		JSONOptions: firecrawlJSONOpts{
			Schema: firecrawlExtractionSchema,
			Prompt: firecrawlExtractionPrompt,
		},
		Timeout: 20000,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, apperror.Internal("failed to marshal Firecrawl payload", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/scrape", bytes.NewReader(body))
	if err != nil {
		return nil, apperror.Internal("failed to create Firecrawl request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
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

	extraction := parsed.Data.JSON
	if extraction == nil {
		extraction = parsed.Data.LLMExtraction
	}
	if extraction == nil {
		slog.Warn("firecrawl: no extraction in response", "url", url, "warning", parsed.Data.Warning)
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
