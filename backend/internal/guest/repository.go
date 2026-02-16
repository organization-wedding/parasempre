package guest

import "context"

type Repository interface {
	List(ctx context.Context) ([]Guest, error)
	GetByID(ctx context.Context, id int64) (*Guest, error)
	GetByPhone(ctx context.Context, phone string) (*Guest, error)
	Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error)
	Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error)
	Delete(ctx context.Context, id int64) error
}
