package gift

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

const (
	scrapeMaxNameLen        = 200
	scrapeMaxDescriptionLen = 2000
)

type ProductScraper interface {
	ScrapeProduct(ctx context.Context, url string) (*ScrapedProduct, error)
}

type Service struct {
	repo     TxAwareRepository
	txRunner database.TxRunner
	scraper  ProductScraper
}

func NewService(repo TxAwareRepository, txRunner database.TxRunner, scraper ProductScraper) *Service {
	return &Service{repo: repo, txRunner: txRunner, scraper: scraper}
}

func (s *Service) List(ctx context.Context, page, limit int, filter ListFilter) (*PagedResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	if filter.Search != nil {
		if q := strings.TrimSpace(*filter.Search); q != "" {
			filter.Search = &q
		} else {
			filter.Search = nil
		}
	}

	if filter.PriceMin != nil && *filter.PriceMin < 0 {
		filter.PriceMin = nil
	}
	if filter.PriceMax != nil && *filter.PriceMax < 0 {
		filter.PriceMax = nil
	}
	if filter.Sort != nil && *filter.Sort != SortPriceAsc && *filter.Sort != SortPriceDesc {
		filter.Sort = nil
	}

	offset := (page - 1) * limit
	gifts, total, err := s.repo.List(ctx, filter, limit, offset)
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

func (s *Service) PreviewImport(ctx context.Context, r io.Reader) (*CSVPreview, error) {
	rows, err := ParseCSVRows(r)
	if err != nil {
		return nil, apperror.Validation(fmt.Sprintf("CSV parse error: %s", err.Error()))
	}

	var keys []string
	for _, row := range rows {
		if len(row.Errors) == 0 && row.DedupeKey != "" {
			keys = append(keys, row.DedupeKey)
		}
	}

	existing := map[string]bool{}
	if len(keys) > 0 {
		found, err := s.repo.FindByDedupeKeys(ctx, keys)
		if err != nil {
			slog.Error("gift.service preview_import: dedupe lookup failed", "error", err)
			return nil, apperror.Internal("failed to check duplicates", err)
		}
		existing = found
	}

	summary := CSVSummary{Total: len(rows)}
	seenInBatch := map[string]bool{}
	for i := range rows {
		if len(rows[i].Errors) > 0 {
			rows[i].Status = CSVRowStatusInvalid
			summary.Invalid++
			continue
		}
		if existing[rows[i].DedupeKey] || seenInBatch[rows[i].DedupeKey] {
			rows[i].Status = CSVRowStatusDuplicate
			summary.Duplicate++
			continue
		}
		rows[i].Status = CSVRowStatusNew
		seenInBatch[rows[i].DedupeKey] = true
		summary.New++
	}

	return &CSVPreview{Rows: rows, Summary: summary}, nil
}

func (s *Service) CommitImport(ctx context.Context, inputs []CreateGiftInput, userRACF string) (*CommitImportResponse, error) {
	if len(inputs) == 0 {
		return nil, apperror.Validation("no rows to import")
	}
	for i, input := range inputs {
		if err := validate.Struct(input); err != nil {
			return nil, apperror.Validation(fmt.Sprintf("row %d: %s", i+1, err.Error()))
		}
	}

	keys := make([]string, len(inputs))
	for i, input := range inputs {
		keys[i] = NormalizeDedupeKey(input.Name)
	}

	existing, err := s.repo.FindByDedupeKeys(ctx, keys)
	if err != nil {
		slog.Error("gift.service commit_import: dedupe lookup failed", "error", err)
		return nil, apperror.Internal("failed to check duplicates", err)
	}

	filteredInputs := make([]CreateGiftInput, 0, len(inputs))
	filteredKeys := make([]string, 0, len(inputs))
	skipped := make([]string, 0)
	seenInBatch := map[string]bool{}
	for i, key := range keys {
		if existing[key] {
			skipped = append(skipped, fmt.Sprintf("%q já existe", inputs[i].Name))
			continue
		}
		if seenInBatch[key] {
			skipped = append(skipped, fmt.Sprintf("%q duplicado no lote", inputs[i].Name))
			continue
		}
		seenInBatch[key] = true
		filteredInputs = append(filteredInputs, inputs[i])
		filteredKeys = append(filteredKeys, key)
	}

	var createdCount int
	if len(filteredInputs) > 0 {
		err = s.txRunner.RunInTx(ctx, func(tx pgx.Tx) error {
			txRepo := s.repo.WithTx(tx)
			created, err := txRepo.BulkCreate(ctx, filteredInputs, filteredKeys, userRACF)
			if err != nil {
				return err
			}
			createdCount = len(created)
			return nil
		})
		if err != nil {
			return nil, apperror.WrapIfNotApp("failed to import gifts", err)
		}
	}

	slog.Info("gift.service commit_import: finished",
		"requested", len(inputs), "created", createdCount, "skipped", len(skipped), "user_racf", userRACF)

	return &CommitImportResponse{
		Created: createdCount,
		Skipped: skipped,
	}, nil
}

func (s *Service) ScrapePreview(ctx context.Context, rawURL, userRACF string) (*ScrapePreviewResponse, error) {
	if s.scraper == nil {
		return nil, apperror.ServiceUnavailable("Busca por link não está configurada neste ambiente.")
	}

	url := strings.TrimSpace(rawURL)
	if !httpsURLRegex.MatchString(url) {
		return nil, apperror.Validation("URL inválida — use uma URL https://")
	}

	scraped, err := s.scraper.ScrapeProduct(ctx, url)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to scrape product URL", err)
	}

	name := truncateRunes(strings.TrimSpace(scraped.Name), scrapeMaxNameLen)
	description := truncateRunes(strings.TrimSpace(scraped.Description), scrapeMaxDescriptionLen)
	imageURL := strings.TrimSpace(scraped.ImageURL)
	if !httpsURLRegex.MatchString(imageURL) {
		imageURL = ""
	}

	if name == "" && imageURL == "" {
		slog.Info("gift.service scrape_preview: empty extraction", "url", url, "user_racf", userRACF)
		return nil, apperror.Validation("Não conseguimos identificar o produto nesta página.")
	}

	priceCents := int64(0)
	priceStr := strings.TrimSpace(scraped.PriceBRL)
	if priceStr != "" {
		normalized := stripBRThousandsSep(priceStr)
		if cents, parseErr := parsePriceBRL(normalized); parseErr == nil && cents > 0 {
			priceCents = cents
		} else {
			slog.Info("gift.service scrape_preview: price parse failed", "url", url, "raw_price", priceStr, "error", parseErr)
		}
	}

	slog.Info("gift.service scrape_preview: success", "url", url, "user_racf", userRACF, "name_len", len(name), "price_cents", priceCents)

	return &ScrapePreviewResponse{
		Name:        name,
		Description: description,
		PriceCents:  priceCents,
		ImageURL:    imageURL,
		StoreURL:    url,
	}, nil
}

func truncateRunes(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}

func stripBRThousandsSep(s string) string {
	if strings.ContainsRune(s, '.') && strings.ContainsRune(s, ',') {
		return strings.ReplaceAll(s, ".", "")
	}
	return s
}
