package gift

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	List(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error)
	GetByID(ctx context.Context, id int64) (*Gift, error)
	Create(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error)
	Update(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error)
	Delete(ctx context.Context, id int64, userRACF string) error
	FindByDedupeKeys(ctx context.Context, keys []string) (map[string]bool, error)
}

type TxAwareRepository interface {
	Repository
	WithTx(tx pgx.Tx) Repository
}
