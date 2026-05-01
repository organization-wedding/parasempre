package payment

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error)
	GetByID(ctx context.Context, id int64) (*GiftTransaction, error)
	GetByMPPaymentID(ctx context.Context, mpPaymentID string) (*GiftTransaction, error)
	UpdateAfterCreate(ctx context.Context, id int64, mpPaymentID string, status string) (*GiftTransaction, error)
	UpdateStatus(ctx context.Context, mpPaymentID string, newStatus string, allowedFrom []string) (int64, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]GiftTransaction, int, error)
	ListAll(ctx context.Context, filter ListFilter, limit, offset int) ([]AdminTransactionRow, int, error)
	Summary(ctx context.Context) (*AdminSummary, error)
}

type TxAwareRepository interface {
	Repository
	WithTx(tx pgx.Tx) Repository
}
