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
		`SELECT id, nome, sobrenome, telefone, relacionamento, confirmacao, created_at, updated_at
		 FROM guests ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guests []Guest
	for rows.Next() {
		var g Guest
		if err := rows.Scan(&g.ID, &g.Nome, &g.Sobrenome, &g.Telefone, &g.Relacionamento, &g.Confirmacao, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		guests = append(guests, g)
	}

	if guests == nil {
		guests = []Guest{}
	}

	return guests, rows.Err()
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (*Guest, error) {
	var g Guest
	err := r.pool.QueryRow(ctx,
		`SELECT id, nome, sobrenome, telefone, relacionamento, confirmacao, created_at, updated_at
		 FROM guests WHERE id = $1`, id).
		Scan(&g.ID, &g.Nome, &g.Sobrenome, &g.Telefone, &g.Relacionamento, &g.Confirmacao, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateGuestInput) (*Guest, error) {
	var g Guest
	err := r.pool.QueryRow(ctx,
		`INSERT INTO guests (nome, sobrenome, telefone, relacionamento)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, nome, sobrenome, telefone, relacionamento, confirmacao, created_at, updated_at`,
		input.Nome, input.Sobrenome, input.Telefone, input.Relacionamento).
		Scan(&g.ID, &g.Nome, &g.Sobrenome, &g.Telefone, &g.Relacionamento, &g.Confirmacao, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) Update(ctx context.Context, id string, input UpdateGuestInput) (*Guest, error) {
	// Build dynamic update query using COALESCE pattern.
	var g Guest
	err := r.pool.QueryRow(ctx,
		`UPDATE guests SET
			nome = COALESCE($1, nome),
			sobrenome = COALESCE($2, sobrenome),
			telefone = COALESCE($3, telefone),
			relacionamento = COALESCE($4, relacionamento),
			confirmacao = COALESCE($5, confirmacao),
			updated_at = now()
		 WHERE id = $6
		 RETURNING id, nome, sobrenome, telefone, relacionamento, confirmacao, created_at, updated_at`,
		input.Nome, input.Sobrenome, input.Telefone, input.Relacionamento, input.Confirmacao, id).
		Scan(&g.ID, &g.Nome, &g.Sobrenome, &g.Telefone, &g.Relacionamento, &g.Confirmacao, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM guests WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
