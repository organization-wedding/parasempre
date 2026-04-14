package auth

import (
	"context"
	"time"

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
