package gift

import (
	"context"
	"log/slog"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

type Service struct {
	repo TxAwareRepository
}

func NewService(repo TxAwareRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, page, limit int, status *string) (*PagedResponse, error) {
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
	gifts, total, err := s.repo.List(ctx, limit, offset, status)
	if err != nil {
		slog.Error("gift.service list: failed", "error", err)
		return nil, apperror.Internal("failed to list gifts", err)
	}
	return &PagedResponse{
		Data:  gifts,
		Page:  page,
		Limit: limit,
		Total: total,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Gift, error) {
	g, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to get gift", err)
	}
	return g, nil
}

func (s *Service) Create(ctx context.Context, input CreateGiftInput, userRACF string) (*Gift, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	dedupeKey := NormalizeDedupeKey(input.Name)
	g, err := s.repo.Create(ctx, input, dedupeKey, userRACF)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to create gift", err)
	}

	slog.Info("gift.service create: gift created", "id", g.ID, "user_racf", userRACF)
	return g, nil
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateGiftInput, userRACF string) (*Gift, error) {
	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	var dedupeKey *string
	if input.Name != nil {
		k := NormalizeDedupeKey(*input.Name)
		dedupeKey = &k
	}

	g, err := s.repo.Update(ctx, id, input, dedupeKey, userRACF)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to update gift", err)
	}

	slog.Info("gift.service update: gift updated", "id", g.ID, "user_racf", userRACF)
	return g, nil
}

func (s *Service) Delete(ctx context.Context, id int64, userRACF string) error {
	if err := s.repo.Delete(ctx, id, userRACF); err != nil {
		return apperror.WrapIfNotApp("failed to delete gift", err)
	}
	slog.Info("gift.service delete: gift soft-deleted", "id", id, "user_racf", userRACF)
	return nil
}
