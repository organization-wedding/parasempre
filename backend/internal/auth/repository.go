package auth

import (
	"context"
	"time"
)

type OTPRecord struct {
	ID        int64
	Phone     string
	Code      string
	ExpiresAt time.Time
	Used      bool
}

type OTPRepository interface {
	Create(ctx context.Context, phone, code string, expiresAt time.Time) error
	FindValid(ctx context.Context, phone, code string) (*OTPRecord, error)
	MarkUsed(ctx context.Context, id int64) error
}
