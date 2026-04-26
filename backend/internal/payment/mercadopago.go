package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

const (
	mpHTTPTimeout     = 25 * time.Second
	mpReplayWindowMS  = 5 * 60 * 1000
	mpEpochSecondsLen = 10
)

type PaymentGateway interface {
	CreatePayment(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error)
	GetPayment(ctx context.Context, mpPaymentID string) (*MPPayment, error)
	VerifyWebhookSignature(headers http.Header, dataID string) bool
}

type MercadoPagoClient struct {
	accessToken   string
	webhookSecret string
	baseURL       string
	notifyURL     string
	http          *http.Client
}

func NewMercadoPagoClient(accessToken, baseURL, webhookSecret, notifyURL string) *MercadoPagoClient {
	if baseURL == "" {
		baseURL = "https://api.mercadopago.com"
	}
	return &MercadoPagoClient{
		accessToken:   accessToken,
		webhookSecret: webhookSecret,
		baseURL:       strings.TrimRight(baseURL, "/"),
		notifyURL:     notifyURL,
		http:          &http.Client{Timeout: mpHTTPTimeout},
	}
}

type CreatePaymentRequest struct {
	TransactionAmount float64 `json:"transaction_amount"`
	Token             string  `json:"token,omitempty"`
	Description       string  `json:"description,omitempty"`
	Installments      int     `json:"installments,omitempty"`
	PaymentMethodID   string  `json:"payment_method_id"`
	IssuerID          string  `json:"issuer_id,omitempty"`
	Payer             MPPayer `json:"payer"`
	ExternalReference string  `json:"external_reference,omitempty"`
	NotificationURL   string  `json:"notification_url,omitempty"`
}

type MPPayer struct {
	Email          string           `json:"email"`
	Identification MPIdentification `json:"identification"`
}

type MPIdentification struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type MPPayment struct {
	ID                 int64                 `json:"id"`
	Status             string                `json:"status"`
	StatusDetail       string                `json:"status_detail"`
	PaymentMethodID    string                `json:"payment_method_id"`
	PaymentTypeID      string                `json:"payment_type_id"`
	TransactionAmount  float64               `json:"transaction_amount"`
	ExternalReference  string                `json:"external_reference"`
	PointOfInteraction *MPPointOfInteraction `json:"point_of_interaction,omitempty"`
}

type MPPointOfInteraction struct {
	TransactionData *MPTransactionData `json:"transaction_data,omitempty"`
}

type MPTransactionData struct {
	QRCode       string `json:"qr_code"`
	QRCodeBase64 string `json:"qr_code_base64"`
	TicketURL    string `json:"ticket_url"`
}

func (c *MercadoPagoClient) CreatePayment(ctx context.Context, payload CreatePaymentRequest, idempotencyKey string) (*MPPayment, error) {
	if c.notifyURL != "" && payload.NotificationURL == "" {
		payload.NotificationURL = c.notifyURL
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, apperror.Internal("failed to marshal MP payment payload", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/payments", bytes.NewReader(body))
	if err != nil {
		return nil, apperror.Internal("failed to create MP payment request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", idempotencyKey)

	resp, err := c.http.Do(req)
	if err != nil {
		slog.Error("mercadopago: create payment request failed", "error", err)
		return nil, apperror.ServiceUnavailable("Falha ao contactar Mercado Pago. Tente novamente.")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperror.Internal("failed to read MP response", err)
	}

	if resp.StatusCode >= 500 {
		slog.Error("mercadopago: 5xx from API", "status", resp.StatusCode)
		return nil, apperror.ServiceUnavailable("Mercado Pago indisponível. Tente novamente em instantes.")
	}

	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
		msg := extractMPError(respBody)
		slog.Warn("mercadopago: validation error", "status", resp.StatusCode, "message", msg)
		return nil, apperror.Validation(msg)
	}

	var parsed MPPayment
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		slog.Error("mercadopago: failed to parse response", "error", err, "status", resp.StatusCode, "body_len", len(respBody))
		return nil, apperror.Internal("Resposta inválida do Mercado Pago.", err)
	}

	if parsed.ID == 0 {
		slog.Error("mercadopago: response missing id", "status", resp.StatusCode)
		return nil, apperror.Internal("Mercado Pago retornou resposta sem ID.", nil)
	}

	return &parsed, nil
}

func (c *MercadoPagoClient) GetPayment(ctx context.Context, mpPaymentID string) (*MPPayment, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/payments/"+mpPaymentID, nil)
	if err != nil {
		return nil, apperror.Internal("failed to create MP get request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.http.Do(req)
	if err != nil {
		slog.Error("mercadopago: get payment request failed", "id", mpPaymentID, "error", err)
		return nil, apperror.ServiceUnavailable("Falha ao consultar Mercado Pago.")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperror.Internal("failed to read MP get response", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, apperror.NotFound("Pagamento não encontrado no Mercado Pago.")
	}
	if resp.StatusCode >= 400 {
		slog.Error("mercadopago: get payment bad response", "status", resp.StatusCode)
		return nil, apperror.ServiceUnavailable("Erro ao consultar Mercado Pago.")
	}

	var parsed MPPayment
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, apperror.Internal("Resposta inválida do Mercado Pago.", err)
	}
	return &parsed, nil
}

func (c *MercadoPagoClient) VerifyWebhookSignature(headers http.Header, dataID string) bool {
	if c.webhookSecret == "" {
		slog.Warn("mercadopago: webhook secret not configured, rejecting")
		return false
	}
	if dataID == "" {
		return false
	}

	signature := headers.Get("x-signature")
	requestID := headers.Get("x-request-id")
	if signature == "" || requestID == "" {
		return false
	}

	ts, hash := parseSignatureHeader(signature)
	if ts == "" || hash == "" {
		return false
	}

	if !timestampWithinWindow(ts, time.Now(), mpReplayWindowMS) {
		slog.Warn("mercadopago: webhook ts outside replay window", "ts", ts)
		return false
	}

	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", dataID, requestID, ts)
	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(hash))
}

func parseSignatureHeader(header string) (ts string, hash string) {
	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		switch key {
		case "ts":
			ts = val
		case "v1":
			hash = val
		}
	}
	return ts, hash
}

func timestampWithinWindow(tsRaw string, now time.Time, windowMS int64) bool {
	tsInt, err := strconv.ParseInt(tsRaw, 10, 64)
	if err != nil {
		return false
	}
	var tsMS int64
	if len(tsRaw) <= mpEpochSecondsLen {
		tsMS = tsInt * 1000
	} else {
		tsMS = tsInt
	}
	nowMS := now.UnixMilli()
	diff := nowMS - tsMS
	if diff < 0 {
		diff = -diff
	}
	return diff <= windowMS
}

func AmountFromCents(cents int64) float64 {
	return float64(cents) / 100.0
}

func CentsFromAmount(amount float64) int64 {
	return int64(math.Round(amount * 100))
}

type mpErrorResponse struct {
	Message string         `json:"message"`
	Error   string         `json:"error"`
	Status  int            `json:"status"`
	Cause   []mpErrorCause `json:"cause"`
}

type mpErrorCause struct {
	Code        any    `json:"code"`
	Description string `json:"description"`
}

func extractMPError(body []byte) string {
	var parsed mpErrorResponse
	if err := json.Unmarshal(body, &parsed); err == nil {
		if len(parsed.Cause) > 0 && parsed.Cause[0].Description != "" {
			return parsed.Cause[0].Description
		}
		if parsed.Message != "" {
			return parsed.Message
		}
	}
	return "Pagamento rejeitado pelo Mercado Pago."
}
