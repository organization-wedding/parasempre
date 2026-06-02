package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testWebhookSecret = "whsec_test_super_secret_value"

func sign(t *testing.T, secret, manifest string) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(manifest))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature_HappyPath(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: testWebhookSecret}
	dataID := "12345"
	requestID := "req-abc-001"
	tsMS := time.Now().UnixMilli()
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%d;", dataID, requestID, tsMS)
	hash := sign(t, testWebhookSecret, manifest)

	hdr := http.Header{}
	hdr.Set("x-signature", fmt.Sprintf("ts=%d,v1=%s", tsMS, hash))
	hdr.Set("x-request-id", requestID)

	if !c.VerifyWebhookSignature(hdr, dataID) {
		t.Fatal("expected signature to verify")
	}
}

func TestVerifyWebhookSignature_AcceptsTimestampInSeconds(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: testWebhookSecret}
	dataID := "12345"
	requestID := "req-001"
	tsSec := time.Now().Unix()
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%d;", dataID, requestID, tsSec)
	hash := sign(t, testWebhookSecret, manifest)

	hdr := http.Header{}
	hdr.Set("x-signature", fmt.Sprintf("ts=%d,v1=%s", tsSec, hash))
	hdr.Set("x-request-id", requestID)

	if !c.VerifyWebhookSignature(hdr, dataID) {
		t.Fatal("expected signature to verify with seconds-precision ts")
	}
}

func TestVerifyWebhookSignature_RejectsTamperedHash(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: testWebhookSecret}
	dataID := "12345"
	requestID := "req-002"
	tsMS := time.Now().UnixMilli()

	hdr := http.Header{}
	hdr.Set("x-signature", fmt.Sprintf("ts=%d,v1=%s", tsMS, strings.Repeat("0", 64)))
	hdr.Set("x-request-id", requestID)

	if c.VerifyWebhookSignature(hdr, dataID) {
		t.Fatal("expected tampered signature to fail")
	}
}

func TestVerifyWebhookSignature_RejectsExpiredTimestamp(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: testWebhookSecret}
	dataID := "12345"
	requestID := "req-003"
	oldTS := time.Now().Add(-10 * time.Minute).UnixMilli()
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%d;", dataID, requestID, oldTS)
	hash := sign(t, testWebhookSecret, manifest)

	hdr := http.Header{}
	hdr.Set("x-signature", fmt.Sprintf("ts=%d,v1=%s", oldTS, hash))
	hdr.Set("x-request-id", requestID)

	if c.VerifyWebhookSignature(hdr, dataID) {
		t.Fatal("expected old timestamp to be rejected")
	}
}

func TestVerifyWebhookSignature_RejectsMissingHeaders(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: testWebhookSecret}
	if c.VerifyWebhookSignature(http.Header{}, "12345") {
		t.Fatal("expected missing x-signature to fail")
	}

	hdr := http.Header{}
	hdr.Set("x-signature", "ts=12345,v1=abc")
	if c.VerifyWebhookSignature(hdr, "12345") {
		t.Fatal("expected missing x-request-id to fail")
	}
}

func TestVerifyWebhookSignature_RejectsEmptyDataID(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: testWebhookSecret}
	hdr := http.Header{}
	hdr.Set("x-signature", "ts=12345,v1=abc")
	hdr.Set("x-request-id", "req-x")
	if c.VerifyWebhookSignature(hdr, "") {
		t.Fatal("expected empty data.id to fail")
	}
}

func TestVerifyWebhookSignature_RejectsWhenSecretMissing(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: ""}
	hdr := http.Header{}
	hdr.Set("x-signature", "ts=12345,v1=abc")
	hdr.Set("x-request-id", "req-x")
	if c.VerifyWebhookSignature(hdr, "12345") {
		t.Fatal("expected missing secret to fail")
	}
}

func TestVerifyWebhookSignature_DataIDMustMatch(t *testing.T) {
	c := &MercadoPagoClient{webhookSecret: testWebhookSecret}
	requestID := "req-abc"
	tsMS := time.Now().UnixMilli()
	signedID := "12345"
	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%d;", signedID, requestID, tsMS)
	hash := sign(t, testWebhookSecret, manifest)

	hdr := http.Header{}
	hdr.Set("x-signature", fmt.Sprintf("ts=%d,v1=%s", tsMS, hash))
	hdr.Set("x-request-id", requestID)

	if c.VerifyWebhookSignature(hdr, "99999") {
		t.Fatal("expected mismatched data.id to fail")
	}
}

func TestParseSignatureHeader(t *testing.T) {
	ts, hash := parseSignatureHeader("ts=1700000000,v1=abcdef0123456789")
	if ts != "1700000000" || hash != "abcdef0123456789" {
		t.Errorf("got ts=%q hash=%q", ts, hash)
	}

	ts, hash = parseSignatureHeader("  ts=1700000000 ,  v1=abc  ")
	if ts != "1700000000" || hash != "abc" {
		t.Errorf("got ts=%q hash=%q after trim", ts, hash)
	}
	ts, hash = parseSignatureHeader("ts=123")
	if hash != "" {
		t.Errorf("expected empty hash when v1 missing, got %q", hash)
	}
	if ts != "123" {
		t.Errorf("expected ts=123, got %q", ts)
	}
}

