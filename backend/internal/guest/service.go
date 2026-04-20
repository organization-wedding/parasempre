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

type UserBridge interface {
	UserExistsByURACF(ctx context.Context, uracf string) (bool, error)
	CreateGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error
	DeleteGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64) error
	GetGuestIDByPhone(ctx context.Context, phone string) (*int64, error)
	GetGuestIDByUserID(ctx context.Context, userID int64) (*int64, error)
}

type Service struct {
	repo     TxAwareRepository
	users    UserBridge
	txRunner database.TxRunner
}

func NewService(repo TxAwareRepository, users UserBridge, txRunner database.TxRunner) *Service {
	return &Service{repo: repo, users: users, txRunner: txRunner}
}

func (s *Service) List(ctx context.Context, page, limit int, userRACF string) (*PagedResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	guests, total, err := s.repo.List(ctx, limit, offset, userRACF)
	if err != nil {
		slog.Error("guest.service list: failed", "error", err)
		return nil, apperror.Internal("failed to list guests", err)
	}
	return &PagedResponse{
		Data:  guests,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id int64, userRACF string) (*Guest, error) {
	guest, err := s.repo.GetByID(ctx, id, userRACF)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to get guest", err)
	}
	return guest, nil
}

func (s *Service) Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	exists, err := s.users.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		slog.Error("guest.service create: user check failed", "user_racf", userRACF, "error", err)
		return nil, apperror.Internal("failed to verify user", err)
	}
	if !exists {
		return nil, apperror.Validation("user-racf does not match any registered user")
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

		if err := s.users.CreateGuestUserTx(ctx, tx, g.ID, input.Phone); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, apperror.WrapIfNotApp("failed to create guest", err)
	}

	slog.Info("guest.service create: guest+user created", "id", created.ID, "user_racf", userRACF)
	return created, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	exists, err := s.users.UserExistsByURACF(ctx, userRACF)
	if err != nil {
		slog.Error("guest.service update: user check failed", "error", err)
		return nil, apperror.Internal("failed to verify user", err)
	}
	if !exists {
		return nil, apperror.Validation("user-racf does not match any registered user")
	}

	if input.FirstName != nil || input.LastName != nil {
		current, err := s.repo.GetByID(ctx, id, userRACF)
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
		return nil, apperror.WrapIfNotApp("failed to update guest", err)
	}
	slog.Info("guest.service update: guest updated", "id", guest.ID, "user_racf", userRACF)
	return guest, nil
}

func (s *Service) Confirm(ctx context.Context, id int64, userID int64) (*Guest, error) {
	return s.setConfirmed(ctx, id, true, userID)
}

func (s *Service) Cancel(ctx context.Context, id int64, userID int64) (*Guest, error) {
	return s.setConfirmed(ctx, id, false, userID)
}

func (s *Service) ConfirmByPhone(ctx context.Context, phone string, userID int64) (*Guest, error) {
	return s.setConfirmedByPhone(ctx, phone, true, userID)
}

func (s *Service) CancelByPhone(ctx context.Context, phone string, userID int64) (*Guest, error) {
	return s.setConfirmedByPhone(ctx, phone, false, userID)
}

func (s *Service) ConfirmFamily(ctx context.Context, familyGroup int64, userID int64) ([]Guest, error) {
	return s.setConfirmedFamily(ctx, familyGroup, true, userID)
}

func (s *Service) CancelFamily(ctx context.Context, familyGroup int64, userID int64) ([]Guest, error) {
	return s.setConfirmedFamily(ctx, familyGroup, false, userID)
}

func (s *Service) ConfirmFamilyByPhone(ctx context.Context, phone string, userID int64) ([]Guest, error) {
	return s.setConfirmedFamilyByPhone(ctx, phone, true, userID)
}

func (s *Service) CancelFamilyByPhone(ctx context.Context, phone string, userID int64) ([]Guest, error) {
	return s.setConfirmedFamilyByPhone(ctx, phone, false, userID)
}

func (s *Service) setConfirmed(ctx context.Context, id int64, confirmed bool, userID int64) (*Guest, error) {
	currentUserGuestID, err := s.users.GetGuestIDByUserID(ctx, userID)
	if err != nil {
		slog.Error("guest.service set_confirmed: failed to get current user's guest", "user_id", userID, "error", err)
		return nil, apperror.Internal("failed to verify guest identity", err)
	}
	if currentUserGuestID == nil {
		return nil, apperror.Forbidden("only guests can confirm their attendance")
	}

	if *currentUserGuestID != id {
		slog.Warn("guest.service set_confirmed: unauthorized attempt", "user_id", userID, "requested_guest_id", id, "actual_guest_id", *currentUserGuestID)
		return nil, apperror.Forbidden("you can only confirm your own attendance")
	}

	current, err := s.repo.GetByID(ctx, id, "")
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to fetch guest", err)
	}
	if current.Confirmed == confirmed {
		slog.Info("guest.service set_confirmed: already in desired state, skipping update", "id", id, "confirmed", confirmed)
		return current, nil
	}

	guest, err := s.repo.SetConfirmed(ctx, id, confirmed, "")
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to update confirmation", err)
	}
	slog.Info("guest.service set_confirmed: success", "id", guest.ID, "confirmed", confirmed, "user_id", userID)
	return guest, nil
}

