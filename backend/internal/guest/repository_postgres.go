package guest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
)

const guestColumns = `id, first_name, last_name, relationship, confirmed, family_group, created_by, updated_by, created_at, updated_at`

func scanGuest(row pgx.Row) (Guest, error) {
	var g Guest
	err := row.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt)
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

func (r *PostgresRepository) List(ctx context.Context, limit, offset int, userRACF string) ([]Guest, int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+guestColumns+`, COUNT(*) OVER() AS total
		 FROM guests WHERE created_by = $1 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`, userRACF, limit, offset)
	if err != nil {
		slog.Error("guest.repo list: query failed", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var guests []Guest
	var total int
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt, &total); err != nil {
			slog.Error("guest.repo list: scan failed", "error", err)
			return nil, 0, err
		}
		guests = append(guests, g)
	}

	if guests == nil {
		guests = []Guest{}
	}

	return guests, total, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64, userRACF string) (*Guest, error) {
	g, err := scanGuest(r.db.QueryRow(ctx,
		`SELECT `+guestColumns+` FROM guests WHERE id = $1 AND created_by = $2`, id, userRACF))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("guest not found")
		}
		slog.Error("guest.repo get_by_id: query failed", "id", id, "error", err)
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) GetByName(ctx context.Context, firstName, lastName string) (*Guest, error) {
	g, err := scanGuest(r.db.QueryRow(ctx,
		`SELECT `+guestColumns+` FROM guests WHERE first_name = $1 AND last_name = $2`, firstName, lastName))
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
	err := r.db.QueryRow(ctx,
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
	err := r.db.QueryRow(ctx, `SELECT COALESCE(MAX(family_group), 0) + 1 FROM guests`).Scan(&nextFamilyGroup)
	if err != nil {
		slog.Error("guest.repo get_next_family_group: query failed", "error", err)
		return 0, err
	}

	return nextFamilyGroup, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateGuestInput, userRACF string) (*Guest, error) {
	g, err := scanGuest(r.db.QueryRow(ctx,
		`INSERT INTO guests (first_name, last_name, relationship, family_group, created_by, updated_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+guestColumns,
		input.FirstName, input.LastName, input.Relationship, *input.FamilyGroup, userRACF, userRACF))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, apperror.Conflict(fmt.Sprintf("a guest named '%s %s' already exists", input.FirstName, input.LastName))
		}
		slog.Error("guest.repo create: insert failed", "error", err)
		return nil, err
	}
	slog.Info("guest.repo create: guest stored", "id", g.ID)
	return &g, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id int64, input UpdateGuestInput, userRACF string) (*Guest, error) {
	g, err := scanGuest(r.db.QueryRow(ctx,
		`UPDATE guests SET
			first_name = COALESCE($1, first_name),
			last_name = COALESCE($2, last_name),
			relationship = COALESCE($3, relationship),
			confirmed = COALESCE($4, confirmed),
			family_group = COALESCE($5, family_group),
			updated_by = $6,
			updated_at = now()
		 WHERE id = $7
		 RETURNING `+guestColumns,
		input.FirstName, input.LastName, input.Relationship, input.Confirmed, input.FamilyGroup, userRACF, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("guest not found")
		}
		slog.Error("guest.repo update: update failed", "id", id, "error", err)
		return nil, err
	}
	slog.Info("guest.repo update: guest updated", "id", g.ID)
	return &g, nil
}

func (r *PostgresRepository) SetConfirmed(ctx context.Context, id int64, confirmed bool, userRACF string) (*Guest, error) {
	g, err := scanGuest(r.db.QueryRow(ctx,
		`UPDATE guests SET confirmed = $1, updated_by = $2, updated_at = now()
		 WHERE id = $3
		 RETURNING `+guestColumns,
		confirmed, userRACF, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("guest not found")
		}
		slog.Error("guest.repo set_confirmed: update failed", "id", id, "error", err)
		return nil, err
	}
	slog.Info("guest.repo set_confirmed: guest updated", "id", g.ID, "confirmed", confirmed)
	return &g, nil
}

func (r *PostgresRepository) SetConfirmedByFamilyGroup(ctx context.Context, familyGroup int64, confirmed bool, userRACF string) ([]Guest, error) {
	rows, err := r.db.Query(ctx,
		`UPDATE guests SET confirmed = $1, updated_by = $2, updated_at = now()
		 WHERE family_group = $3
		 RETURNING `+guestColumns,
		confirmed, userRACF, familyGroup)
	if err != nil {
		slog.Error("guest.repo set_confirmed_by_family_group: update failed", "family_group", familyGroup, "error", err)
		return nil, err
	}
	defer rows.Close()

	var guests []Guest
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Confirmed, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			slog.Error("guest.repo set_confirmed_by_family_group: scan failed", "error", err)
			return nil, err
		}
		guests = append(guests, g)
	}

	if guests == nil {
		guests = []Guest{}
	}

	slog.Info("guest.repo set_confirmed_by_family_group: guests updated", "family_group", familyGroup, "confirmed", confirmed, "count", len(guests))
	return guests, rows.Err()
}

func (r *PostgresRepository) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM guests WHERE id = $1`, id)
	if err != nil {
		slog.Error("guest.repo delete: delete failed", "id", id, "error", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("guest not found")
	}
	slog.Info("guest.repo delete: guest deleted", "id", id)
	return nil
}

func (r *PostgresRepository) GetFamilyGroupByPhone(ctx context.Context, phone string) (*int64, error) {
	var familyGroup int64
	err := r.db.QueryRow(ctx,
		`SELECT g.family_group FROM guests g
		 JOIN guest_users u ON u.guest_id = g.id
		 WHERE u.phone = $1`, phone).Scan(&familyGroup)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		slog.Error("guest.repo get_family_group_by_phone: query failed", "phone", phone, "error", err)
		return nil, err
	}
	return &familyGroup, nil
}
