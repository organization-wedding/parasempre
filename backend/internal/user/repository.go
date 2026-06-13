package user

import "context"

type Repository interface {
	GetByURACF(ctx context.Context, uracf string) (*User, error)
	GetByGuestID(ctx context.Context, guestID int64) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByRole(ctx context.Context, role string) (*User, error)
	GetMeByURACF(ctx context.Context, uracf string) (*MeResponse, error)
	Create(ctx context.Context, u *User) (*User, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*User, error)
	Delete(ctx context.Context, id int64) error
	UnlinkGuestID(ctx context.Context, guestID int64) error
	List(ctx context.Context) ([]UserListItem, error)
	UpdateLastLogin(ctx context.Context, userID int64) error
	LogAction(ctx context.Context, userID int64, action string, details map[string]any) error
}
