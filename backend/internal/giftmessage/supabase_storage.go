package giftmessage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SupabaseStorage struct {
	baseURL    string
	bucket     string
	serviceKey string
	httpClient *http.Client
}

func NewSupabaseStorage(baseURL, bucket, serviceKey string) *SupabaseStorage {
	return &SupabaseStorage{
		baseURL:    strings.TrimRight(baseURL, "/"),
		bucket:     bucket,
		serviceKey: serviceKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SupabaseStorage) Upload(ctx context.Context, key, mime string, r io.Reader, size int64) error {
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, escapeKey(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, r)
	if err != nil {
		return fmt.Errorf("supabase storage: build upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", mime)
	req.Header.Set("x-upsert", "false")
	if size > 0 {
		req.ContentLength = size
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("supabase storage: upload http error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("supabase storage: upload status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (s *SupabaseStorage) SignURLs(ctx context.Context, keys []string, ttl time.Duration) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	if len(keys) == 0 {
		return out, nil
	}
	endpoint := fmt.Sprintf("%s/storage/v1/object/sign/%s", s.baseURL, s.bucket)
	body, err := json.Marshal(map[string]any{
		"paths":     keys,
		"expiresIn": int(ttl.Seconds()),
	})
	if err != nil {
		return nil, fmt.Errorf("supabase storage: marshal sign body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("supabase storage: build sign request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase storage: sign http error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("supabase storage: sign status %d: %s", resp.StatusCode, string(raw))
	}
	var items []struct {
		Path         string `json:"path"`
		SignedURL    string `json:"signedURL"`
		Error        string `json:"error"`
		ErrorMessage string `json:"errorMessage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("supabase storage: decode sign response: %w", err)
	}
	for _, it := range items {
		if it.SignedURL == "" {
			slog.Warn("supabase storage: sign returned empty URL", "path", it.Path, "error", it.Error, "message", it.ErrorMessage)
			continue
		}
		// signedURL vem como /object/sign/{bucket}/{path}?token=...
		out[it.Path] = s.baseURL + "/storage/v1" + it.SignedURL
	}
	return out, nil
}

func (s *SupabaseStorage) Delete(ctx context.Context, key string) error {
	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, escapeKey(key))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("supabase storage: build delete request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("supabase storage: delete http error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode != http.StatusNotFound {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("supabase storage: delete status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func escapeKey(key string) string {
	parts := strings.Split(key, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}
