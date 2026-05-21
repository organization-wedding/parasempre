package user

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ferjunior7/parasempre/backend/internal/database"
)

const userColumns = `id, guest_id, role, uracf, phone, last_login_at, created_at, updated_at`

func scanUser(row pgx.Row) (User, error) {
	var u User
	err := row.Scan(&u.ID, &u.GuestID, &u.Role, &u.URACF, &u.Phone, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt)
	return u, err
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

func (r *PostgresRepository) GetByURACF(ctx context.Context, uracf string) (*User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE uracf = $1`, uracf))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user.repo get_by_uracf: query failed", "uracf", uracf, "error", err)
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetMeByURACF(ctx context.Context, uracf string) (*MeResponse, error) {
	var me MeResponse
	err := r.db.QueryRow(ctx,
		`SELECT u.role, u.guest_id, g.first_name, g.last_name, g.family_group
		 FROM users u
		 LEFT JOIN guests g ON g.id = u.guest_id
		 WHERE u.uracf = $1`, uracf).
		Scan(&me.Role, &me.GuestID, &me.FirstName, &me.LastName, &me.FamilyGroup)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user.repo get_me_by_uracf: query failed", "uracf", uracf, "error", err)
		return nil, err
	}
	return &me, nil
}

func (r *PostgresRepository) GetByGuestID(ctx context.Context, guestID int64) (*User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE guest_id = $1`, guestID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user.repo get_by_guest_id: query failed", "guest_id", guestID, "error", err)
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
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE phone = $1`, phone))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user.repo get_by_phone: query failed", "phone", phone, "error", err)
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user.repo get_by_id: query failed", "id", id, "error", err)
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) GetByRole(ctx context.Context, role string) (*User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+userColumns+` FROM users WHERE role = $1`, role))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user.repo get_by_role: query failed", "role", role, "error", err)
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) Create(ctx context.Context, u *User) (*User, error) {
	created, err := scanUser(r.db.QueryRow(ctx,
		`INSERT INTO users (guest_id, role, uracf, phone)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+userColumns,
		u.GuestID, u.Role, u.URACF, u.Phone))
	if err != nil {
		slog.Error("user.repo create: insert failed", "uracf", u.URACF, "error", err)
		return nil, err
	}
	return &created, nil
}

func (r *PostgresRepository) DeleteByGuestID(ctx context.Context, guestID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE guest_id = $1`, guestID)
	if err != nil {
		slog.Error("user.repo delete_by_guest_id: delete failed", "guest_id", guestID, "error", err)
	}
	return err
}

func (r *PostgresRepository) Update(ctx context.Context, id int64, input UpdateInput) (*User, error) {
	u, err := scanUser(r.db.QueryRow(ctx,
		`UPDATE users SET
			role = COALESCE($2, role),
			phone = COALESCE($3, phone),
			updated_at = now()
		WHERE id = $1
		RETURNING `+userColumns, id, input.Role, input.Phone))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("user.repo update: update failed", "id", id, "error", err)
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		slog.Error("user.repo delete: delete failed", "id", id, "error", err)
		return err
	}
	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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
