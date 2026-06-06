package guest

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type Repository interface {
	List(ctx context.Context, limit, offset int) ([]Guest, int, error)
	ListByFamilyGroup(ctx context.Context, familyGroup int64) ([]Guest, error)
	GetByIDAny(ctx context.Context, id int64) (*Guest, error)
	GetByIDs(ctx context.Context, ids []int64) ([]Guest, error)
	GetByName(ctx context.Context, firstName, lastName string) (*Guest, error)
	FamilyGroupExists(ctx context.Context, familyGroup int64) (bool, error)
	GetNextFamilyGroup(ctx context.Context) (int64, error)
	Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error)
	Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error)
	Delete(ctx context.Context, id int64) error
	SetAttending(ctx context.Context, id int64, attending bool, userRACF string) (*Guest, error)
	SetAttendingByFamilyGroup(ctx context.Context, familyGroup int64, attending bool, userRACF string) ([]Guest, error)
	SetAttendingByIDs(ctx context.Context, ids []int64, attending bool, userRACF string) ([]Guest, error)
	GetFamilyGroupByPhone(ctx context.Context, phone string) (*int64, error)
}

type TxAwareRepository interface {
	Repository
	WithTx(tx pgx.Tx) Repository
}
