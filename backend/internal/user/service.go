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

// TxAwareRepository extends Repository with transaction support.
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

	// Find user by phone (phone is now on users table only)
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
		return &CheckResponse{Exists: true, Role: u.Role}, nil
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

// GetGuestIDByPhone implements guest.UserPhoneLookup.
// Busca um user pelo phone e retorna o guest_id associado.
func (s *Service) GetGuestIDByPhone(ctx context.Context, phone string) (*int64, error) {
	u, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if u == nil || u.GuestID == nil {
		return nil, nil
	}
	return u.GuestID, nil
}

func (s *Service) PhoneExists(ctx context.Context, phone string) (bool, error) {
	u, err := s.repo.GetByPhone(ctx, phone)
	if err != nil {
		return false, err
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

// CreateGuestUserTx creates a user linked to a guest within an existing transaction.
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

const uracfChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateURACF generates a random 5-character uppercase alphanumeric string.
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
