package user

import "context"

type Repository interface {
	GetByURACF(ctx context.Context, uracf string) (*User, error)
	GetByGuestID(ctx context.Context, guestID int64) (*User, error)
	Create(ctx context.Context, u *User) (*User, error)
	List(ctx context.Context) ([]UserListItem, error)
}
