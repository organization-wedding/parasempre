package auth

import (
	"context"
	"testing"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type mockOTPRepo struct {
	createFn    func(ctx context.Context, phone, code string, expiresAt time.Time) error
	findValidFn func(ctx context.Context, phone, code string) (*OTPRecord, error)
	markUsedFn  func(ctx context.Context, id int64) error
}

func (m *mockOTPRepo) Create(ctx context.Context, phone, code string, expiresAt time.Time) error {
	return m.createFn(ctx, phone, code, expiresAt)
}

func (m *mockOTPRepo) FindValid(ctx context.Context, phone, code string) (*OTPRecord, error) {
	return m.findValidFn(ctx, phone, code)
}

func (m *mockOTPRepo) MarkUsed(ctx context.Context, id int64) error {
	return m.markUsedFn(ctx, id)
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
	var markedID int64
	repo := &mockOTPRepo{
		findValidFn: func(ctx context.Context, phone, code string) (*OTPRecord, error) {
			return &OTPRecord{ID: 1, Phone: phone, Code: code, ExpiresAt: time.Now().Add(5 * time.Minute)}, nil
		},
		markUsedFn: func(ctx context.Context, id int64) error {
			markedID = id
			return nil
		},
	}

	svc := NewOTPService(repo, nil)
	err := svc.VerifyOTP(context.Background(), "11999999999", "123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if markedID != 1 {
		t.Fatalf("expected mark used id 1, got %d", markedID)
	}
}

func TestVerifyOTPInvalid(t *testing.T) {
	repo := &mockOTPRepo{
		findValidFn: func(ctx context.Context, phone, code string) (*OTPRecord, error) {
			return nil, nil
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