func TestTimestampWithinWindow_HandlesBothPrecisions(t *testing.T) {
	now := time.Now()
	tsMS := strconv.FormatInt(now.UnixMilli(), 10)
	if !timestampWithinWindow(tsMS, now, 60_000) {
		t.Error("expected current ms timestamp to fit window")
	}
	tsSec := strconv.FormatInt(now.Unix(), 10)
	if !timestampWithinWindow(tsSec, now, 60_000) {
		t.Error("expected current seconds timestamp to fit window")
	}
	tsOld := strconv.FormatInt(now.Add(-1*time.Hour).UnixMilli(), 10)
	if timestampWithinWindow(tsOld, now, 60_000) {
		t.Error("expected hour-old timestamp to be rejected")
	}
	if timestampWithinWindow("not-a-number", now, 60_000) {
		t.Error("expected non-numeric to be rejected")
	}
}

func TestAmountFromCents_RoundTrip(t *testing.T) {
	cases := []int64{1, 100, 199, 1234567, 19990, 100000000}
	for _, c := range cases {
		amt := AmountFromCents(c)
		got := CentsFromAmount(amt)
		if got != c {
			t.Errorf("round-trip failed: %d -> %v -> %d", c, amt, got)
		}
	}
}

func TestCreatePayment_BuildsCorrectRequest(t *testing.T) {
	var capturedBody []byte
	var capturedHeaders http.Header
	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.Path
		capturedHeaders = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		capturedBody = body
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                 555,
			"status":             "approved",
			"transaction_amount": 199.90,
		})
	}))
	defer server.Close()

	c := NewMercadoPagoClient("TEST-TOKEN", server.URL, testWebhookSecret, "https://api.example.com/webhook")
	resp, err := c.CreatePayment(context.Background(), CreatePaymentRequest{
		TransactionAmount: 199.90,
		Token:             "card_tok",
		PaymentMethodID:   "visa",
		Installments:      1,
		Description:       "test",
		Payer: MPPayer{
			Email:          "buyer@example.com",
			Identification: MPIdentification{Type: "CPF", Number: "39053344705"},
		},
		ExternalReference: "gift_tx:1",
	}, "idem-key-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != 555 || resp.Status != "approved" {
		t.Errorf("unexpected response: %+v", resp)
	}

	if capturedURL != "/v1/payments" {
		t.Errorf("expected POST /v1/payments, got %s", capturedURL)
	}
	if got := capturedHeaders.Get("X-Idempotency-Key"); got != "idem-key-abc" {
		t.Errorf("expected idempotency key header, got %q", got)
	}
	if got := capturedHeaders.Get("Authorization"); got != "Bearer TEST-TOKEN" {
		t.Errorf("expected Bearer token, got %q", got)
	}

	var sent map[string]any
	if err := json.Unmarshal(capturedBody, &sent); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if sent["notification_url"] != "https://api.example.com/webhook" {
		t.Errorf("expected notification_url to be set from client default, got %v", sent["notification_url"])
	}
	if sent["external_reference"] != "gift_tx:1" {
		t.Errorf("expected external_reference, got %v", sent["external_reference"])
	}
}

func TestCreatePayment_HandlesValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"message": "invalid",
			"cause":   []map[string]any{{"description": "Invalid card token"}},
		})
	}))
	defer server.Close()
	c := NewMercadoPagoClient("TEST-TOKEN", server.URL, testWebhookSecret, "")
	_, err := c.CreatePayment(context.Background(), CreatePaymentRequest{}, "idem-1")
	if err == nil {
		t.Fatal("expected error from 400 response")
	}
	if !strings.Contains(err.Error(), "Invalid card token") {
		t.Errorf("expected MP error message in error, got %q", err.Error())
	}
}

func TestCreatePayment_HandlesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":"down"}`))
	}))
	defer server.Close()
	c := NewMercadoPagoClient("TEST-TOKEN", server.URL, testWebhookSecret, "")
	_, err := c.CreatePayment(context.Background(), CreatePaymentRequest{}, "idem-1")
	if err == nil {
		t.Fatal("expected error from 500 response")
	}
}

func TestExtractMPError(t *testing.T) {
	body := []byte(`{"message":"validation","cause":[{"description":"Card declined"}]}`)
	if got := extractMPError(body); got != "Card declined" {
		t.Errorf("expected cause description, got %q", got)
	}

	body2 := []byte(`{"message":"validation"}`)
	if got := extractMPError(body2); got != "validation" {
		t.Errorf("expected fallback to message, got %q", got)
	}

	body3 := []byte(`not-json`)
	if got := extractMPError(body3); got == "" {
		t.Error("expected default message for non-json body")
	}
}
