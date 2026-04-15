package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type mockOTPRepo struct {
	createFn            func(ctx context.Context, phone, code string, expiresAt time.Time) error
	verifyAndMarkUsedFn func(ctx context.Context, phone, code string) (bool, error)
	sendCooldownFn      func(ctx context.Context, phone string) (time.Duration, error)
}

func (m *mockOTPRepo) Create(ctx context.Context, phone, code string, expiresAt time.Time) error {
	return m.createFn(ctx, phone, code, expiresAt)
}

func (m *mockOTPRepo) VerifyAndMarkUsed(ctx context.Context, phone, code string) (bool, error) {
	return m.verifyAndMarkUsedFn(ctx, phone, code)
}

func (m *mockOTPRepo) SendCooldown(ctx context.Context, phone string) (time.Duration, error) {
	if m.sendCooldownFn == nil {
		return 0, nil
	}
	return m.sendCooldownFn(ctx, phone)
}

type mockSender struct {
	sendFn func(phone, message string) error
}

func (m *mockSender) SendMessage(phone, message string) error {
	return m.sendFn(phone, message)
}

func TestSendOTP(t *testing.T) {
	var savedPhone, savedCode string
	var sentMessage string

	repo := &mockOTPRepo{
		createFn: func(ctx context.Context, phone, code string, expiresAt time.Time) error {
			savedPhone = phone
			savedCode = code
			return nil
		},
	}
	sender := &mockSender{
		sendFn: func(phone, message string) error {
			sentMessage = message
			return nil
		},
	}

	svc := NewOTPService(repo, sender)
	if err := svc.SendOTP(context.Background(), "11999999999"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if savedPhone != "11999999999" {
		t.Fatalf("expected phone 11999999999, got %q", savedPhone)
	}
	if len(savedCode) != 6 {
		t.Fatalf("expected 6-digit code, got %q", savedCode)
	}
	if sentMessage == "" {
		t.Fatal("expected message to be sent")
	}
}

func TestSendOTPRateLimited(t *testing.T) {
	repo := &mockOTPRepo{
		sendCooldownFn: func(ctx context.Context, phone string) (time.Duration, error) {
			return 40 * time.Second, nil
		},
	}
	svc := NewOTPService(repo, &mockSender{sendFn: func(phone, message string) error { return nil }})

	err := svc.SendOTP(context.Background(), "11999999999")
	if err == nil {
		t.Fatal("expected rate limit error")
	}
	var rle *apperror.RateLimitedError
	if !errors.As(err, &rle) {
		t.Fatalf("expected *apperror.RateLimitedError, got %T", err)
	}
	if rle.Code != 429 {
		t.Fatalf("expected 429, got %d", rle.Code)
	}
	if rle.RetryAfter != 40*time.Second {
		t.Fatalf("expected RetryAfter 40s, got %v", rle.RetryAfter)
	}
}

func TestVerifyOTPValid(t *testing.T) {
	repo := &mockOTPRepo{
		verifyAndMarkUsedFn: func(ctx context.Context, phone, code string) (bool, error) {
			return true, nil
		},
	}

	svc := NewOTPService(repo, nil)
	err := svc.VerifyOTP(context.Background(), "11999999999", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyOTPInvalid(t *testing.T) {
	repo := &mockOTPRepo{
		verifyAndMarkUsedFn: func(ctx context.Context, phone, code string) (bool, error) {
			return false, nil
		},
	}

	svc := NewOTPService(repo, nil)
	err := svc.VerifyOTP(context.Background(), "11999999999", "000000")
	if err == nil {
		t.Fatal("expected error for invalid OTP")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T", err)
	}
	if ae.Code != 401 {
		t.Fatalf("expected 401, got %d", ae.Code)
	}
}
