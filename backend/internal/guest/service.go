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

func lastN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

type UserBridge interface {
	UserExistsByURACF(ctx context.Context, uracf string) (bool, error)
	CreateGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error
	DeleteGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64) error
	GetGuestIDByPhone(ctx context.Context, phone string) (*int64, error)
	GetGuestIDByUserID(ctx context.Context, userID int64) (*int64, error)
	GetURACFByUserID(ctx context.Context, userID int64) (string, error)
}

type Service struct {
	repo     TxAwareRepository
	users    UserBridge
	txRunner database.TxRunner
}

func NewService(repo TxAwareRepository, users UserBridge, txRunner database.TxRunner) *Service {
	return &Service{repo: repo, users: users, txRunner: txRunner}
}

func (s *Service) ListMyFamily(ctx context.Context, userID int64) ([]Guest, error) {
	guestID, err := s.users.GetGuestIDByUserID(ctx, userID)
	if err != nil {
		slog.Error("guest.service list_my_family: failed to get current user's guest", "user_id", userID, "error", err)
		return nil, apperror.Internal("failed to verify guest identity", err)
	}
	if guestID == nil {
		return []Guest{}, nil
	}

	currentGuest, err := s.repo.GetByIDAny(ctx, *guestID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to fetch current guest", err)
	}

	guests, err := s.repo.ListByFamilyGroup(ctx, currentGuest.FamilyGroup)
	if err != nil {
		return nil, apperror.Internal("failed to list family guests", err)
	}
	return guests, nil
}

func (s *Service) SetConfirmedBatch(ctx context.Context, input BatchConfirmInput, userID int64) ([]Guest, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	currentUserGuestID, err := s.users.GetGuestIDByUserID(ctx, userID)
	if err != nil {
		slog.Error("guest.service set_confirmed_batch: failed to get current user's guest", "user_id", userID, "error", err)
		return nil, apperror.Internal("failed to verify guest identity", err)
	}
	if currentUserGuestID == nil {
		return nil, apperror.Forbidden("only guests can confirm attendance")
	}

	currentGuest, err := s.repo.GetByIDAny(ctx, *currentUserGuestID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to fetch current guest", err)
	}

	targets, err := s.repo.GetByIDs(ctx, input.GuestIDs)
	if err != nil {
		return nil, apperror.Internal("failed to fetch target guests", err)
	}
	if len(targets) != len(input.GuestIDs) {
		return nil, apperror.NotFound("one or more guests not found")
	}
	for _, target := range targets {
		if target.FamilyGroup != currentGuest.FamilyGroup {
			slog.Warn("guest.service set_confirmed_batch: unauthorized cross-family attempt", "user_id", userID, "target_id", target.ID, "caller_family", currentGuest.FamilyGroup, "target_family", target.FamilyGroup)
			return nil, apperror.Forbidden("you can only confirm guests in your own family")
		}
	}

	updatedByRACF, err := s.users.GetURACFByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to resolve caller URACF", err)
	}

	var updated []Guest
	if err := s.txRunner.RunInTx(ctx, func(tx pgx.Tx) error {
		txRepo := s.repo.WithTx(tx)
		guests, err := txRepo.SetAttendingByIDs(ctx, input.GuestIDs, input.Attending, updatedByRACF)
		if err != nil {
			return err
		}
		updated = guests
		return nil
	}); err != nil {
		return nil, apperror.WrapIfNotApp("failed to update batch confirmation", err)
	}

	slog.Info("guest.service set_confirmed_batch: success", "ids", input.GuestIDs, "attending", input.Attending, "count", len(updated), "user_id", userID)
	return updated, nil
}

