package guest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
)

const guestColumns = `id, first_name, last_name, relationship, attending, family_group, created_by, updated_by, created_at, updated_at`

func scanGuest(row pgx.Row) (Guest, error) {
	var g Guest
	err := row.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Attending, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt)
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

// List returns the wedding's shared guest list. Guests belong to the wedding
// (co-administered by groom and bride), not to whoever created the row, so the
// listing is intentionally NOT scoped by created_by.
func (r *PostgresRepository) List(ctx context.Context, limit, offset int, filters ListFilters) ([]Guest, int, error) {
	var conds []string
	var args []any
	n := 1

	if filters.Search != "" {
		conds = append(conds, fmt.Sprintf("(first_name || ' ' || last_name) ILIKE $%d", n))
		args = append(args, "%"+filters.Search+"%")
		n++
	}
	if filters.Relationship == "P" || filters.Relationship == "R" {
		conds = append(conds, fmt.Sprintf("relationship = $%d", n))
		args = append(args, filters.Relationship)
		n++
	}
	switch filters.Attending {
	case "attending":
		conds = append(conds, "attending IS TRUE")
	case "declined":
		conds = append(conds, "attending IS FALSE")
	case "pending":
		conds = append(conds, "attending IS NULL")
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ") + " "
	}
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx,
		`SELECT `+guestColumns+`, COUNT(*) OVER() AS total
		 FROM guests `+where+`ORDER BY created_at DESC
		 LIMIT $`+strconv.Itoa(n)+` OFFSET $`+strconv.Itoa(n+1), args...)
	if err != nil {
		slog.Error("guest.repo list: query failed", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var guests []Guest
	var total int
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Attending, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt, &total); err != nil {
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

func (r *PostgresRepository) Stats(ctx context.Context) (Stats, error) {
	var s Stats
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*),
		        COUNT(*) FILTER (WHERE attending IS TRUE),
		        COUNT(*) FILTER (WHERE attending IS NULL),
		        COUNT(*) FILTER (WHERE attending IS FALSE)
		 FROM guests`).Scan(&s.Total, &s.Confirmed, &s.Pending, &s.Declined)
	if err != nil {
		slog.Error("guest.repo stats: query failed", "error", err)
		return Stats{}, err
	}
	return s, nil
}

func (r *PostgresRepository) GetByIDAny(ctx context.Context, id int64) (*Guest, error) {
	g, err := scanGuest(r.db.QueryRow(ctx,
		`SELECT `+guestColumns+` FROM guests WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("guest not found")
		}
		slog.Error("guest.repo get_by_id_any: query failed", "id", id, "error", err)
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) ListByFamilyGroup(ctx context.Context, familyGroup int64) ([]Guest, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+guestColumns+` FROM guests WHERE family_group = $1 ORDER BY id`, familyGroup)
	if err != nil {
		slog.Error("guest.repo list_by_family_group: query failed", "family_group", familyGroup, "error", err)
		return nil, err
	}
	defer rows.Close()

	guests := []Guest{}
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Attending, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			slog.Error("guest.repo list_by_family_group: scan failed", "error", err)
			return nil, err
		}
		guests = append(guests, g)
	}
	return guests, rows.Err()
}

func (r *PostgresRepository) GetByIDs(ctx context.Context, ids []int64) ([]Guest, error) {
	if len(ids) == 0 {
		return []Guest{}, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+guestColumns+` FROM guests WHERE id = ANY($1::bigint[])`, ids)
	if err != nil {
		slog.Error("guest.repo get_by_ids: query failed", "ids", ids, "error", err)
		return nil, err
	}
	defer rows.Close()

	guests := []Guest{}
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Attending, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			slog.Error("guest.repo get_by_ids: scan failed", "error", err)
			return nil, err
		}
		guests = append(guests, g)
	}
	return guests, rows.Err()
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
			attending = COALESCE($4, attending),
			family_group = COALESCE($5, family_group),
			updated_by = $6,
			updated_at = now()
		 WHERE id = $7
		 RETURNING `+guestColumns,
		input.FirstName, input.LastName, input.Relationship, input.Attending, input.FamilyGroup, userRACF, id))
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

func (r *PostgresRepository) SetAttending(ctx context.Context, id int64, attending bool, userRACF string) (*Guest, error) {
	g, err := scanGuest(r.db.QueryRow(ctx,
		`UPDATE guests SET attending = $1, updated_by = $2, updated_at = now()
		 WHERE id = $3
		 RETURNING `+guestColumns,
		attending, userRACF, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("guest not found")
		}
		slog.Error("guest.repo set_attending: update failed", "id", id, "error", err)
		return nil, err
	}
	slog.Info("guest.repo set_attending: guest updated", "id", g.ID, "attending", attending)
	return &g, nil
}

func (r *PostgresRepository) SetAttendingByIDs(ctx context.Context, ids []int64, attending bool, userRACF string) ([]Guest, error) {
	if len(ids) == 0 {
		return []Guest{}, nil
	}
	rows, err := r.db.Query(ctx,
		`UPDATE guests SET attending = $1, updated_by = $2, updated_at = now()
		 WHERE id = ANY($3::bigint[])
		 RETURNING `+guestColumns,
		attending, userRACF, ids)
	if err != nil {
		slog.Error("guest.repo set_attending_by_ids: update failed", "ids", ids, "error", err)
		return nil, err
	}
	defer rows.Close()

	guests := []Guest{}
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Attending, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			slog.Error("guest.repo set_attending_by_ids: scan failed", "error", err)
			return nil, err
		}
		guests = append(guests, g)
	}
	slog.Info("guest.repo set_attending_by_ids: guests updated", "ids", ids, "attending", attending, "count", len(guests))
	return guests, rows.Err()
}

func (r *PostgresRepository) SetAttendingByFamilyGroup(ctx context.Context, familyGroup int64, attending bool, userRACF string) ([]Guest, error) {
	rows, err := r.db.Query(ctx,
		`UPDATE guests SET attending = $1, updated_by = $2, updated_at = now()
		 WHERE family_group = $3
		 RETURNING `+guestColumns,
		attending, userRACF, familyGroup)
	if err != nil {
		slog.Error("guest.repo set_attending_by_family_group: update failed", "family_group", familyGroup, "error", err)
		return nil, err
	}
	defer rows.Close()

	var guests []Guest
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.FirstName, &g.LastName, &g.Relationship, &g.Attending, &g.FamilyGroup, &g.CreatedBy, &g.UpdatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			slog.Error("guest.repo set_attending_by_family_group: scan failed", "error", err)
			return nil, err
		}
		guests = append(guests, g)
	}

	if guests == nil {
		guests = []Guest{}
	}

	slog.Info("guest.repo set_attending_by_family_group: guests updated", "family_group", familyGroup, "attending", attending, "count", len(guests))
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
