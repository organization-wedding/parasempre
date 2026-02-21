package user

import (
	"context"
	"errors"
	"log/slog"
	"regexp"

	"github.com/ferjunior7/parasempre/backend/internal/apperr"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
)

// Sentinel errors kept exported so that errors.Is() keeps working in tests
// and in any caller that needs to distinguish error kinds.
var (
	ErrAlreadyRegistered = errors.New("user already registered")
	ErrGuestNotFound     = errors.New("no guest found with this phone")
	ErrURACFTaken        = errors.New("uracf already in use")
)

type Service struct {
	repo      Repository
	guestRepo guest.Repository
}

func NewService(repo Repository, guestRepo guest.Repository) *Service {
	return &Service{repo: repo, guestRepo: guestRepo}
}

var (
	phoneRegex = regexp.MustCompile(`^\d{2}9\d{8}$`)
	uracfRegex = regexp.MustCompile(`^[A-Z0-9]{5}$`)
)

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	if input.Phone == "" {
		slog.Warn("user.service register: missing phone")
		return nil, apperr.Validation("o telefone é obrigatório")
	}
	if !phoneRegex.MatchString(input.Phone) {
		slog.Warn("user.service register: invalid phone", "phone", input.Phone)
		return nil, apperr.Validation("telefone inválido. Use o formato: DDD + 9 + 8 dígitos (ex: 11912345678)")
	}
	if input.URACF == "" {
		slog.Warn("user.service register: missing uracf", "phone", input.Phone)
		return nil, apperr.Validation("o identificador de acesso é obrigatório")
	}
	if !uracfRegex.MatchString(input.URACF) {
		slog.Warn("user.service register: invalid uracf", "uracf", input.URACF)
		return nil, apperr.Validation("identificador de acesso inválido")
	}

	g, err := s.guestRepo.GetByPhone(ctx, input.Phone)
	if err != nil {
		slog.Error("user.service register: guest lookup failed", "phone", input.Phone, "error", err)
		return nil, apperr.Internal(err)
	}
	if g == nil {
		slog.Warn("user.service register: guest not found", "phone", input.Phone)
		return nil, apperr.NotFound("Nenhum convidado encontrado com este telefone", ErrGuestNotFound)
	}

	existing, err := s.repo.GetByGuestID(ctx, g.ID)
	if err != nil {
		slog.Error("user.service register: guest user lookup failed", "guest_id", g.ID, "error", err)
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		slog.Warn("user.service register: guest already registered", "guest_id", g.ID)
		return nil, apperr.Conflict("Este convidado já possui cadastro", ErrAlreadyRegistered)
	}

	existingByURACF, err := s.repo.GetByURACF(ctx, input.URACF)
	if err != nil {
		slog.Error("user.service register: uracf lookup failed", "uracf", input.URACF, "error", err)
		return nil, apperr.Internal(err)
	}
	if existingByURACF != nil {
		slog.Warn("user.service register: uracf taken", "uracf", input.URACF)
		return nil, apperr.Conflict("Este identificador de acesso já está em uso", ErrURACFTaken)
	}

	u := &User{
		GuestID: &g.ID,
		Role:    "guest",
		URACF:   input.URACF,
	}

	created, err := s.repo.Create(ctx, u)
	if err != nil {
		slog.Error("user.service register: create failed", "guest_id", g.ID, "uracf", input.URACF, "error", err)
		return nil, apperr.Internal(err)
	}
	slog.Info("user.service register: user created", "id", created.ID, "guest_id", g.ID)
	return created, nil
}

func (s *Service) List(ctx context.Context) ([]UserListItem, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("user.service list: failed", "error", err)
		return nil, apperr.Internal(err)
	}
	return items, nil
}

func (s *Service) GetMe(ctx context.Context, uracf string) (*User, error) {
	u, err := s.repo.GetByURACF(ctx, uracf)
	if err != nil {
		slog.Error("user.service get_me: lookup failed", "uracf", uracf, "error", err)
		return nil, apperr.Internal(err)
	}
	if u == nil {
		return nil, apperr.NotFound("Usuário não encontrado", nil)
	}
	return u, nil
}

func (s *Service) CheckByPhone(ctx context.Context, phone string) (*CheckResponse, error) {
	if phone == "" {
		slog.Warn("user.service check: missing phone")
		return nil, apperr.Validation("o telefone é obrigatório")
	}
	if !phoneRegex.MatchString(phone) {
		slog.Warn("user.service check: invalid phone", "phone", phone)
		return nil, apperr.Validation("telefone inválido. Use o formato: DDD + 9 + 8 dígitos (ex: 11912345678)")
	}

	g, err := s.guestRepo.GetByPhone(ctx, phone)
	if err != nil {
		slog.Error("user.service check: guest lookup failed", "phone", phone, "error", err)
		return nil, apperr.Internal(err)
	}
	if g == nil {
		slog.Info("user.service check: guest not found", "phone", phone)
		return &CheckResponse{Exists: false}, nil
	}

	u, err := s.repo.GetByGuestID(ctx, g.ID)
	if err != nil {
		slog.Error("user.service check: user lookup failed", "guest_id", g.ID, "error", err)
		return nil, apperr.Internal(err)
	}
	if u == nil {
		slog.Info("user.service check: no user for guest", "guest_id", g.ID)
		return &CheckResponse{Exists: false}, nil
	}

	slog.Info("user.service check: user exists", "guest_id", g.ID, "role", u.Role)
	return &CheckResponse{Exists: true, Role: u.Role}, nil
}

type CoupleData struct {
	URACF string
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

	created, err := s.repo.Create(ctx, u)
	if err != nil {
		slog.Error("seed: failed to create user", "role", role, "error", err)
		return
	}
	slog.Info("seed: user created", "role", role, "id", created.ID)
}
