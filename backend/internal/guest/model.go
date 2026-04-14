package guest

import "time"

type Guest struct {
	ID           int64     `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Relationship string    `json:"relationship"`
	Confirmed    bool      `json:"confirmed"`
	FamilyGroup  int64     `json:"family_group"`
	CreatedBy    string    `json:"created_by"`
	UpdatedBy    string    `json:"updated_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateGuestInput struct {
	FirstName    string  `json:"first_name"   validate:"required"`
	LastName     string  `json:"last_name"    validate:"required"`
	Relationship string  `json:"relationship" validate:"required,relationship"`
	FamilyGroup  *int64  `json:"family_group" validate:"omitempty,gt=0"`
	Phone        *string `json:"phone"        validate:"omitempty,brphone"`
}

type UpdateGuestInput struct {
	FirstName    *string `json:"first_name"`
	LastName     *string `json:"last_name"`
	Relationship *string `json:"relationship" validate:"omitempty,relationship"`
	Confirmed    *bool   `json:"confirmed"`
	FamilyGroup  *int64  `json:"family_group"`
}

type PagedResponse struct {
	Data  []Guest `json:"data"`
	Page  int     `json:"page"`
	Limit int     `json:"limit"`
	Total int     `json:"total"`
}

type ImportRowError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

type ImportResponse struct {
	SuccessCount int              `json:"success_count"`
	ErrorCount   int              `json:"error_count"`
	Total        int              `json:"total"`
	Errors       []ImportRowError `json:"errors"`
}