func (s *Service) List(ctx context.Context, page, limit int, filters ListFilters) (*PagedResponse, error) {
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
	guests, total, err := s.repo.List(ctx, limit, offset, filters)
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

func (s *Service) Stats(ctx context.Context) (*Stats, error) {
	stats, err := s.repo.Stats(ctx)
	if err != nil {
		slog.Error("guest.service stats: failed", "error", err)
		return nil, apperror.Internal("failed to compute guest stats", err)
	}
	return &stats, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Guest, error) {
	guest, err := s.repo.GetByIDAny(ctx, id)
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
		current, err := s.repo.GetByIDAny(ctx, id)
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
	return s.setAttending(ctx, id, true, userID)
}

func (s *Service) Cancel(ctx context.Context, id int64, userID int64) (*Guest, error) {
	return s.setAttending(ctx, id, false, userID)
}

func (s *Service) ConfirmByPhone(ctx context.Context, phone string, userID int64) (*Guest, error) {
	return s.setAttendingByPhone(ctx, phone, true, userID)
}

func (s *Service) CancelByPhone(ctx context.Context, phone string, userID int64) (*Guest, error) {
	return s.setAttendingByPhone(ctx, phone, false, userID)
}

func (s *Service) ConfirmFamily(ctx context.Context, familyGroup int64, userID int64) ([]Guest, error) {
	return s.setAttendingFamily(ctx, familyGroup, true, userID)
}

func (s *Service) CancelFamily(ctx context.Context, familyGroup int64, userID int64) ([]Guest, error) {
	return s.setAttendingFamily(ctx, familyGroup, false, userID)
}

func (s *Service) ConfirmFamilyByPhone(ctx context.Context, phone string, userID int64) ([]Guest, error) {
	return s.setAttendingFamilyByPhone(ctx, phone, true, userID)
}

func (s *Service) CancelFamilyByPhone(ctx context.Context, phone string, userID int64) ([]Guest, error) {
	return s.setAttendingFamilyByPhone(ctx, phone, false, userID)
}

func (s *Service) setAttending(ctx context.Context, id int64, attending bool, userID int64) (*Guest, error) {
	currentUserGuestID, err := s.users.GetGuestIDByUserID(ctx, userID)
	if err != nil {
		slog.Error("guest.service set_attending: failed to get current user's guest", "user_id", userID, "error", err)
		return nil, apperror.Internal("failed to verify guest identity", err)
	}
	if currentUserGuestID == nil {
		return nil, apperror.Forbidden("only guests can confirm their attendance")
	}

	currentGuest, err := s.repo.GetByIDAny(ctx, *currentUserGuestID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to fetch current guest", err)
	}

	target, err := s.repo.GetByIDAny(ctx, id)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to fetch guest", err)
	}

	if target.FamilyGroup != currentGuest.FamilyGroup {
		slog.Warn("guest.service set_attending: unauthorized cross-family attempt", "user_id", userID, "requested_guest_id", id, "caller_family", currentGuest.FamilyGroup, "target_family", target.FamilyGroup)
		return nil, apperror.Forbidden("you can only confirm guests in your own family")
	}

	if target.Attending != nil && *target.Attending == attending {
		slog.Info("guest.service set_attending: already in desired state, skipping update", "id", id, "attending", attending)
		return target, nil
	}

	updatedByRACF, err := s.users.GetURACFByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to resolve caller URACF", err)
	}

	updated, err := s.repo.SetAttending(ctx, id, attending, updatedByRACF)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to update attending", err)
	}
	slog.Info("guest.service set_attending: success", "id", updated.ID, "attending", attending, "user_id", userID)
	return updated, nil
}

func (s *Service) setAttendingByPhone(ctx context.Context, phone string, attending bool, userID int64) (*Guest, error) {
	guestID, err := s.users.GetGuestIDByPhone(ctx, phone)
	if err != nil {
		slog.Error("guest.service set_attending_by_phone: phone lookup failed", "phone_suffix", lastN(phone, 4), "error", err)
		return nil, apperror.Internal("failed to find guest by phone", err)
	}
	if guestID == nil {
		return nil, apperror.NotFound("no guest found for this phone number")
	}

	return s.setAttending(ctx, *guestID, attending, userID)
}

func (s *Service) setAttendingFamily(ctx context.Context, familyGroup int64, attending bool, userID int64) ([]Guest, error) {
	currentUserGuestID, err := s.users.GetGuestIDByUserID(ctx, userID)
	if err != nil {
		slog.Error("guest.service set_attending_family: failed to get current user's guest", "user_id", userID, "error", err)
		return nil, apperror.Internal("failed to verify guest identity", err)
	}
	if currentUserGuestID == nil {
		return nil, apperror.Forbidden("only guests can confirm family attendance")
	}

	currentGuest, err := s.repo.GetByIDAny(ctx, *currentUserGuestID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to fetch current guest", err)
	}

	if currentGuest.FamilyGroup != familyGroup {
		slog.Warn("guest.service set_attending_family: unauthorized attempt", "user_id", userID, "requested_family", familyGroup, "actual_family", currentGuest.FamilyGroup)
		return nil, apperror.Forbidden("you can only confirm your own family's attendance")
	}

	familyGroupExists, err := s.repo.FamilyGroupExists(ctx, familyGroup)
	if err != nil {
		slog.Error("guest.service set_attending_family: family group lookup failed", "family_group", familyGroup, "error", err)
		return nil, apperror.Internal("failed to validate family_group", err)
	}
	if !familyGroupExists {
		return nil, apperror.NotFound("family group not found")
	}

	updatedByRACF, err := s.users.GetURACFByUserID(ctx, userID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to resolve caller URACF", err)
	}

	guests, err := s.repo.SetAttendingByFamilyGroup(ctx, familyGroup, attending, updatedByRACF)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to update family attending", err)
	}

	slog.Info("guest.service set_attending_family: success", "family_group", familyGroup, "attending", attending, "count", len(guests), "user_id", userID)
	return guests, nil
}

func (s *Service) setAttendingFamilyByPhone(ctx context.Context, phone string, attending bool, userID int64) ([]Guest, error) {
	familyGroup, err := s.repo.GetFamilyGroupByPhone(ctx, phone)
	if err != nil {
		slog.Error("guest.service set_attending_family_by_phone: phone lookup failed", "phone_suffix", lastN(phone, 4), "error", err)
		return nil, apperror.Internal("failed to find family by phone", err)
	}
	if familyGroup == nil {
		return nil, apperror.NotFound("no family found for this phone number")
	}

	return s.setAttendingFamily(ctx, *familyGroup, attending, userID)
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
