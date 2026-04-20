package user

import (
	"context"
	"crypto/rand"
	"log/slog"
	"math/big"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

type TxAwareRepository interface {
	Repository
	WithTx(tx pgx.Tx) Repository
}

type Service struct {
	repo      Repository
	txRepo    TxAwareRepository
	guestRepo guest.Repository
}

func NewService(repo Repository, guestRepo guest.Repository) *Service {
	return &Service{repo: repo, guestRepo: guestRepo}
}

func NewServiceWithTx(repo TxAwareRepository, guestRepo guest.Repository) *Service {
	return &Service{repo: repo, txRepo: repo, guestRepo: guestRepo}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByPhone(ctx, input.Phone)
	if err != nil {
		slog.Error("user.service register: user lookup failed", "phone", input.Phone, "error", err)
		return nil, apperror.Internal("failed to lookup user", err)
	}
	if existing != nil {
		return nil, apperror.Conflict("user already registered with this phone")
	}

	existingByURACF, err := s.repo.GetByURACF(ctx, input.URACF)
	if err != nil {
		slog.Error("user.service register: uracf lookup failed", "uracf", input.URACF, "error", err)
		return nil, apperror.Internal("failed to check uracf", err)
	}
	if existingByURACF != nil {
		return nil, apperror.Conflict("uracf already in use")
	}

	u := &User{
		Role:  "guest",
		URACF: input.URACF,
		Phone: &input.Phone,
	}

	created, err := s.repo.Create(ctx, u)
	if err != nil {
		slog.Error("user.service register: create failed", "uracf", input.URACF, "error", err)
		return nil, apperror.Internal("failed to create user", err)
	}
	slog.Info("user.service register: user created", "id", created.ID)
	return created, nil
}

func (s *Service) List(ctx context.Context) ([]UserListItem, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return nil, apperror.Internal("failed to list users", err)
	}
	return items, nil
}

func (s *Service) GetMe(ctx context.Context, uracf string) (*User, error) {
	u, err := s.repo.GetByURACF(ctx, uracf)
	if err != nil {
		slog.Error("user.service get_me: lookup failed", "uracf", uracf, "error", err)
		return nil, apperror.Internal("failed to get user", err)
	}
	return u, nil
}

func (s *Service) CheckByPhone(ctx context.Context, phone string) (*CheckResponse, error) {
	type phoneInput struct {
		Phone string `validate:"required,brphone"`
	}
	if err := validate.Struct(phoneInput{Phone: phone}); err != nil {
		return nil, err
	}

	u, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		slog.Error("user.service check: user lookup failed", "phone", phone, "error", err)
		return nil, apperror.Internal("failed to check phone", err)
	}
	if u != nil {
		return &CheckResponse{Exists: true}, nil
	}

	return &CheckResponse{Exists: false}, nil
}

func (s *Service) FindOrCreateByPhone(ctx context.Context, phone string) (int64, string, string, error) {
	u, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		slog.Error("user.service find_or_create: user lookup failed", "phone", phone, "error", err)
		return 0, "", "", apperror.Internal("failed to find user", err)
	}
	if u != nil {
		return u.ID, u.URACF, u.Role, nil
	}

	return 0, "", "", apperror.NotFound("no user found with this phone")
}

func (s *Service) GetGuestIDByPhone(ctx context.Context, phone string) (*int64, error) {
	u, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to find guest by phone", err)
	}
	if u == nil || u.GuestID == nil {
		return nil, nil
	}
	return u.GuestID, nil
}

func (s *Service) GetGuestIDByUserID(ctx context.Context, userID int64) (*int64, error) {
	u, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to find user", err)
	}
	if u == nil || u.GuestID == nil {
		return nil, nil
	}
	return u.GuestID, nil
}

func (s *Service) UserExistsByURACF(ctx context.Context, uracf string) (bool, error) {
	u, err := s.repo.GetByURACF(ctx, uracf)
	if err != nil {
		return false, apperror.WrapIfNotApp("failed to check user existence", err)
	}
	return u != nil, nil
}

func (s *Service) PhoneExists(ctx context.Context, phone string) (bool, error) {
	u, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		return false, apperror.WrapIfNotApp("failed to check phone existence", err)
	}
	return u != nil, nil
}

func (s *Service) RecordLogin(ctx context.Context, userID int64) {
	if err := s.repo.UpdateLastLogin(ctx, userID); err != nil {
		slog.Error("user.service record_login: update last_login failed", "user_id", userID, "error", err)
	}
	if err := s.repo.LogAction(ctx, userID, "login", nil); err != nil {
		slog.Error("user.service record_login: log action failed", "user_id", userID, "error", err)
	}
}

