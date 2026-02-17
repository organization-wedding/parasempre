package guest

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

var (
	ErrNotFound   = errors.New("guest not found")
	ErrValidation = errors.New("validation error")
)

// UserChecker verifies whether a RACF belongs to a registered user.
type UserChecker interface {
	UserExistsByURACF(ctx context.Context, uracf string) (bool, error)
}

type Service struct {
	repo        Repository
	userChecker UserChecker
}

func NewService(repo Repository, userChecker UserChecker) *Service {
	return &Service{repo: repo, userChecker: userChecker}
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

	exists, err := s.userChecker.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if !exists {
		return nil, errors.New("user-racf does not match any registered user")
	}

	existing, err := s.repo.GetByName(ctx, input.FirstName, input.LastName)
	if err != nil {
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("a guest named '%s %s' already exists", input.FirstName, input.LastName)
	}

	if input.Phone != "" {
		existingByPhone, err := s.repo.GetByPhone(ctx, input.Phone)
		if err != nil {
			return nil, fmt.Errorf("failed to check phone uniqueness: %w", err)
		}
		if existingByPhone != nil {
			return nil, fmt.Errorf("a guest with phone '%s' already exists", input.Phone)
		}
	}

	return s.repo.Create(ctx, input, userRACF)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	if err := validateUpdate(input); err != nil {
		return nil, err
	}

	exists, err := s.userChecker.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if !exists {
		return nil, errors.New("user-racf does not match any registered user")
	}

	if input.Phone != nil && *input.Phone != "" {
		existingByPhone, err := s.repo.GetByPhone(ctx, *input.Phone)
		if err != nil {
			return nil, fmt.Errorf("failed to check phone uniqueness: %w", err)
		}
		if existingByPhone != nil && existingByPhone.ID != id {
			return nil, fmt.Errorf("a guest with phone '%s' already exists", *input.Phone)
		}
	}

	if input.FirstName != nil || input.LastName != nil {
		current, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		firstName := current.FirstName
		lastName := current.LastName
		if input.FirstName != nil {
			firstName = *input.FirstName
		}
		if input.LastName != nil {
			lastName = *input.LastName
		}
		existing, err := s.repo.GetByName(ctx, firstName, lastName)
		if err != nil {
			return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, fmt.Errorf("a guest named '%s %s' already exists", firstName, lastName)
		}
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
	if input.FamilyGroup == nil || *input.FamilyGroup <= 0 {
		return errors.New("family_group is required and must be greater than 0")
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
