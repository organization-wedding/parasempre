package guest

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) List(ctx context.Context) ([]Guest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, first_name, last_name, phone, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at
		 FROM guests ORDER BY created_at DESC`)
	if err != nil {
		slog.Error("guest.repo list: query failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	var guests []Guest
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Phone, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			slog.Error("guest.repo list: scan failed", "error", err)
			return nil, err
		}
		guests = append(guests, g)
	}

	if guests == nil {
		guests = []Guest{}
	}

	return guests, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*Guest, error) {
	var g Guest
	err := r.pool.QueryRow(ctx,
		`SELECT id, first_name, last_name, phone, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at
		 FROM guests WHERE id = $1`, id).
		Scan(&g.ID, &g.FirstName, &g.LastName, &g.Phone, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("guest.repo get_by_id: not found", "id", id)
			return nil, ErrNotFound
		}
		slog.Error("guest.repo get_by_id: query failed", "id", id, "error", err)
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) GetByPhone(ctx context.Context, phone string) (*Guest, error) {
	var g Guest
	err := r.pool.QueryRow(ctx,
		`SELECT id, first_name, last_name, phone, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at
		 FROM guests WHERE phone = $1`, phone).
		Scan(&g.ID, &g.FirstName, &g.LastName, &g.Phone, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("guest.repo get_by_phone: query failed", "phone", phone, "error", err)
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) GetByName(ctx context.Context, firstName, lastName string) (*Guest, error) {
	var g Guest
	err := r.pool.QueryRow(ctx,
		`SELECT id, first_name, last_name, phone, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at
		 FROM guests WHERE first_name = $1 AND last_name = $2`, firstName, lastName).
		Scan(&g.ID, &g.FirstName, &g.LastName, &g.Phone, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("guest.repo get_by_name: query failed", "first_name", firstName, "last_name", lastName, "error", err)
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) FamilyGroupExists(ctx context.Context, familyGroup int64) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM guests WHERE family_group = $1)`, familyGroup).
		Scan(&exists)
	if err != nil {
		slog.Error("guest.repo family_group_exists: query failed", "family_group", familyGroup, "error", err)
		return false, err
	}

	return exists, nil
}

func (r *PostgresRepository) GetNextFamilyGroup(ctx context.Context) (int64, error) {
	var nextFamilyGroup int64
	err := r.pool.QueryRow(ctx, `SELECT COALESCE(MAX(family_group), 0) + 1 FROM guests`).Scan(&nextFamilyGroup)
	if err != nil {
		slog.Error("guest.repo get_next_family_group: query failed", "error", err)
		return 0, err
	}

	return nextFamilyGroup, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
	var phone *string
	if input.Phone != "" {
		phone = &input.Phone
	}

	var g Guest
	err := r.pool.QueryRow(ctx,
		`INSERT INTO guests (first_name, last_name, phone, relationship, family_group, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, first_name, last_name, phone, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at`,
		input.FirstName, input.LastName, phone, input.Relationship, *input.FamilyGroup, userRACF, userRACF).
		Scan(&g.ID, &g.FirstName, &g.LastName, &g.Phone, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		slog.Error("guest.repo create: insert failed", "error", err)
		return nil, err
	}
	slog.Info("guest.repo create: guest stored", "id", g.ID)
	return &g, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	var g Guest
	err := r.pool.QueryRow(ctx,
		`UPDATE guests SET
			first_name = COALESCE($1, first_name),
			last_name = COALESCE($2, last_name),
			phone = COALESCE($3, phone),
			relationship = COALESCE($4, relationship),
			confirmed = COALESCE($5, confirmed),
			family_group = COALESCE($6, family_group),
			updated_by = $7,
			updated_at = now()
		 WHERE id = $8
		 RETURNING id, first_name, last_name, phone, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at`,
		input.FirstName, input.LastName, input.Phone, input.Relationship, input.Confirmed, input.FamilyGroup, userRACF, id).
		Scan(&g.ID, &g.FirstName, &g.LastName, &g.Phone, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("guest.repo update: not found", "id", id)
			return nil, ErrNotFound
		}
		slog.Error("guest.repo update: update failed", "id", id, "error", err)
		return nil, err
	}
	slog.Info("guest.repo update: guest updated", "id", g.ID)
	return &g, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM guests WHERE id = $1`, id)
	if err != nil {
		slog.Error("guest.repo delete: delete failed", "id", id, "error", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		slog.Warn("guest.repo delete: not found", "id", id)
		return ErrNotFound
	}
	slog.Info("guest.repo delete: guest deleted", "id", id)
	return nil
}