func (s *Service) setConfirmedByPhone(ctx context.Context, phone string, confirmed bool, userID int64) (*Guest, error) {
	guestID, err := s.users.GetGuestIDByPhone(ctx, phone)
	if err != nil {
		slog.Error("guest.service set_confirmed_by_phone: phone lookup failed", "phone", phone, "error", err)
		return nil, apperror.Internal("failed to find guest by phone", err)
	}
	if guestID == nil {
		return nil, apperror.NotFound("no guest found for this phone number")
	}

	return s.setConfirmed(ctx, *guestID, confirmed, userID)
}

func (s *Service) setConfirmedFamily(ctx context.Context, familyGroup int64, confirmed bool, userID int64) ([]Guest, error) {
	currentUserGuestID, err := s.users.GetGuestIDByUserID(ctx, userID)
	if err != nil {
		slog.Error("guest.service set_confirmed_family: failed to get current user's guest", "user_id", userID, "error", err)
		return nil, apperror.Internal("failed to verify guest identity", err)
	}
	if currentUserGuestID == nil {
		return nil, apperror.Forbidden("only guests can confirm family attendance")
	}

	currentGuest, err := s.repo.GetByID(ctx, *currentUserGuestID, "")
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to fetch current guest", err)
	}

	if currentGuest.FamilyGroup != familyGroup {
		slog.Warn("guest.service set_confirmed_family: unauthorized attempt", "user_id", userID, "requested_family", familyGroup, "actual_family", currentGuest.FamilyGroup)
		return nil, apperror.Forbidden("you can only confirm your own family's attendance")
	}

	familyGroupExists, err := s.repo.FamilyGroupExists(ctx, familyGroup)
	if err != nil {
		slog.Error("guest.service set_confirmed_family: family group lookup failed", "family_group", familyGroup, "error", err)
		return nil, apperror.Internal("failed to validate family_group", err)
	}
	if !familyGroupExists {
		return nil, apperror.NotFound("family group not found")
	}

	guests, err := s.repo.SetConfirmedByFamilyGroup(ctx, familyGroup, confirmed, "")
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to update family confirmation", err)
	}

	slog.Info("guest.service set_confirmed_family: success", "family_group", familyGroup, "confirmed", confirmed, "count", len(guests), "user_id", userID)
	return guests, nil
}

func (s *Service) setConfirmedFamilyByPhone(ctx context.Context, phone string, confirmed bool, userID int64) ([]Guest, error) {
	familyGroup, err := s.repo.GetFamilyGroupByPhone(ctx, phone)
	if err != nil {
		slog.Error("guest.service set_confirmed_family_by_phone: phone lookup failed", "phone", phone, "error", err)
		return nil, apperror.Internal("failed to find family by phone", err)
	}
	if familyGroup == nil {
		return nil, apperror.NotFound("no family found for this phone number")
	}

	return s.setConfirmedFamily(ctx, *familyGroup, confirmed, userID)
}

func (s *Service) Import(ctx context.Context, guests []CreateGuestInput, userRACF string) ImportResponse {
	var successCount int
	var rowErrors []ImportRowError
	const dataRowStart = 2
	for i, input := range guests {
		rowNumber := i + dataRowStart
		if _, err := s.Create(ctx, input, userRACF); err != nil {
			slog.Warn("guest.service import: row failed", "row", rowNumber, "error", err)
			rowErrors = append(rowErrors, ImportRowError{Row: rowNumber, Error: err.Error()})
			continue
		}
		successCount++
	}
	if rowErrors == nil {
		rowErrors = []ImportRowError{}
	}
	return ImportResponse{
		SuccessCount: successCount,
		ErrorCount:   len(rowErrors),
		Total:        len(guests),
		Errors:       rowErrors,
	}
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if err := s.txRunner.RunInTx(ctx, func(tx pgx.Tx) error {
		if err := s.users.DeleteGuestUserTx(ctx, tx, id); err != nil {
			return err
		}
		txRepo := s.repo.WithTx(tx)
		return txRepo.Delete(ctx, id)
	}); err != nil {
		return apperror.WrapIfNotApp("failed to delete guest", err)
	}
	slog.Info("guest.service delete: guest deleted", "id", id)
	return nil
}
