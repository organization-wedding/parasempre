package user

import "time"

type User struct {
	ID        int64     `json:"id"`
	GuestID   *int64    `json:"guest_id,omitempty"`
	Role      string    `json:"role"`
	URACF     string    `json:"uracf"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterInput struct {
	Phone string `json:"phone"`
	URACF string `json:"uracf"`
}

type CheckResponse struct {
	Exists bool   `json:"exists"`
	Role   string `json:"role,omitempty"`
}

type UserListItem struct {
	URACF     string `json:"uracf"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
