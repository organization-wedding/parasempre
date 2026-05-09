package giftmessage

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	Create(ctx context.Context, in CreateRow) (*GiftMessage, error)
	GetByID(ctx context.Context, id int64) (*GiftMessage, error)
	GetByTransactionID(ctx context.Context, txID int64) (*GiftMessage, error)
	ListByGift(ctx context.Context, giftID int64, limit, offset int) ([]GiftMessage, int, error)
	ListAll(ctx context.Context, limit, offset int) ([]GiftMessage, int, error)
	SoftDelete(ctx context.Context, id, byUserID int64) error
}

type TxAwareRepository interface {
	Repository
	WithTx(tx pgx.Tx) Repository
}