func (s *Service) CreateGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64, phone *string) error {
	uracf, err := GenerateURACF()
	if err != nil {
		slog.Error("user.service create_guest_user: uracf generation failed", "error", err)
		return apperror.Internal("failed to generate uracf", err)
	}

	u := &User{
		GuestID: &guestID,
		Role:    "guest",
		URACF:   uracf,
		Phone:   phone,
	}

	txRepo := s.txRepo.WithTx(tx)
	if _, err := txRepo.Create(ctx, u); err != nil {
		slog.Error("user.service create_guest_user: create failed", "guest_id", guestID, "error", err)
		return apperror.Internal("failed to create guest user", err)
	}

	slog.Info("user.service create_guest_user: user created", "guest_id", guestID, "uracf", uracf)
	return nil
}

func (s *Service) DeleteGuestUserTx(ctx context.Context, tx pgx.Tx, guestID int64) error {
	txRepo := s.txRepo.WithTx(tx)
	if err := txRepo.DeleteByGuestID(ctx, guestID); err != nil {
		slog.Error("user.service delete_guest_user: delete failed", "guest_id", guestID, "error", err)
		return apperror.Internal("failed to delete guest user", err)
	}
	slog.Info("user.service delete_guest_user: user deleted", "guest_id", guestID)
	return nil
}

const uracfChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateURACF() (string, error) {
	result := make([]byte, 5)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(uracfChars))))
		if err != nil {
			return "", err
		}
		result[i] = uracfChars[n.Int64()]
	}
	return string(result), nil
}

func (s *Service) SeedCouple(ctx context.Context, groom, bride CoupleData) {
	s.seedPerson(ctx, groom, "groom")
	s.seedPerson(ctx, bride, "bride")
}

func (s *Service) seedPerson(ctx context.Context, data CoupleData, role string) {
	if data.URACF == "" {
		slog.Info("seed: skipping, uracf is empty", "role", role)
		return
	}

	existingRole, err := s.repo.GetByRole(ctx, role)
	if err != nil {
		slog.Error("seed: failed to check existing role", "role", role, "error", err)
		return
	}
	if existingRole != nil {
		slog.Info("seed: skipping, role already exists", "role", role, "existing_uracf", existingRole.URACF)
		return
	}

	existing, err := s.repo.GetByURACF(ctx, data.URACF)
	if err != nil {
		return
	}
	if existing != nil {
		return
	}

	u := &User{
		Role:  role,
		URACF: data.URACF,
	}
	if data.Phone != "" {
		u.Phone = &data.Phone
	}

	created, err := s.repo.Create(ctx, u)
	if err != nil {
		slog.Error("seed: failed to create user", "role", role, "error", err)
		return
	}
	slog.Info("seed: user created", "role", role, "id", created.ID)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*User, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		slog.Error("user.service update: lookup failed", "id", id, "error", err)
		return nil, apperror.Internal("failed to lookup user", err)
	}
	if existing == nil {
		return nil, apperror.NotFound("user not found")
	}

	if input.Phone != nil && *input.Phone != "" {
		phoneUser, err := s.repo.GetByPhone(ctx, *input.Phone)
		if err != nil {
			slog.Error("user.service update: phone lookup failed", "phone", *input.Phone, "error", err)
			return nil, apperror.Internal("failed to check phone", err)
		}
		if phoneUser != nil && phoneUser.ID != id {
			return nil, apperror.Conflict("phone already in use")
		}
	}

	if input.Role != nil && (*input.Role == "groom" || *input.Role == "bride") {
		existingRole, err := s.repo.GetByRole(ctx, *input.Role)
		if err != nil {
			slog.Error("user.service update: role lookup failed", "role", *input.Role, "error", err)
			return nil, apperror.Internal("failed to check role", err)
		}
		if existingRole != nil && existingRole.ID != id {
			return nil, apperror.Conflict(*input.Role + " already exists")
		}
	}

	updated, err := s.repo.Update(ctx, id, input)
	if err != nil {
		slog.Error("user.service update: update failed", "id", id, "error", err)
		return nil, apperror.Internal("failed to update user", err)
	}

	slog.Info("user.service update: user updated", "id", id)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		slog.Error("user.service delete: lookup failed", "id", id, "error", err)
		return apperror.Internal("failed to lookup user", err)
	}
	if existing == nil {
		return apperror.NotFound("user not found")
	}

	if existing.Role == "groom" || existing.Role == "bride" {
		return apperror.Forbidden("cannot delete groom or bride users")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		slog.Error("user.service delete: delete failed", "id", id, "error", err)
		return apperror.Internal("failed to delete user", err)
	}

	slog.Info("user.service delete: user deleted", "id", id)
	return nil
}
