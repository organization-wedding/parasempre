package payment

import "time"

const (
	PaymentMethodCreditCard = "credit_card"
	PaymentMethodPix        = "pix"

	StatusPending   = "pending"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusRefunded  = "refunded"
	StatusCancelled = "cancelled"

	MPMethodPix = "pix"
)

var knownStatuses = map[string]bool{
	StatusPending:   true,
	StatusApproved:  true,
	StatusRejected:  true,
	StatusRefunded:  true,
	StatusCancelled: true,
}

type GiftTransaction struct {
	ID               int64     `json:"id"`
	GiftID           int64     `json:"gift_id"`
	UserID           int64     `json:"user_id"`
	PaymentMethod    string    `json:"payment_method"`
	MPPaymentID      *string   `json:"mp_payment_id,omitempty"`
	MPPreferenceID   *string   `json:"mp_preference_id,omitempty"`
	AmountCents      int64     `json:"amount_cents"`
	Status           string    `json:"status"`
	IdempotencyKey   *string   `json:"-"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	GiftNameSnapshot string    `json:"-"`
}

func (t GiftTransaction) ToPublic() PublicTransaction {
	return PublicTransaction{
		ID:            t.ID,
		GiftID:        t.GiftID,
		GiftName:      t.GiftNameSnapshot,
		PaymentMethod: t.PaymentMethod,
		Status:        t.Status,
		AmountCents:   t.AmountCents,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
	}
}

type PublicTransaction struct {
	ID            int64     `json:"id"`
	GiftID        int64     `json:"gift_id"`
	GiftName      string    `json:"gift_name"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
	AmountCents   int64     `json:"amount_cents"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Pix           *PixData  `json:"pix,omitempty"`
}

type AdminTransaction struct {
	PublicTransaction
	UserID      int64   `json:"user_id"`
	UserURACF   string  `json:"user_uracf"`
	UserPhone   *string `json:"user_phone"`
	MPPaymentID *string `json:"mp_payment_id,omitempty"`
}

type StatusBreakdown struct {
	Status     string `json:"status"`
	Count      int    `json:"count"`
	TotalCents int64  `json:"total_cents"`
}

type AdminSummary struct {
	Total              int               `json:"total"`
	ApprovedTotalCents int64             `json:"approved_total_cents"`
	ByStatus           []StatusBreakdown `json:"by_status"`
}

type PagedTransactions[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type ListFilter struct {
	Status *string
	GiftID *int64
}

type AdminTransactionRow struct {
	GiftTransaction
	UserURACF string
	UserPhone *string
}

type PayerIdentification struct {
	Type   string `json:"type"   validate:"required,eq=CPF"`
	Number string `json:"number" validate:"required,cpf"`
}

type Payer struct {
	Email          string              `json:"email"          validate:"required,email"`
	Identification PayerIdentification `json:"identification" validate:"required"`
}

type CreatePurchaseInput struct {
	PaymentMethodID string  `json:"payment_method_id" validate:"required"`
	Token           *string `json:"token"`
	IssuerID        *string `json:"issuer_id"`
	Installments    *int    `json:"installments"      validate:"omitempty,min=1,max=12"`
	Payer           Payer   `json:"payer"             validate:"required"`
	IdempotencyKey  string  `json:"idempotency_key"   validate:"required,min=8,max=64"`
}

type PixData struct {
	QRCode       string `json:"qr_code,omitempty"`
	QRCodeBase64 string `json:"qr_code_base64,omitempty"`
	TicketURL    string `json:"ticket_url,omitempty"`
}

type PurchaseResponse struct {
	TransactionID int64    `json:"transaction_id"`
	MPPaymentID   *string  `json:"mp_payment_id,omitempty"`
	Status        string   `json:"status"`
	StatusDetail  string   `json:"status_detail,omitempty"`
	PaymentMethod string   `json:"payment_method"`
	AmountCents   int64    `json:"amount_cents"`
	Pix           *PixData `json:"pix,omitempty"`
}

type CreateGiftTransactionInput struct {
	GiftID           int64
	UserID           int64
	PaymentMethod    string
	AmountCents      int64
	Status           string
	IdempotencyKey   string
	GiftNameSnapshot string
}
