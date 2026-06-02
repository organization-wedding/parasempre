package payment

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

const txColumns = `id, gift_id, user_id, payment_method, mp_payment_id, mp_preference_id, amount_cents, status, idempotency_key, created_at, updated_at, gift_name_snapshot`
const gtTxColumns = `gt.id, gt.gift_id, gt.user_id, gt.payment_method, gt.mp_payment_id, gt.mp_preference_id, gt.amount_cents, gt.status, gt.idempotency_key, gt.created_at, gt.updated_at, gt.gift_name_snapshot`

func scanTx(row pgx.Row) (GiftTransaction, error) {
	var t GiftTransaction
	err := row.Scan(
		&t.ID, &t.GiftID, &t.UserID, &t.PaymentMethod, &t.MPPaymentID, &t.MPPreferenceID,
		&t.AmountCents, &t.Status, &t.IdempotencyKey, &t.CreatedAt, &t.UpdatedAt, &t.GiftNameSnapshot,
	)
	return t, err
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

func (r *PostgresRepository) Create(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error) {
	t, err := scanTx(r.db.QueryRow(ctx,
		`INSERT INTO gift_transactions
		    (gift_id, user_id, payment_method, amount_cents, status, idempotency_key, gift_name_snapshot)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING `+txColumns,
		input.GiftID, input.UserID, input.PaymentMethod, input.AmountCents, input.Status,
		input.IdempotencyKey, input.GiftNameSnapshot))
	if err != nil {
		if appErr := mapPgError(err); appErr != nil {
			return nil, appErr
		}
		slog.Error("payment.repo create: insert failed", "error", err)
		return nil, err
	}
	slog.Info("payment.repo create: transaction stored", "id", t.ID, "gift_id", t.GiftID)
	return &t, nil
}

func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*GiftTransaction, error) {
	t, err := scanTx(r.db.QueryRow(ctx,
		`SELECT `+txColumns+` FROM gift_transactions WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("transaction not found")
		}
		slog.Error("payment.repo get_by_id: query failed", "id", id, "error", err)
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRepository) GetByMPPaymentID(ctx context.Context, mpPaymentID string) (*GiftTransaction, error) {
	t, err := scanTx(r.db.QueryRow(ctx,
		`SELECT `+txColumns+` FROM gift_transactions WHERE mp_payment_id = $1`, mpPaymentID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("transaction not found")
		}
		slog.Error("payment.repo get_by_mp_payment_id: query failed", "mp_payment_id", mpPaymentID, "error", err)
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRepository) UpdateAfterCreate(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
	var mpID any
	if mpPaymentID == "" {
		mpID = nil
	} else {
		mpID = mpPaymentID
	}
	t, err := scanTx(r.db.QueryRow(ctx,
		`UPDATE gift_transactions
		 SET mp_payment_id = $1, status = $2, updated_at = now()
		 WHERE id = $3
		 RETURNING `+txColumns,
		mpID, status, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("transaction not found")
		}
		if appErr := mapPgError(err); appErr != nil {
			return nil, appErr
		}
		slog.Error("payment.repo update_after_create: failed", "id", id, "error", err)
		return nil, err
	}
	return &t, nil
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, mpPaymentID, newStatus string, allowedFrom []string) (int64, error) {
	if len(allowedFrom) == 0 {
		return 0, fmt.Errorf("allowedFrom must not be empty")
	}
	tag, err := r.db.Exec(ctx,
		`UPDATE gift_transactions
		 SET status = $1, updated_at = now()
		 WHERE mp_payment_id = $2 AND status = ANY($3)`,
		newStatus, mpPaymentID, allowedFrom)
	if err != nil {
		slog.Error("payment.repo update_status: failed", "mp_payment_id", mpPaymentID, "error", err)
		return 0, err
	}
	return tag.RowsAffected(), nil
}

func (r *PostgresRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]GiftTransaction, int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+txColumns+`, COUNT(*) OVER() AS total
		   FROM gift_transactions
		  WHERE user_id = $1
		  ORDER BY created_at DESC
		  LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		slog.Error("payment.repo list_by_user_id: query failed", "user_id", userID, "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var txs []GiftTransaction
	var total int
	for rows.Next() {
		var t GiftTransaction
		if err := rows.Scan(
			&t.ID, &t.GiftID, &t.UserID, &t.PaymentMethod, &t.MPPaymentID, &t.MPPreferenceID,
			&t.AmountCents, &t.Status, &t.IdempotencyKey, &t.CreatedAt, &t.UpdatedAt, &t.GiftNameSnapshot,
			&total,
		); err != nil {
			slog.Error("payment.repo list_by_user_id: scan failed", "error", err)
			return nil, 0, err
		}
		txs = append(txs, t)
	}
	if txs == nil {
		txs = []GiftTransaction{}
	}
	return txs, total, rows.Err()
}

func (r *PostgresRepository) ListAll(ctx context.Context, filter ListFilter, limit, offset int) ([]AdminTransactionRow, int, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+gtTxColumns+`, u.uracf, u.phone, COUNT(*) OVER() AS total
		   FROM gift_transactions gt
		   JOIN users u ON u.id = gt.user_id
		  WHERE ($1::text IS NULL OR gt.status = $1)
		    AND ($2::bigint IS NULL OR gt.gift_id = $2)
		  ORDER BY gt.created_at DESC
		  LIMIT $3 OFFSET $4`,
		filter.Status, filter.GiftID, limit, offset)
	if err != nil {
		slog.Error("payment.repo list_all: query failed", "error", err)
		return nil, 0, err
	}
	defer rows.Close()

	var result []AdminTransactionRow
	var total int
	for rows.Next() {
		var row AdminTransactionRow
		t := &row.GiftTransaction
		if err := rows.Scan(
			&t.ID, &t.GiftID, &t.UserID, &t.PaymentMethod, &t.MPPaymentID, &t.MPPreferenceID,
			&t.AmountCents, &t.Status, &t.IdempotencyKey, &t.CreatedAt, &t.UpdatedAt, &t.GiftNameSnapshot,
			&row.UserURACF, &row.UserPhone,
			&total,
		); err != nil {
			slog.Error("payment.repo list_all: scan failed", "error", err)
			return nil, 0, err
		}
		result = append(result, row)
	}
	if result == nil {
		result = []AdminTransactionRow{}
	}
	return result, total, rows.Err()
}

func (r *PostgresRepository) Summary(ctx context.Context) (*AdminSummary, error) {
	var summary AdminSummary

	row := r.db.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount_cents) FILTER (WHERE status = 'approved'), 0)
		   FROM gift_transactions`)
	if err := row.Scan(&summary.Total, &summary.ApprovedTotalCents); err != nil {
		slog.Error("payment.repo summary: totals query failed", "error", err)
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT status, COUNT(*), COALESCE(SUM(amount_cents), 0)
		   FROM gift_transactions
		  GROUP BY status
		  ORDER BY status`)
	if err != nil {
		slog.Error("payment.repo summary: by_status query failed", "error", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var b StatusBreakdown
		if err := rows.Scan(&b.Status, &b.Count, &b.TotalCents); err != nil {
			slog.Error("payment.repo summary: scan failed", "error", err)
			return nil, err
		}
		summary.ByStatus = append(summary.ByStatus, b)
	}
	if summary.ByStatus == nil {
		summary.ByStatus = []StatusBreakdown{}
	}
	return &summary, rows.Err()
}

func mapPgError(err error) *apperror.AppError {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}
	switch pgErr.Code {
	case pgerrcode.UniqueViolation:
		switch pgErr.ConstraintName {
		case "gift_transactions_mp_payment_id_unique":
			return apperror.Conflict("Transação já registrada para este pagamento.")
		case "gift_transactions_idempotency_key_unique":
			return apperror.Conflict("Pedido duplicado: tente novamente.")
		}
		return apperror.Conflict("Transação em conflito com registro existente.")
	case pgerrcode.CheckViolation:
		return apperror.Validation(fmt.Sprintf("Dados de pagamento inválidos (%s).", pgErr.ConstraintName))
	case pgerrcode.ForeignKeyViolation:
		return apperror.Validation("Presente ou usuário referenciado não existe.")
	}
	return nil
}
