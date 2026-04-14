package guest

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	List(ctx context.Context, limit, offset int) ([]Guest, int, error)
	GetByID(ctx context.Context, id int64) (*Guest, error)
	GetByName(ctx context.Context, firstName, lastName string) (*Guest, error)
	FamilyGroupExists(ctx context.Context, familyGroup int64) (bool, error)
	GetNextFamilyGroup(ctx context.Context) (int64, error)
	Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error)
	Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error)
	Delete(ctx context.Context, id int64) error
	SetConfirmed(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error)
}

// TxAwareRepository extends Repository with transaction support.
type TxAwareRepository interface {
	Repository
	WithTx(tx pgx.Tx) Repository
}
