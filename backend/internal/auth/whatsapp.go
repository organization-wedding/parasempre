package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type WhatsAppSender interface {
	SendMessage(phone, message string) error
}

type EvoAPISender struct {
	baseURL  string
	apiKey   string
	instance string
}

func NewEvoAPISender(baseURL, apiKey, instance string) *EvoAPISender {
	return &EvoAPISender{baseURL: baseURL, apiKey: apiKey, instance: instance}
}

func (s *EvoAPISender) SendMessage(phone, message string) error {
	payload := map[string]any{
		"number": "55" + phone,
		"text":   message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal WhatsApp payload: %w", err)
	}

	url := fmt.Sprintf("%s/message/sendText/%s", s.baseURL, s.instance)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create WhatsApp request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("whatsapp: request failed", "error", err)
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		slog.Error("whatsapp: bad response", "status", resp.StatusCode)
		return fmt.Errorf("WhatsApp API returned status %d", resp.StatusCode)
	}

	return nil
}
