package user

import (
	"context"
	"errors"
	"log/slog"
	"regexp"

	"github.com/ferjunior7/parasempre/backend/internal/guest"
)

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
		return nil, errors.New("phone is required")
	}
	if !phoneRegex.MatchString(input.Phone) {
		return nil, errors.New("phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)")
	}
	if input.URACF == "" {
		return nil, errors.New("uracf is required")
	}
	if !uracfRegex.MatchString(input.URACF) {
		return nil, errors.New("uracf must be exactly 5 uppercase alphanumeric characters")
	}

	g, err := s.guestRepo.GetByPhone(ctx, input.Phone)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, ErrGuestNotFound
	}

	existing, err := s.repo.GetByGuestID(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrAlreadyRegistered
	}

	existingByURACF, err := s.repo.GetByURACF(ctx, input.URACF)
	if err != nil {
		return nil, err
	}
	if existingByURACF != nil {
		return nil, ErrURACFTaken
	}

	u := &User{
		GuestID: &g.ID,
		Role:    "guest",
		URACF:   input.URACF,
	}

	return s.repo.Create(ctx, u)
}

func (s *Service) CheckByPhone(ctx context.Context, phone string) (*CheckResponse, error) {
	if phone == "" {
		return nil, errors.New("phone is required")
	}
	if !phoneRegex.MatchString(phone) {
		return nil, errors.New("phone must be a valid BR mobile number (11 digits: DDD + 9 + 8 digits)")
	}

	g, err := s.guestRepo.GetByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return &CheckResponse{Exists: false}, nil
	}

	u, err := s.repo.GetByGuestID(ctx, g.ID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return &CheckResponse{Exists: false}, nil
	}

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
		slog.Error("seed: failed to check existing user", "role", role, "error", err)
		return
	}
	if existing != nil {
		slog.Info("seed: user already exists", "role", role, "uracf", data.URACF)
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
