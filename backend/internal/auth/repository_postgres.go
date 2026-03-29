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

func (r *PostgresOTPRepository) FindValid(ctx context.Context, phone, code string) (*OTPRecord, error) {
	var rec OTPRecord
	err := r.db.QueryRow(ctx,
		`SELECT id, phone, code, expires_at, used
		 FROM otp_codes
		 WHERE phone = $1 AND code = $2 AND used = false AND expires_at > now()
		 ORDER BY created_at DESC
		 LIMIT 1`, phone, code).
		Scan(&rec.ID, &rec.Phone, &rec.Code, &rec.ExpiresAt, &rec.Used)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *PostgresOTPRepository) MarkUsed(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE otp_codes SET used = true WHERE id = $1`, id)
	return err
}
