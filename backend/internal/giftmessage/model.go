package giftmessage

import "time"

const (
	MediaKindImage = "image"
	MediaKindAudio = "audio"
	MediaKindVideo = "video"
)

type GiftMessage struct {
	ID                int64
	GiftTransactionID int64
	GiftID            int64
	UserID            int64
	AuthorName        string
	Content           string
	MediaObjectKey    *string
	MediaKind         *string
	MediaSizeBytes    *int64
	MediaMimeType     *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	DeletedBy         *int64
}

type CreateInput struct {
	AuthorName string `json:"author_name" validate:"required,min=1,max=120"`
	Content    string `json:"content"     validate:"required,min=1,max=500"`
}

type CreateRow struct {
	GiftTransactionID int64
	GiftID            int64
	UserID            int64
	AuthorName        string
	Content           string
	MediaObjectKey    *string
	MediaKind         *string
	MediaSizeBytes    *int64
	MediaMimeType     *string
}

type PublicMessage struct {
	ID         int64     `json:"id"`
	GiftID     int64     `json:"gift_id"`
	AuthorName string    `json:"author_name"`
	Content    string    `json:"content"`
	MediaURL   *string   `json:"media_url"`
	MediaKind  *string   `json:"media_kind"`
	CreatedAt  time.Time `json:"created_at"`
}

type AdminMessage struct {
	PublicMessage
	UserID            int64 `json:"user_id"`
	GiftTransactionID int64 `json:"gift_transaction_id"`
}

type Paged[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type TransactionSnapshot struct {
	ID     int64
	GiftID int64
	UserID int64
	Status string
}
