package guest

import (
	"context"
	"errors"
	"regexp"
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

func (s *Service) GetByID(ctx context.Context, id int64) (*Guest, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
	if err := validateCreate(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input, userRACF)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	if err := validateUpdate(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input, userRACF)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

var phoneRegex = regexp.MustCompile(`^\d{2}9\d{8}$`)

func validateCreate(input CreateGuestInput) error {
	if input.FirstName == "" {
		return errors.New("first_name is required")
	}
	if input.LastName == "" {
		return errors.New("last_name is required")
	}
	if input.Phone != "" && !phoneRegex.MatchString(input.Phone) {
		return errors.New("phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)")
	}
	if input.Relationship != "P" && input.Relationship != "R" {
		return errors.New("relationship must be 'P' or 'R'")
	}
	return nil
}

func validateUpdate(input UpdateGuestInput) error {
	if input.Phone != nil && *input.Phone != "" && !phoneRegex.MatchString(*input.Phone) {
		return errors.New("phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)")
	}
	if input.Relationship != nil && *input.Relationship != "P" && *input.Relationship != "R" {
		return errors.New("relationship must be 'P' or 'R'")
	}
	return nil
}
