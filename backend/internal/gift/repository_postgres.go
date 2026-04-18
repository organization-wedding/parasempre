package gift

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
)

const giftColumns = `id, name, description, price_cents, image_url, store_url, status, dedupe_key, created_by, updated_by, created_at, updated_at, deleted_at`

func scanGift(row pgx.Row) (Gift, error) {
	var g Gift
	err := row.Scan(&g.ID, &g.Name, &g.Description, &g.PriceCents, &g.ImageURL, &g.StoreURL, &g.Status, &g.DedupeKey, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt, &g.DeletedAt)
	return g, err
}

type PostgresRepository struct {
	db database.DBTX
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: pool}
}

func (r *PostgresRepository) WithTx(tx pgx.Tx) Repository {
	return &PostgresRepository{db: tx}
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int, statusFilter *string) ([]Gift, int, error) {
	query := `SELECT ` + giftColumns + `, COUNT(*) OVER() AS total FROM gifts WHERE deleted_at IS NULL`
	args := []any{}
	if statusFilter != nil {
		query += ` AND status = $1`
		args = append(args, *statusFilter)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprint(len(args)+1) + ` OFFSET $` + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		slog.Error("gift.repo list: query failed", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var gifts []Gift
	var total int
	for rows.Next() {
		var g Gift
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.PriceCents, &g.ImageURL, &g.StoreURL, &g.Status, &g.DedupeKey, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt, &g.DeletedAt, &total); err != nil {
			slog.Error("gift.repo list: scan failed", "error", err)
			return nil, 0, err
		}
		gifts = append(gifts, g)
	}

	if gifts == nil {
		gifts = []Gift{}
	}

	return gifts, total, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*Gift, error) {
	g, err := scanGift(r.db.QueryRow(ctx,
		`SELECT `+giftColumns+` FROM gifts WHERE id = $1 AND deleted_at IS NULL`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("gift not found")
		}
		slog.Error("gift.repo get_by_id: query failed", "id", id, "error", err)
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateGiftInput, dedupeKey, userRACF string) (*Gift, error) {
	status := "active"
	if input.Status != nil {
		status = *input.Status
	}

	g, err := scanGift(r.db.QueryRow(ctx,
		`INSERT INTO gifts (name, description, price_cents, image_url, store_url, status, dedupe_key, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING `+giftColumns,
		input.Name, input.Description, input.PriceCents, input.ImageURL, input.StoreURL, status, dedupeKey, userRACF, userRACF))
	if err != nil {
		if appErr := mapPgError(err); appErr != nil {
			return nil, appErr
		}
		slog.Error("gift.repo create: insert failed", "error", err)
		return nil, err
	}
	slog.Info("gift.repo create: gift stored", "id", g.ID)
	return &g, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id int64, input UpdateGiftInput, dedupeKey *string, userRACF string) (*Gift, error) {
	g, err := scanGift(r.db.QueryRow(ctx,
		`UPDATE gifts SET
			name = COALESCE($1, name),
			description = COALESCE($2, description),
			price_cents = COALESCE($3, price_cents),
			image_url = COALESCE($4, image_url),
			store_url = COALESCE($5, store_url),
			status = COALESCE($6, status),
			dedupe_key = COALESCE($7, dedupe_key),
			updated_by = $8,
			updated_at = now()
		 WHERE id = $9 AND deleted_at IS NULL
		 RETURNING `+giftColumns,
		input.Name, input.Description, input.PriceCents, input.ImageURL, input.StoreURL, input.Status, dedupeKey, userRACF, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("gift not found")
		}
		if appErr := mapPgError(err); appErr != nil {
			return nil, appErr
		}
		slog.Error("gift.repo update: update failed", "id", id, "error", err)
		return nil, err
	}
	slog.Info("gift.repo update: gift updated", "id", g.ID)
	return &g, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64, userRACF string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE gifts SET deleted_at = now(), updated_by = $1, updated_at = now(), status = 'inactive'
		 WHERE id = $2 AND deleted_at IS NULL`,
		userRACF, id)
	if err != nil {
		slog.Error("gift.repo delete: soft-delete failed", "id", id, "error", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("gift not found")
	}
	slog.Info("gift.repo delete: gift soft-deleted", "id", id, "by", userRACF)
	return nil
}

func (r *PostgresRepository) FindByDedupeKeys(ctx context.Context, keys []string) (map[string]bool, error) {
	found := make(map[string]bool, len(keys))
	if len(keys) == 0 {
		return found, nil
	}

	rows, err := r.db.Query(ctx,
		`SELECT dedupe_key FROM gifts WHERE dedupe_key = ANY($1) AND deleted_at IS NULL`, keys)
	if err != nil {
		slog.Error("gift.repo find_by_dedupe_keys: query failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			slog.Error("gift.repo find_by_dedupe_keys: scan failed", "error", err)
			return nil, err
		}
		found[key] = true
	}

	return found, rows.Err()
}

func mapPgError(err error) *apperror.AppError {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		if pgErr.ConstraintName == "gifts_dedupe_key_unique" {
			return apperror.Conflict("a gift with this name already exists")
		}
		return apperror.Conflict("unique constraint violated")
	case pgerrcode.CheckViolation:
		return apperror.Validation(fmt.Sprintf("invalid gift data: %s", pgErr.ConstraintName))
	}
	return nil
}
