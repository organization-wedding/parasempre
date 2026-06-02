//go:build integration
// +build integration

package payment

import (
	"context"
	"testing"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
)

func setup(t *testing.T) (*PostgresRepository, context.Context, int64, int64) {
	t.Helper()
	pool := database.NewTestPool(t)
	database.CleanTable(t, pool, "gift_transactions")
	database.CleanTable(t, pool, "gifts")
	database.CleanTable(t, pool, "users")
	database.CleanTable(t, pool, "guests")
	ctx := context.Background()

	var giftID int64
	err := pool.QueryRow(ctx,
		`INSERT INTO gifts (name, price_cents, dedupe_key, created_by, updated_by)
		 VALUES ('Test Gift', 19990, 'test gift', 'TST01', 'TST01')
		 RETURNING id`,
	).Scan(&giftID)
	if err != nil {
		t.Fatalf("seed gift failed: %v", err)
	}

	var userID int64
	err = pool.QueryRow(ctx,
		`INSERT INTO users (uracf, role) VALUES ('TST01', 'guest') RETURNING id`,
	).Scan(&userID)
	if err != nil {
		t.Fatalf("seed user failed: %v", err)
	}

	return NewPostgresRepository(pool), ctx, giftID, userID
}

func TestIntegrationCreateAndGetTransaction(t *testing.T) {
	repo, ctx, giftID, userID := setup(t)

	tx, err := repo.Create(ctx, CreateGiftTransactionInput{
		GiftID:         giftID,
		UserID:         userID,
		PaymentMethod:  PaymentMethodCreditCard,
		AmountCents:    19990,
		Status:         StatusPending,
		IdempotencyKey: "idem-key-1",
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if tx.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if tx.Status != StatusPending {
		t.Fatalf("expected status pending, got %q", tx.Status)
	}

	fetched, err := repo.GetByID(ctx, tx.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if fetched.IdempotencyKey == nil || *fetched.IdempotencyKey != "idem-key-1" {
		t.Errorf("idempotency_key not persisted, got %v", fetched.IdempotencyKey)
	}
}

func TestIntegrationDuplicateIdempotencyKeyConflicts(t *testing.T) {
	repo, ctx, giftID, userID := setup(t)

	in := CreateGiftTransactionInput{
		GiftID:         giftID,
		UserID:         userID,
		PaymentMethod:  PaymentMethodCreditCard,
		AmountCents:    19990,
		Status:         StatusPending,
		IdempotencyKey: "duplicate-key",
	}
	if _, err := repo.Create(ctx, in); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}
	_, err := repo.Create(ctx, in)
	if err == nil {
		t.Fatal("expected conflict on duplicate idempotency key")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok || ae.Code != 409 {
		t.Fatalf("expected 409 AppError, got %v", err)
	}
}

func TestIntegrationDuplicateMPPaymentIDConflicts(t *testing.T) {
	repo, ctx, giftID, userID := setup(t)

	tx1, err := repo.Create(ctx, CreateGiftTransactionInput{
		GiftID: giftID, UserID: userID, PaymentMethod: PaymentMethodCreditCard,
		AmountCents: 19990, Status: StatusPending, IdempotencyKey: "k1",
	})
	if err != nil {
		t.Fatal(err)
	}
	tx2, err := repo.Create(ctx, CreateGiftTransactionInput{
		GiftID: giftID, UserID: userID, PaymentMethod: PaymentMethodCreditCard,
		AmountCents: 19990, Status: StatusPending, IdempotencyKey: "k2",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repo.UpdateAfterCreate(ctx, tx1.ID, "mp-555", StatusApproved); err != nil {
		t.Fatalf("first UpdateAfterCreate failed: %v", err)
	}
	_, err = repo.UpdateAfterCreate(ctx, tx2.ID, "mp-555", StatusApproved)
	if err == nil {
		t.Fatal("expected conflict on duplicate mp_payment_id")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok || ae.Code != 409 {
		t.Fatalf("expected 409, got %v", err)
	}
}

func TestIntegrationUpdateStatusGuardsTransitions(t *testing.T) {
	repo, ctx, giftID, userID := setup(t)

	tx, err := repo.Create(ctx, CreateGiftTransactionInput{
		GiftID: giftID, UserID: userID, PaymentMethod: PaymentMethodCreditCard,
		AmountCents: 19990, Status: StatusPending, IdempotencyKey: "k-tx",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.UpdateAfterCreate(ctx, tx.ID, "mp-100", StatusApproved); err != nil {
		t.Fatalf("UpdateAfterCreate failed: %v", err)
	}

	rows, err := repo.UpdateStatus(ctx, "mp-100", StatusApproved, []string{StatusPending})
	if err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}
	if rows != 0 {
		t.Errorf("expected 0 rows affected for invalid transition, got %d", rows)
	}

	rows, err = repo.UpdateStatus(ctx, "mp-100", StatusRefunded, []string{StatusApproved})
	if err != nil {
		t.Fatalf("UpdateStatus refunded failed: %v", err)
	}
	if rows != 1 {
		t.Errorf("expected 1 row affected for approved->refunded, got %d", rows)
	}

	fetched, err := repo.GetByMPPaymentID(ctx, "mp-100")
	if err != nil {
		t.Fatalf("GetByMPPaymentID failed: %v", err)
	}
	if fetched.Status != StatusRefunded {
		t.Errorf("expected status refunded, got %q", fetched.Status)
	}
}

func TestIntegrationGetByMPPaymentID_NotFound(t *testing.T) {
	repo, ctx, _, _ := setup(t)
	_, err := repo.GetByMPPaymentID(ctx, "does-not-exist")
	if err == nil {
		t.Fatal("expected NotFound, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok || ae.Code != 404 {
		t.Fatalf("expected 404, got %v", err)
	}
}

func TestIntegrationCreateRejectsInvalidStatus(t *testing.T) {
	repo, ctx, giftID, userID := setup(t)
	_, err := repo.Create(ctx, CreateGiftTransactionInput{
		GiftID: giftID, UserID: userID, PaymentMethod: PaymentMethodCreditCard,
		AmountCents: 19990, Status: "bogus", IdempotencyKey: "k-bogus",
	})
	if err == nil {
		t.Fatal("expected check violation")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok || ae.Code != 400 {
		t.Fatalf("expected 400 validation, got %v", err)
	}
}

func TestIntegrationCreateRejectsInvalidPaymentMethod(t *testing.T) {
	repo, ctx, giftID, userID := setup(t)
	_, err := repo.Create(ctx, CreateGiftTransactionInput{
		GiftID: giftID, UserID: userID, PaymentMethod: "boleto",
		AmountCents: 19990, Status: StatusPending, IdempotencyKey: "k-bol",
	})
	if err == nil {
		t.Fatal("expected check violation for unknown payment_method")
	}
}

func TestIntegrationCreateRejectsZeroAmount(t *testing.T) {
	repo, ctx, giftID, userID := setup(t)
	_, err := repo.Create(ctx, CreateGiftTransactionInput{
		GiftID: giftID, UserID: userID, PaymentMethod: PaymentMethodCreditCard,
		AmountCents: 0, Status: StatusPending, IdempotencyKey: "k-zero",
	})
	if err == nil {
		t.Fatal("expected check violation for zero amount")
	}
}
