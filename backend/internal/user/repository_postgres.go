package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) GetByURACF(ctx context.Context, uracf string) (*User, error) {
	var u User
	err := r.pool.QueryRow(ctx,
		`SELECT id, guest_id, role, uracf, created_at, updated_at
		 FROM users WHERE uracf = $1`, uracf).
		Scan(&u.ID, &u.GuestID, &u.Role, &u.URACF, &u.CreatedAt, &u.UpdatedAt)
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
	err := r.pool.QueryRow(ctx,
		`SELECT id, guest_id, role, uracf, created_at, updated_at
		 FROM users WHERE guest_id = $1`, guestID).
		Scan(&u.ID, &u.GuestID, &u.Role, &u.URACF, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]UserListItem, error) {
	rows, err := r.pool.Query(ctx, `
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

func (r *PostgresRepository) Create(ctx context.Context, u *User) (*User, error) {
	var created User
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (guest_id, role, uracf)
		 VALUES ($1, $2, $3)
		 RETURNING id, guest_id, role, uracf, created_at, updated_at`,
		u.GuestID, u.Role, u.URACF).
		Scan(&created.ID, &created.GuestID, &created.Role, &created.URACF, &created.CreatedAt, &created.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &created, nil
}
