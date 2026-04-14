package auth

import (
	"context"
	"time"
)

type OTPRepository interface {
	Create(ctx context.Context, phone, code string, expiresAt time.Time) error
	VerifyAndMarkUsed(ctx context.Context, phone, code string) (bool, error)
}
