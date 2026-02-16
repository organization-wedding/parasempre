package guest

import (
	"context"
	"errors"
)

var (
	ErrNotFound   = errors.New("guest not found")
	ErrValidation = errors.New("validation error")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]Guest, error) {
	return s.repo.List(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*Guest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateGuestInput) (*Guest, error) {
	if err := validateCreate(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateGuestInput) (*Guest, error) {
	if err := validateUpdate(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func validateCreate(input CreateGuestInput) error {
	if input.Nome == "" {
		return errors.New("nome is required")
	}
	if input.Sobrenome == "" {
		return errors.New("sobrenome is required")
	}
	if input.Telefone == "" {
		return errors.New("telefone is required")
	}
	if input.Relacionamento != "noivo" && input.Relacionamento != "noiva" {
		return errors.New("relacionamento must be 'noivo' or 'noiva'")
	}
	return nil
}

func validateUpdate(input UpdateGuestInput) error {
	if input.Relacionamento != nil && *input.Relacionamento != "noivo" && *input.Relacionamento != "noiva" {
		return errors.New("relacionamento must be 'noivo' or 'noiva'")
	}
	return nil
}
