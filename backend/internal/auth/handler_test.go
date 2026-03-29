package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockUserFinder struct {
	findFn func(ctx context.Context, phone string) (int64, string, string, error)
}

func (m *mockUserFinder) FindOrCreateByPhone(ctx context.Context, phone string) (int64, string, string, error) {
	return m.findFn(ctx, phone)
}

type mockPhoneChecker struct {
	checkFn func(ctx context.Context, phone string) (bool, error)
}

func (m *mockPhoneChecker) PhoneExists(ctx context.Context, phone string) (bool, error) {
	return m.checkFn(ctx, phone)
}

type mockLoginRecorder struct {
	recordFn func(ctx context.Context, userID int64)
}

func (m *mockLoginRecorder) RecordLogin(ctx context.Context, userID int64) {
	if m.recordFn != nil {
		m.recordFn(ctx, userID)
	}
}

func newTestHandler() *Handler {
	otpRepo := &mockOTPRepo{
		createFn: func(ctx context.Context, phone, code string, expiresAt time.Time) error {
			return nil
		},
		findValidFn: func(ctx context.Context, phone, code string) (*OTPRecord, error) {
			if code == "123456" {
				return &OTPRecord{ID: 1, Phone: phone, Code: code, ExpiresAt: time.Now().Add(5 * time.Minute)}, nil
			}
			return nil, nil
		},
		markUsedFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	sender := &mockSender{
		sendFn: func(phone, message string) error { return nil },
	}

	otpSvc := NewOTPService(otpRepo, sender)
	jwtSvc := NewJWTService("test-secret", 1*time.Hour)

	userFinder := &mockUserFinder{
		findFn: func(ctx context.Context, phone string) (int64, string, string, error) {
			return 1, "USR01", "guest", nil
		},
	}
	phoneCheck := &mockPhoneChecker{
		checkFn: func(ctx context.Context, phone string) (bool, error) {
			return true, nil
		},
	}
	loginRecorder := &mockLoginRecorder{}

	return NewHandler(otpSvc, jwtSvc, userFinder, phoneCheck, loginRecorder)
}

func TestHandleSendOTP(t *testing.T) {
	h := newTestHandler()

	body := `{"phone":"11999999999"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/otp/send", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleSendOTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSendOTPInvalidPhone(t *testing.T) {
	h := newTestHandler()

	body := `{"phone":"abc"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/otp/send", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleSendOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSendOTPPhoneNotFound(t *testing.T) {
	h := newTestHandler()
	h.phoneCheck = &mockPhoneChecker{
		checkFn: func(ctx context.Context, phone string) (bool, error) {
			return false, nil
		},
	}

	body := `{"phone":"11999999999"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/otp/send", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleSendOTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVerifyOTP(t *testing.T) {
	h := newTestHandler()

	body := `{"phone":"11999999999","code":"123456"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleVerifyOTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp TokenResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
	if resp.URACF != "USR01" {
		t.Fatalf("expected URACF USR01, got %q", resp.URACF)
	}
	if resp.Role != "guest" {
		t.Fatalf("expected role guest, got %q", resp.Role)
	}
}

func TestHandleVerifyOTPInvalidCode(t *testing.T) {
	h := newTestHandler()

	body := `{"phone":"11999999999","code":"000000"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleVerifyOTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleVerifyOTPMissingCode(t *testing.T) {
	h := newTestHandler()

	body := `{"phone":"11999999999"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/otp/verify", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.HandleVerifyOTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
