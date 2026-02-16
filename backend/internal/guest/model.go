package guest

import "time"

type Guest struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Phone        *string   `json:"phone"`
	Relationship string    `json:"relationship"`
	Confirmed    bool      `json:"confirmed"`
	FamilyGroup  int64     `json:"family_group"`
	CreatedBy    string    `json:"created_by"`
	UpdatedBy    string    `json:"updated_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateGuestInput struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Relationship string `json:"relationship"`
	FamilyGroup  int64  `json:"family_group"`
}

type UpdateGuestInput struct {
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	Phone        *string `json:"phone"`
	Relationship *string `json:"relationship"`
	Confirmed    *bool   `json:"confirmed"`
	FamilyGroup  *int64  `json:"family_group"`
}
