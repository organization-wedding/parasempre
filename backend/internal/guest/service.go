package guest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	guests, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("guest.service list: failed", "error", err)
		return nil, err
	}
	return guests, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Guest, error) {
	guest, err := s.repo.GetByID(ctx, id)
	if err != nil {
		slog.Error("guest.service get_by_id: failed", "id", id, "error", err)
		return nil, err
	}
	return guest, nil
}

func (s *Service) Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
	if err := validateCreate(input); err != nil {
		slog.Warn("guest.service create: validation failed", "first_name", input.FirstName, "last_name", input.LastName, "error", err)
		return nil, err
	}

	exists, err := s.userChecker.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		slog.Error("guest.service create: user check failed", "user_racf", userRACF, "error", err)
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if !exists {
		slog.Warn("guest.service create: unknown user racf", "user_racf", userRACF)
		return nil, errors.New("usuário não autorizado para realizar esta operação")
	}

	existing, err := s.repo.GetByName(ctx, input.FirstName, input.LastName)
	if err != nil {
		slog.Error("guest.service create: name lookup failed", "first_name", input.FirstName, "last_name", input.LastName, "error", err)
		return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
	}
	if existing != nil {
		slog.Warn("guest.service create: duplicate name", "first_name", input.FirstName, "last_name", input.LastName)
		return nil, fmt.Errorf("já existe um convidado com o nome '%s %s'", input.FirstName, input.LastName)
	}

	if input.Phone != "" {
		existingByPhone, err := s.repo.GetByPhone(ctx, input.Phone)
		if err != nil {
			slog.Error("guest.service create: phone lookup failed", "phone", input.Phone, "error", err)
			return nil, fmt.Errorf("failed to check phone uniqueness: %w", err)
		}
		if existingByPhone != nil {
			slog.Warn("guest.service create: duplicate phone", "phone", input.Phone)
			return nil, fmt.Errorf("o telefone '%s' já está cadastrado para outro convidado", input.Phone)
		}
	}

	if input.FamilyGroup != nil {
		familyGroupExists, err := s.repo.FamilyGroupExists(ctx, *input.FamilyGroup)
		if err != nil {
			slog.Error("guest.service create: family_group lookup failed", "family_group", *input.FamilyGroup, "error", err)
			return nil, fmt.Errorf("failed to validate family_group: %w", err)
		}
		if !familyGroupExists {
			slog.Warn("guest.service create: family_group not found", "family_group", *input.FamilyGroup)
			return nil, errors.New("grupo familiar não encontrado")
		}
	} else {
		nextFamilyGroup, err := s.repo.GetNextFamilyGroup(ctx)
		if err != nil {
			slog.Error("guest.service create: failed to get next family_group", "error", err)
			return nil, fmt.Errorf("failed to generate family_group: %w", err)
		}
		input.FamilyGroup = &nextFamilyGroup
	}

	guest, err := s.repo.Create(ctx, input, userRACF)
	if err != nil {
		slog.Error("guest.service create: repository create failed", "user_racf", userRACF, "error", err)
		return nil, err
	}
	slog.Info("guest.service create: guest created", "id", guest.ID, "user_racf", userRACF)
	return guest, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	if err := validateUpdate(input); err != nil {
		slog.Warn("guest.service update: validation failed", "id", id, "error", err)
		return nil, err
	}

	exists, err := s.userChecker.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		slog.Error("guest.service update: user check failed", "id", id, "user_racf", userRACF, "error", err)
		return nil, fmt.Errorf("failed to verify user: %w", err)
	}
	if !exists {
		slog.Warn("guest.service update: unknown user racf", "id", id, "user_racf", userRACF)
		return nil, errors.New("usuário não autorizado para realizar esta operação")
	}

	if input.Phone != nil && *input.Phone != "" {
		existingByPhone, err := s.repo.GetByPhone(ctx, *input.Phone)
		if err != nil {
			slog.Error("guest.service update: phone lookup failed", "id", id, "phone", *input.Phone, "error", err)
			return nil, fmt.Errorf("failed to check phone uniqueness: %w", err)
		}
		if existingByPhone != nil && existingByPhone.ID != id {
			slog.Warn("guest.service update: duplicate phone", "id", id, "phone", *input.Phone)
			return nil, fmt.Errorf("o telefone '%s' já está cadastrado para outro convidado", *input.Phone)
		}
	}

	if input.FirstName != nil || input.LastName != nil {
		current, err := s.repo.GetByID(ctx, id)
		if err != nil {
			slog.Error("guest.service update: current guest fetch failed", "id", id, "error", err)
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
			slog.Error("guest.service update: name lookup failed", "id", id, "first_name", firstName, "last_name", lastName, "error", err)
			return nil, fmt.Errorf("failed to check name uniqueness: %w", err)
		}
		if existing != nil && existing.ID != id {
			slog.Warn("guest.service update: duplicate name", "id", id, "first_name", firstName, "last_name", lastName)
			return nil, fmt.Errorf("já existe um convidado com o nome '%s %s'", firstName, lastName)
		}
	}

	guest, err := s.repo.Update(ctx, id, input, userRACF)
	if err != nil {
		slog.Error("guest.service update: repository update failed", "id", id, "user_racf", userRACF, "error", err)
		return nil, err
	}
	slog.Info("guest.service update: guest updated", "id", guest.ID, "user_racf", userRACF)
	return guest, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("guest.service delete: failed", "id", id, "error", err)
		return err
	}
	slog.Info("guest.service delete: guest deleted", "id", id)
	return nil
}

var phoneRegex = regexp.MustCompile(`^\d{2}9\d{8}$`)

func validateCreate(input CreateGuestInput) error {
	if input.FirstName == "" {
		return errors.New("o nome é obrigatório")
	}
	if input.LastName == "" {
		return errors.New("o sobrenome é obrigatório")
	}
	if input.Phone != "" && !phoneRegex.MatchString(input.Phone) {
		return errors.New("telefone inválido. Use o formato: DDD + 9 + 8 dígitos (ex: 11912345678)")
	}
	if input.Relationship != "P" && input.Relationship != "R" {
		return errors.New("tipo de relacionamento inválido")
	}
	if input.FamilyGroup != nil && *input.FamilyGroup <= 0 {
		return errors.New("grupo familiar deve ser maior que zero")
	}
	return nil
}

func validateUpdate(input UpdateGuestInput) error {
	if input.Phone != nil && *input.Phone != "" && !phoneRegex.MatchString(*input.Phone) {
		return errors.New("telefone inválido. Use o formato: DDD + 9 + 8 dígitos (ex: 11912345678)")
	}
	if input.Relationship != nil && *input.Relationship != "P" && *input.Relationship != "R" {
		return errors.New("tipo de relacionamento inválido")
	}
	return nil
}
