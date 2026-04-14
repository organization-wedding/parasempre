package user

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ferjunior7/parasempre/backend/internal/database"
)

type PostgresRepository struct {
	db database.DBTX
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: pool}
}

func (r *PostgresRepository) WithTx(tx pgx.Tx) Repository {
	return &PostgresRepository{db: tx}
}

func (r *PostgresRepository) GetByURACF(ctx context.Context, uracf string) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx,
		`SELECT id, guest_id, role, uracf, phone, last_login_at, created_at, updated_at
		 FROM users WHERE uracf = $1`, uracf).
		Scan(&u.ID, &u.GuestID, &u.Role, &u.URACF, &u.Phone, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetByGuestID(ctx context.Context, guestID int64) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx,
		`SELECT id, guest_id, role, uracf, phone, last_login_at, created_at, updated_at
		 FROM users WHERE guest_id = $1`, guestID).
		Scan(&u.ID, &u.GuestID, &u.Role, &u.URACF, &u.Phone, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]UserListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.uracf, u.role,
		       COALESCE(g.first_name, '') AS first_name,
		       COALESCE(g.last_name,  '') AS last_name
		FROM users u
		LEFT JOIN guests g ON g.id = u.guest_id
		ORDER BY
			CASE u.role WHEN 'groom' THEN 0 WHEN 'bride' THEN 1 ELSE 2 END,
			u.uracf
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []UserListItem
	for rows.Next() {
		var item UserListItem
		if err := rows.Scan(&item.URACF, &item.Role, &item.FirstName, &item.LastName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []UserListItem{}
	}
	return items, rows.Err()
}

func (r *PostgresRepository) GetByPhone(ctx context.Context, phone string) (*User, error) {
	var u User
	err := r.db.QueryRow(ctx,
		`SELECT id, guest_id, role, uracf, phone, last_login_at, created_at, updated_at
		 FROM users WHERE phone = $1`, phone).
		Scan(&u.ID, &u.GuestID, &u.Role, &u.URACF, &u.Phone, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) Create(ctx context.Context, u *User) (*User, error) {
	var created User
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (guest_id, role, uracf, phone)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, guest_id, role, uracf, phone, last_login_at, created_at, updated_at`,
		u.GuestID, u.Role, u.URACF, u.Phone).
		Scan(&created.ID, &created.GuestID, &created.Role, &created.URACF, &created.Phone, &created.LastLoginAt, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}

func (r *PostgresRepository) DeleteByGuestID(ctx context.Context, guestID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE guest_id = $1`, guestID)
	return err
}

func (r *PostgresRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_login_at = now(), updated_at = now() WHERE id = $1`, userID)
	return err
}

func (r *PostgresRepository) LogAction(ctx context.Context, userID int64, action string, details map[string]any) error {
	var detailsJSON []byte
	if details != nil {
		var err error
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return err
		}
	}
	_, err := r.db.Exec(ctx,
		`INSERT INTO audit_log (user_id, action, details) VALUES ($1, $2, $3)`,
		userID, action, detailsJSON)
	return err
}
