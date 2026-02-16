package guest

import "context"

type Repository interface {
	List(ctx context.Context) ([]Guest, error)
	GetByID(ctx context.Context, id string) (*Guest, error)
	Create(ctx context.Context, input CreateGuestInput) (*Guest, error)
	Update(ctx context.Context, id string, input UpdateGuestInput) (*Guest, error)
	Delete(ctx context.Context, id string) error
}
