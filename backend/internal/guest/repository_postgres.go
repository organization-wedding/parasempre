package guest

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

func (r *PostgresRepository) List(ctx context.Context) ([]Guest, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, first_name, last_name, phone, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at
		 FROM guests ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guests []Guest
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Phone, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
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
			return nil, ErrNotFound
		}
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
		return nil, err
	}
	return &g, nil
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
		return nil, err
	}
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
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM guests WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
