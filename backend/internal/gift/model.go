package gift

import "time"

type Gift struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	PriceCents  int64      `json:"price_cents"`
	ImageURL    *string    `json:"image_url,omitempty"`
	StoreURL    *string    `json:"store_url,omitempty"`
	Status      string     `json:"status"`
	DedupeKey   string     `json:"dedupe_key"`
	CreatedBy   string     `json:"created_by"`
	UpdatedBy   string     `json:"updated_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateGiftInput struct {
	Name        string  `json:"name"        validate:"required,min=1,max=200"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	PriceCents  int64   `json:"price_cents" validate:"required,gt=0"`
	ImageURL    *string `json:"image_url"   validate:"omitempty,url,startswith=https://"`
	StoreURL    *string `json:"store_url"   validate:"omitempty,url,startswith=https://"`
	Status      *string `json:"status"      validate:"omitempty,giftstatus"`
}

type UpdateGiftInput struct {
	Name        *string `json:"name"        validate:"omitempty,min=1,max=200"`
	Description *string `json:"description" validate:"omitempty,max=2000"`
	PriceCents  *int64  `json:"price_cents" validate:"omitempty,gt=0"`
	ImageURL    *string `json:"image_url"   validate:"omitempty,url,startswith=https://"`
	StoreURL    *string `json:"store_url"   validate:"omitempty,url,startswith=https://"`
	Status      *string `json:"status"      validate:"omitempty,giftstatus"`
}

type PagedResponse struct {
	Data  []Gift `json:"data"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Total int    `json:"total"`
}

type PublicGift struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	PriceCents  int64     `json:"price_cents"`
	ImageURL    *string   `json:"image_url,omitempty"`
	StoreURL    *string   `json:"store_url,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (g Gift) ToPublic() PublicGift {
	return PublicGift{
		ID:          g.ID,
		Name:        g.Name,
		Description: g.Description,
		PriceCents:  g.PriceCents,
		ImageURL:    g.ImageURL,
		StoreURL:    g.StoreURL,
		Status:      g.Status,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

type PublicPagedResponse struct {
	Data  []PublicGift `json:"data"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
	Total int          `json:"total"`
}
