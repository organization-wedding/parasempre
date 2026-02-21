package guest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/ferjunior7/parasempre/backend/internal/apperr"
)

// ErrNotFound is the sentinel returned by the repository when a guest row
// does not exist.  It is kept exported so that mocks and tests can reference
// it via errors.Is.
var ErrNotFound = errors.New("guest not found")

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
		return nil, apperr.Internal(err)
	}
	return guests, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Guest, error) {
	guest, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			slog.Warn("guest.service get_by_id: not found", "id", id)
			return nil, apperr.NotFound("Convidado não encontrado", err)
		}
		slog.Error("guest.service get_by_id: failed", "id", id, "error", err)
		return nil, apperr.Internal(err)
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
		return nil, apperr.Internal(err)
	}
	if !exists {
		slog.Warn("guest.service create: unknown user racf", "user_racf", userRACF)
		return nil, apperr.Forbidden("usuário não autorizado para realizar esta operação")
	}

	existing, err := s.repo.GetByName(ctx, input.FirstName, input.LastName)
	if err != nil {
		slog.Error("guest.service create: name lookup failed", "first_name", input.FirstName, "last_name", input.LastName, "error", err)
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		slog.Warn("guest.service create: duplicate name", "first_name", input.FirstName, "last_name", input.LastName)
		return nil, apperr.Conflict(fmt.Sprintf("já existe um convidado com o nome '%s %s'", input.FirstName, input.LastName), nil)
	}

	if input.Phone != "" {
		existingByPhone, err := s.repo.GetByPhone(ctx, input.Phone)
		if err != nil {
			slog.Error("guest.service create: phone lookup failed", "phone", input.Phone, "error", err)
			return nil, apperr.Internal(err)
		}
		if existingByPhone != nil {
			slog.Warn("guest.service create: duplicate phone", "phone", input.Phone)
			return nil, apperr.Conflict(fmt.Sprintf("o telefone '%s' já está cadastrado para outro convidado", input.Phone), nil)
		}
	}

	if input.FamilyGroup != nil {
		familyGroupExists, err := s.repo.FamilyGroupExists(ctx, *input.FamilyGroup)
		if err != nil {
			slog.Error("guest.service create: family_group lookup failed", "family_group", *input.FamilyGroup, "error", err)
			return nil, apperr.Internal(err)
		}
		if !familyGroupExists {
			slog.Warn("guest.service create: family_group not found", "family_group", *input.FamilyGroup)
			return nil, apperr.NotFound("grupo familiar não encontrado", nil)
		}
	} else {
		nextFamilyGroup, err := s.repo.GetNextFamilyGroup(ctx)
		if err != nil {
			slog.Error("guest.service create: failed to get next family_group", "error", err)
			return nil, apperr.Internal(err)
		}
		input.FamilyGroup = &nextFamilyGroup
	}

	guest, err := s.repo.Create(ctx, input, userRACF)
	if err != nil {
		slog.Error("guest.service create: repository create failed", "user_racf", userRACF, "error", err)
		return nil, apperr.Internal(err)
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
		return nil, apperr.Internal(err)
	}
	if !exists {
		slog.Warn("guest.service update: unknown user racf", "id", id, "user_racf", userRACF)
		return nil, apperr.Forbidden("usuário não autorizado para realizar esta operação")
	}

	if input.Phone != nil && *input.Phone != "" {
		existingByPhone, err := s.repo.GetByPhone(ctx, *input.Phone)
		if err != nil {
			slog.Error("guest.service update: phone lookup failed", "id", id, "phone", *input.Phone, "error", err)
			return nil, apperr.Internal(err)
		}
		if existingByPhone != nil && existingByPhone.ID != id {
			slog.Warn("guest.service update: duplicate phone", "id", id, "phone", *input.Phone)
			return nil, apperr.Conflict(fmt.Sprintf("o telefone '%s' já está cadastrado para outro convidado", *input.Phone), nil)
		}
	}

	if input.FirstName != nil || input.LastName != nil {
		current, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				slog.Warn("guest.service update: guest not found", "id", id)
				return nil, apperr.NotFound("Convidado não encontrado", err)
			}
			slog.Error("guest.service update: current guest fetch failed", "id", id, "error", err)
			return nil, apperr.Internal(err)
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
			return nil, apperr.Internal(err)
		}
		if existing != nil && existing.ID != id {
			slog.Warn("guest.service update: duplicate name", "id", id, "first_name", firstName, "last_name", lastName)
			return nil, apperr.Conflict(fmt.Sprintf("já existe um convidado com o nome '%s %s'", firstName, lastName), nil)
		}
	}

	guest, err := s.repo.Update(ctx, id, input, userRACF)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			slog.Warn("guest.service update: guest not found", "id", id)
			return nil, apperr.NotFound("Convidado não encontrado", err)
		}
		slog.Error("guest.service update: repository update failed", "id", id, "user_racf", userRACF, "error", err)
		return nil, apperr.Internal(err)
	}
	slog.Info("guest.service update: guest updated", "id", guest.ID, "user_racf", userRACF)
	return guest, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			slog.Warn("guest.service delete: guest not found", "id", id)
			return apperr.NotFound("Convidado não encontrado", err)
		}
		slog.Error("guest.service delete: failed", "id", id, "error", err)
		return apperr.Internal(err)
	}
	slog.Info("guest.service delete: guest deleted", "id", id)
	return nil
}

var phoneRegex = regexp.MustCompile(`^\d{2}9\d{8}$`)

func validateCreate(input CreateGuestInput) error {
	if input.FirstName == "" {
		return apperr.Validation("o nome é obrigatório")
	}
	if input.LastName == "" {
		return apperr.Validation("o sobrenome é obrigatório")
	}
	if input.Phone != "" && !phoneRegex.MatchString(input.Phone) {
		return apperr.Validation("telefone inválido. Use o formato: DDD + 9 + 8 dígitos (ex: 11912345678)")
	}
	if input.Relationship != "P" && input.Relationship != "R" {
		return apperr.Validation("tipo de relacionamento inválido")
	}
	if input.FamilyGroup != nil && *input.FamilyGroup <= 0 {
		return apperr.Validation("grupo familiar deve ser maior que zero")
	}
	return nil
}

func validateUpdate(input UpdateGuestInput) error {
	if input.Phone != nil && *input.Phone != "" && !phoneRegex.MatchString(*input.Phone) {
		return apperr.Validation("telefone inválido. Use o formato: DDD + 9 + 8 dígitos (ex: 11912345678)")
	}
	if input.Relationship != nil && *input.Relationship != "P" && *input.Relationship != "R" {
		return apperr.Validation("tipo de relacionamento inválido")
	}
	return nil
}
