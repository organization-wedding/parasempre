package giftmessage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
)

const messageColumns = `id, gift_transaction_id, gift_id, user_id, author_name, content, media_object_key, media_kind, media_size_bytes, media_mime_type, created_at, updated_at, deleted_at, deleted_by`

func scanMessage(row pgx.Row) (GiftMessage, error) {
	var m GiftMessage
	err := row.Scan(
		&m.ID, &m.GiftTransactionID, &m.GiftID, &m.UserID,
		&m.AuthorName, &m.Content,
		&m.MediaObjectKey, &m.MediaKind, &m.MediaSizeBytes, &m.MediaMimeType,
		&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.DeletedBy,
	)
	return m, err
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

func (r *PostgresRepository) Create(ctx context.Context, in CreateRow) (*GiftMessage, error) {
	m, err := scanMessage(r.db.QueryRow(ctx,
		`INSERT INTO gift_messages
		    (gift_transaction_id, gift_id, user_id, author_name, content,
		     media_object_key, media_kind, media_size_bytes, media_mime_type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING `+messageColumns,
		in.GiftTransactionID, in.GiftID, in.UserID, in.AuthorName, in.Content,
		in.MediaObjectKey, in.MediaKind, in.MediaSizeBytes, in.MediaMimeType,
	))
	if err != nil {
		if appErr := mapPgError(err); appErr != nil {
			return nil, appErr
		}
		slog.Error("giftmessage.repo create: insert failed", "error", err)
		return nil, err
	}
	slog.Info("giftmessage.repo create: stored", "id", m.ID, "tx_id", m.GiftTransactionID)
	return &m, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*GiftMessage, error) {
	m, err := scanMessage(r.db.QueryRow(ctx,
		`SELECT `+messageColumns+` FROM gift_messages WHERE id = $1 AND deleted_at IS NULL`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("mensagem não encontrada")
		}
		slog.Error("giftmessage.repo get_by_id: query failed", "id", id, "error", err)
		return nil, err
	}
	return &m, nil
}

func (r *PostgresRepository) GetByTransactionID(ctx context.Context, txID int64) (*GiftMessage, error) {
	m, err := scanMessage(r.db.QueryRow(ctx,
		`SELECT `+messageColumns+` FROM gift_messages
		  WHERE gift_transaction_id = $1 AND deleted_at IS NULL`, txID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("mensagem não encontrada")
		}
		slog.Error("giftmessage.repo get_by_tx: query failed", "tx_id", txID, "error", err)
		return nil, err
	}
	return &m, nil
}

func (r *PostgresRepository) ListByGift(ctx context.Context, giftID int64, limit, offset int) ([]GiftMessage, int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+messageColumns+`, COUNT(*) OVER() AS total
		   FROM gift_messages
		  WHERE gift_id = $1 AND deleted_at IS NULL
		  ORDER BY created_at DESC
		  LIMIT $2 OFFSET $3`,
		giftID, limit, offset)
	if err != nil {
		slog.Error("giftmessage.repo list_by_gift: query failed", "gift_id", giftID, "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var msgs []GiftMessage
	var total int
	for rows.Next() {
		var m GiftMessage
		if err := rows.Scan(
			&m.ID, &m.GiftTransactionID, &m.GiftID, &m.UserID,
			&m.AuthorName, &m.Content,
			&m.MediaObjectKey, &m.MediaKind, &m.MediaSizeBytes, &m.MediaMimeType,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.DeletedBy,
			&total,
		); err != nil {
			slog.Error("giftmessage.repo list_by_gift: scan failed", "error", err)
			return nil, 0, err
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []GiftMessage{}
	}
	return msgs, total, rows.Err()
}

func (r *PostgresRepository) ListAll(ctx context.Context, limit, offset int) ([]GiftMessage, int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+messageColumns+`, COUNT(*) OVER() AS total
		   FROM gift_messages
		  WHERE deleted_at IS NULL
		  ORDER BY created_at DESC
		  LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		slog.Error("giftmessage.repo list_all: query failed", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var msgs []GiftMessage
	var total int
	for rows.Next() {
		var m GiftMessage
		if err := rows.Scan(
			&m.ID, &m.GiftTransactionID, &m.GiftID, &m.UserID,
			&m.AuthorName, &m.Content,
			&m.MediaObjectKey, &m.MediaKind, &m.MediaSizeBytes, &m.MediaMimeType,
			&m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.DeletedBy,
			&total,
		); err != nil {
			slog.Error("giftmessage.repo list_all: scan failed", "error", err)
			return nil, 0, err
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []GiftMessage{}
	}
	return msgs, total, rows.Err()
}

func (r *PostgresRepository) SoftDelete(ctx context.Context, id, byUserID int64) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE gift_messages
		    SET deleted_at = now(), deleted_by = $1, updated_at = now()
		  WHERE id = $2 AND deleted_at IS NULL`,
		byUserID, id)
	if err != nil {
		slog.Error("giftmessage.repo soft_delete: failed", "id", id, "error", err)
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperror.NotFound("mensagem não encontrada")
	}
	slog.Info("giftmessage.repo soft_delete: removed", "id", id, "by", byUserID)
	return nil
}

func mapPgError(err error) *apperror.AppError {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		if pgErr.ConstraintName == "gift_messages_gift_transaction_id_key" {
			return apperror.Conflict("já existe uma mensagem para essa transação")
		}
		return apperror.Conflict("mensagem em conflito com registro existente")
	case pgerrcode.CheckViolation:
		return apperror.Validation(fmt.Sprintf("dados inválidos para mensagem (%s)", pgErr.ConstraintName))
	case pgerrcode.ForeignKeyViolation:
		return apperror.Validation("transação, presente ou usuário referenciado não existe")
	}
	return nil
}
