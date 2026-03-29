package guest

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

// UserChecker verifies whether a RACF belongs to a registered user.
type UserChecker interface {
	UserExistsByURACF(ctx context.Context, uracf string) (bool, error)
}

// UserCreator creates a user linked to a guest within a transaction.
type UserCreator interface {
	CreateGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error
}

type Service struct {
	repo        TxAwareRepository
	userChecker UserChecker
	userCreator UserCreator
	txRunner    database.TxRunner
}

func NewService(repo TxAwareRepository, userChecker UserChecker, userCreator UserCreator, txRunner database.TxRunner) *Service {
	return &Service{repo: repo, userChecker: userChecker, userCreator: userCreator, txRunner: txRunner}
}

func (s *Service) List(ctx context.Context) ([]Guest, error) {
	guests, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("guest.service list: failed", "error", err)
		return nil, apperror.Internal("failed to list guests", err)
	}
	return guests, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Guest, error) {
	guest, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if _, ok := apperror.IsAppError(err); ok {
			return nil, err
		}
		slog.Error("guest.service get_by_id: failed", "id", id, "error", err)
		return nil, apperror.Internal("failed to get guest", err)
	}
	return guest, nil
}

func (s *Service) Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	exists, err := s.userChecker.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		slog.Error("guest.service create: user check failed", "user_racf", userRACF, "error", err)
		return nil, apperror.Internal("failed to verify user", err)
	}
	if !exists {
		return nil, apperror.Validation("user-racf does not match any registered user")
	}

	existing, err := s.repo.GetByName(ctx, input.FirstName, input.LastName)
	if err != nil {
		slog.Error("guest.service create: name lookup failed", "error", err)
		return nil, apperror.Internal("failed to check name uniqueness", err)
	}
	if existing != nil {
		return nil, apperror.Conflict(fmt.Sprintf("a guest named '%s %s' already exists", input.FirstName, input.LastName))
	}

	if input.FamilyGroup != nil {
		familyGroupExists, err := s.repo.FamilyGroupExists(ctx, *input.FamilyGroup)
		if err != nil {
			slog.Error("guest.service create: family_group lookup failed", "error", err)
			return nil, apperror.Internal("failed to validate family_group", err)
		}
		if !familyGroupExists {
			return nil, apperror.Validation("family_group not found")
		}
	} else {
		nextFamilyGroup, err := s.repo.GetNextFamilyGroup(ctx)
		if err != nil {
			slog.Error("guest.service create: failed to get next family_group", "error", err)
			return nil, apperror.Internal("failed to generate family_group", err)
		}
		input.FamilyGroup = &nextFamilyGroup
	}

	var created *Guest
	if err := s.txRunner.RunInTx(ctx, func(tx pgx.Tx) error {
		txRepo := s.repo.WithTx(tx)
		g, err := txRepo.Create(ctx, input, userRACF)
		if err != nil {
			return err
		}
		created = g

		if err := s.userCreator.CreateGuestUserTx(ctx, tx, g.ID, input.Phone); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if _, ok := apperror.IsAppError(err); ok {
			return nil, err
		}
		slog.Error("guest.service create: transaction failed", "error", err)
		return nil, apperror.Internal("failed to create guest", err)
	}

	slog.Info("guest.service create: guest+user created", "id", created.ID, "user_racf", userRACF)
	return created, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	exists, err := s.userChecker.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		slog.Error("guest.service update: user check failed", "error", err)
		return nil, apperror.Internal("failed to verify user", err)
	}
	if !exists {
		return nil, apperror.Validation("user-racf does not match any registered user")
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
			slog.Error("guest.service update: name lookup failed", "error", err)
			return nil, apperror.Internal("failed to check name uniqueness", err)
		}
		if existing != nil && existing.ID != id {
			return nil, apperror.Conflict(fmt.Sprintf("a guest named '%s %s' already exists", firstName, lastName))
		}
	}

	guest, err := s.repo.Update(ctx, id, input, userRACF)
	if err != nil {
		if _, ok := apperror.IsAppError(err); ok {
			return nil, err
		}
		slog.Error("guest.service update: repository update failed", "error", err)
		return nil, apperror.Internal("failed to update guest", err)
	}
	slog.Info("guest.service update: guest updated", "id", guest.ID, "user_racf", userRACF)
	return guest, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if _, ok := apperror.IsAppError(err); ok {
			return err
		}
		slog.Error("guest.service delete: failed", "id", id, "error", err)
		return apperror.Internal("failed to delete guest", err)
	}
	slog.Info("guest.service delete: guest deleted", "id", id)
	return nil
}
