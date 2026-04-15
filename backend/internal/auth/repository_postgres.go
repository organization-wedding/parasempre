package auth

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ferjunior7/parasempre/backend/internal/database"
)

type PostgresOTPRepository struct {
	db database.DBTX
}

func NewPostgresOTPRepository(pool *pgxpool.Pool) *PostgresOTPRepository {
	return &PostgresOTPRepository{db: pool}
}

func (r *PostgresOTPRepository) Create(ctx context.Context, phone, code string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO otp_codes (phone, code, expires_at) VALUES ($1, $2, $3)`,
		phone, code, expiresAt)
	return err
}

func (r *PostgresOTPRepository) VerifyAndMarkUsed(ctx context.Context, phone, code string) (bool, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE otp_codes SET used = true
		 WHERE phone = $1 AND code = $2 AND expires_at > now() AND used = false`,
		phone, code)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *PostgresOTPRepository) SendCooldown(ctx context.Context, phone string) (time.Duration, error) {
	const query = `
		(SELECT CEIL(EXTRACT(EPOCH FROM (created_at + interval '60 seconds' - now())))::int AS wait_seconds
		 FROM otp_codes
		 WHERE phone = $1 AND created_at > now() - interval '60 seconds'
		 ORDER BY created_at DESC
		 LIMIT 1)
		UNION ALL
		(SELECT CEIL(EXTRACT(EPOCH FROM (created_at + interval '1 hour' - now())))::int AS wait_seconds
		 FROM otp_codes
		 WHERE phone = $1 AND created_at > now() - interval '1 hour'
		 ORDER BY created_at DESC
		 OFFSET 9 LIMIT 1)
		ORDER BY wait_seconds DESC
		LIMIT 1`

	var seconds int
	err := r.db.QueryRow(ctx, query, phone).Scan(&seconds)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}
