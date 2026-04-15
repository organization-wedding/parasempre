package user

import "time"

type User struct {
	ID          int64      `json:"id"`
	GuestID     *int64     `json:"guest_id,omitempty"`
	Role        string     `json:"role"`
	URACF       string     `json:"uracf"`
	Phone       *string    `json:"phone,omitempty"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type RegisterInput struct {
	Phone string `json:"phone" validate:"required,brphone"`
	URACF string `json:"uracf" validate:"required,uracf"`
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

type CoupleData struct {
	URACF string
	Phone string
}

