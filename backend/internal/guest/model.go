package guest

import "time"

type Guest struct {
	ID             string    `json:"id"`
	Nome           string    `json:"nome"`
	Sobrenome      string    `json:"sobrenome"`
	Telefone       string    `json:"telefone"`
	Relacionamento string    `json:"relacionamento"`
	Confirmacao    bool      `json:"confirmacao"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateGuestInput struct {
	Nome           string `json:"nome"`
	Sobrenome      string `json:"sobrenome"`
	Telefone       string `json:"telefone"`
	Relacionamento string `json:"relacionamento"`
}

type UpdateGuestInput struct {
	Nome           *string `json:"nome"`
	Sobrenome      *string `json:"sobrenome"`
	Telefone       *string `json:"telefone"`
	Relacionamento *string `json:"relacionamento"`
	Confirmacao    *bool   `json:"confirmacao"`
}
