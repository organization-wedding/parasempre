package auth

import (
	"context"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type mockOTPRepo struct {
	createFn           func(ctx context.Context, phone, code string, expiresAt time.Time) error
	verifyAndMarkUsedFn func(ctx context.Context, phone, code string) (bool, error)
}

func (m *mockOTPRepo) Create(ctx context.Context, phone, code string, expiresAt time.Time) error {
	return m.createFn(ctx, phone, code, expiresAt)
}

func (m *mockOTPRepo) VerifyAndMarkUsed(ctx context.Context, phone, code string) (bool, error) {
	return m.verifyAndMarkUsedFn(ctx, phone, code)
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
	err := svc.SendOTP(context.Background(), "11999999999")
	if err != nil {
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
