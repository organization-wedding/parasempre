package payment

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

type mockRepository struct {
	createFn       func(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error)
	getByIDFn      func(ctx context.Context, id int64) (*GiftTransaction, error)
	getByMPFn      func(ctx context.Context, mpPaymentID string) (*GiftTransaction, error)
	updateAfterFn  func(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error)
	updateStatusFn func(ctx context.Context, mpPaymentID, newStatus string, allowedFrom []string) (int64, error)
}

func (m *mockRepository) Create(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error) {
	return m.createFn(ctx, input)
}
func (m *mockRepository) GetByID(ctx context.Context, id int64) (*GiftTransaction, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockRepository) GetByMPPaymentID(ctx context.Context, mpPaymentID string) (*GiftTransaction, error) {
	return m.getByMPFn(ctx, mpPaymentID)
}
func (m *mockRepository) UpdateAfterCreate(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
	return m.updateAfterFn(ctx, id, mpPaymentID, status)
}
func (m *mockRepository) UpdateStatus(ctx context.Context, mpPaymentID, newStatus string, allowedFrom []string) (int64, error) {
	return m.updateStatusFn(ctx, mpPaymentID, newStatus, allowedFrom)
}
func (m *mockRepository) WithTx(_ pgx.Tx) Repository { return m }
func (m *mockRepository) ListByUserID(_ context.Context, _ int64, _, _ int) ([]GiftTransaction, int, error) {
	return nil, 0, nil
}
func (m *mockRepository) ListAll(_ context.Context, _ ListFilter, _, _ int) ([]AdminTransactionRow, int, error) {
	return nil, 0, nil
}
func (m *mockRepository) Summary(_ context.Context) (*AdminSummary, error) {
	return &AdminSummary{}, nil
}

type mockTxRunner struct{}

func (m *mockTxRunner) RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error { return fn(nil) }

type mockGateway struct {
	createFn func(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error)
	getFn    func(ctx context.Context, mpPaymentID string) (*MPPayment, error)
	verifyFn func(headers http.Header, dataID string) bool
}

func (g *mockGateway) CreatePayment(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error) {
	return g.createFn(ctx, req, idempotencyKey)
}
func (g *mockGateway) GetPayment(ctx context.Context, mpPaymentID string) (*MPPayment, error) {
	return g.getFn(ctx, mpPaymentID)
}
func (g *mockGateway) VerifyWebhookSignature(headers http.Header, dataID string) bool {
	if g.verifyFn != nil {
		return g.verifyFn(headers, dataID)
	}
	return true
}

type mockGiftFinder struct {
	getFn func(ctx context.Context, id int64) (*GiftSnapshot, error)
}

func (m *mockGiftFinder) GetByID(ctx context.Context, id int64) (*GiftSnapshot, error) {
	return m.getFn(ctx, id)
}

type mockAuditLogger struct {
	calls []auditCall
}

type auditCall struct {
	userID  int64
	action  string
	details map[string]any
}

func (m *mockAuditLogger) LogAction(_ context.Context, userID int64, action string, details map[string]any) error {
	m.calls = append(m.calls, auditCall{userID: userID, action: action, details: details})
	return nil
}

func sampleGift() *GiftSnapshot {
	return &GiftSnapshot{ID: 1, Name: "Panela", PriceCents: 19990, Status: "active"}
}

func sampleTx() *GiftTransaction {
	return &GiftTransaction{
		ID: 100, GiftID: 1, UserID: 42,
		PaymentMethod: PaymentMethodCreditCard,
		AmountCents:   19990,
		Status:        StatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func validCardInput() CreatePurchaseInput {
	tok := "card_token_xyz"
	issuer := "310"
	installments := 3
	return CreatePurchaseInput{
		PaymentMethodID: "visa",
		Token:           &tok,
		IssuerID:        &issuer,
		Installments:    &installments,
		Payer: Payer{
			Email:          "buyer@example.com",
			Identification: PayerIdentification{Type: "CPF", Number: "390.533.447-05"},
		},
		IdempotencyKey: "test-idem-key-12345",
	}
}

func validPixInput() CreatePurchaseInput {
	return CreatePurchaseInput{
		PaymentMethodID: "pix",
		Payer: Payer{
			Email:          "buyer@example.com",
			Identification: PayerIdentification{Type: "CPF", Number: "390.533.447-05"},
		},
		IdempotencyKey: "test-idem-key-pix-12",
	}
}

func assertAppError(t *testing.T, err error, wantCode int, wantMsg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantMsg)
	}
	ae, ok := apperror.IsAppError(err)
	if !ok {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if ae.Code != wantCode {
		t.Fatalf("expected code %d, got %d (msg=%q)", wantCode, ae.Code, ae.Message)
	}
	if !strings.Contains(ae.Message, wantMsg) {
		t.Fatalf("expected message containing %q, got %q", wantMsg, ae.Message)
	}
}

func TestCreatePurchase_HappyPathCard(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error) {
			if input.AmountCents != 19990 {
				t.Errorf("expected AmountCents=19990, got %d", input.AmountCents)
			}
			if input.IdempotencyKey == "" {
				t.Error("expected idempotency key set")
			}
			tx := sampleTx()
			tx.IdempotencyKey = &input.IdempotencyKey
			return tx, nil
		},
		updateAfterFn: func(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
			tx := sampleTx()
			tx.MPPaymentID = &mpPaymentID
			tx.Status = status
			return tx, nil
		},
	}
	gw := &mockGateway{
		createFn: func(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error) {
			if idempotencyKey == "" {
				t.Error("expected idempotency key passed to MP client")
			}
			if req.TransactionAmount != 199.90 {
				t.Errorf("expected TransactionAmount=199.90, got %v", req.TransactionAmount)
			}
			if req.Payer.Identification.Number != "39053344705" {
				t.Errorf("expected stripped CPF '39053344705', got %q", req.Payer.Identification.Number)
			}
			return &MPPayment{ID: 555, Status: "approved", TransactionAmount: 199.90}, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{
		getFn: func(ctx context.Context, id int64) (*GiftSnapshot, error) { return sampleGift(), nil },
	}, &mockAuditLogger{})
	resp, err := svc.CreatePurchase(context.Background(), 1, 42, validCardInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != StatusApproved {
		t.Errorf("expected status approved, got %q", resp.Status)
	}
	if resp.MPPaymentID == nil || *resp.MPPaymentID != "555" {
		t.Errorf("expected mp_payment_id '555', got %v", resp.MPPaymentID)
	}
}

func TestCreatePurchase_HappyPathPix(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error) {
			if input.PaymentMethod != PaymentMethodPix {
				t.Errorf("expected pix method, got %q", input.PaymentMethod)
			}
			return sampleTx(), nil
		},
		updateAfterFn: func(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
			tx := sampleTx()
			tx.MPPaymentID = &mpPaymentID
			tx.Status = status
			tx.PaymentMethod = PaymentMethodPix
			return tx, nil
		},
	}
	gw := &mockGateway{
		createFn: func(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error) {
			if req.Token != "" {
				t.Errorf("expected no token for PIX, got %q", req.Token)
			}
			return &MPPayment{
				ID:                777,
				Status:            "pending",
				TransactionAmount: 199.90,
				PointOfInteraction: &MPPointOfInteraction{
					TransactionData: &MPTransactionData{
						QRCode:       "00020126...",
						QRCodeBase64: "iVBORw0KGgoA...",
					},
				},
			}, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{
		getFn: func(ctx context.Context, id int64) (*GiftSnapshot, error) { return sampleGift(), nil },
	}, &mockAuditLogger{})
	resp, err := svc.CreatePurchase(context.Background(), 1, 42, validPixInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Pix == nil {
		t.Fatal("expected pix data in response")
	}
	if resp.Pix.QRCodeBase64 == "" {
		t.Error("expected qr_code_base64 to be propagated")
	}
}

func TestCreatePurchase_RejectsCreditCardWithoutToken(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockTxRunner{}, &mockGateway{}, &mockGiftFinder{
		getFn: func(ctx context.Context, id int64) (*GiftSnapshot, error) { return sampleGift(), nil },
	}, &mockAuditLogger{})
	in := validCardInput()
	in.Token = nil
	_, err := svc.CreatePurchase(context.Background(), 1, 42, in)
	assertAppError(t, err, http.StatusBadRequest, "token é obrigatório")
}

func TestCreatePurchase_RejectsInactiveGift(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockTxRunner{}, &mockGateway{}, &mockGiftFinder{
		getFn: func(ctx context.Context, id int64) (*GiftSnapshot, error) {
			g := sampleGift()
			g.Status = "inactive"
			return g, nil
		},
	}, &mockAuditLogger{})
	_, err := svc.CreatePurchase(context.Background(), 1, 42, validCardInput())
	assertAppError(t, err, http.StatusNotFound, "presente não encontrado")
}

func TestCreatePurchase_RejectsUnauthenticated(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockTxRunner{}, &mockGateway{}, &mockGiftFinder{}, &mockAuditLogger{})
	_, err := svc.CreatePurchase(context.Background(), 1, 0, validCardInput())
	assertAppError(t, err, http.StatusUnauthorized, "autenticação obrigatória")
}

func TestCreatePurchase_PersistsPendingOnTimeout(t *testing.T) {
	var persistedStatus string
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error) {
			return sampleTx(), nil
		},
		updateAfterFn: func(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
			persistedStatus = status
			tx := sampleTx()
			tx.Status = status
			return tx, nil
		},
	}
	gw := &mockGateway{
		createFn: func(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error) {
			return nil, apperror.ServiceUnavailable("MP timeout")
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{
		getFn: func(ctx context.Context, id int64) (*GiftSnapshot, error) { return sampleGift(), nil },
	}, &mockAuditLogger{})
	_, err := svc.CreatePurchase(context.Background(), 1, 42, validCardInput())
	assertAppError(t, err, http.StatusServiceUnavailable, "MP timeout")
	if persistedStatus != StatusPending {
		t.Errorf("expected persisted status pending after timeout, got %q", persistedStatus)
	}
}

func TestCreatePurchase_PersistsRejectedOnValidationError(t *testing.T) {
	var persistedStatus string
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error) {
			return sampleTx(), nil
		},
		updateAfterFn: func(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
			persistedStatus = status
			tx := sampleTx()
			tx.Status = status
			return tx, nil
		},
	}
	gw := &mockGateway{
		createFn: func(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error) {
			return nil, apperror.Validation("Card declined")
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{
		getFn: func(ctx context.Context, id int64) (*GiftSnapshot, error) { return sampleGift(), nil },
	}, &mockAuditLogger{})
	_, err := svc.CreatePurchase(context.Background(), 1, 42, validCardInput())
	assertAppError(t, err, http.StatusBadRequest, "Card declined")
	if persistedStatus != StatusRejected {
		t.Errorf("expected persisted status rejected after validation error, got %q", persistedStatus)
	}
}

func TestHandleWebhookEvent_TransitionsPendingToApproved(t *testing.T) {
	var capturedFrom []string
	repo := &mockRepository{
		getByMPFn: func(ctx context.Context, mp string) (*GiftTransaction, error) {
			tx := sampleTx()
			mpID := mp
			tx.MPPaymentID = &mpID
			return tx, nil
		},
		updateStatusFn: func(ctx context.Context, mpPaymentID, newStatus string, allowedFrom []string) (int64, error) {
			capturedFrom = allowedFrom
			if newStatus != StatusApproved {
				t.Errorf("expected newStatus=approved, got %q", newStatus)
			}
			return 1, nil
		},
	}
	gw := &mockGateway{
		getFn: func(ctx context.Context, id string) (*MPPayment, error) {
			return &MPPayment{ID: 999, Status: "approved", TransactionAmount: 199.90}, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{}, &mockAuditLogger{})
	if err := svc.HandleWebhookEvent(context.Background(), "999"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capturedFrom) != 1 || capturedFrom[0] != StatusPending {
		t.Errorf("expected allowedFrom=[pending], got %v", capturedFrom)
	}
}

func TestHandleWebhookEvent_RejectsAmountMismatch(t *testing.T) {
	repo := &mockRepository{
		getByMPFn: func(ctx context.Context, mp string) (*GiftTransaction, error) {
			tx := sampleTx()
			tx.AmountCents = 19990 // R$ 199.90
			return tx, nil
		},
	}
	gw := &mockGateway{
		getFn: func(ctx context.Context, id string) (*MPPayment, error) {
			return &MPPayment{ID: 999, Status: "approved", TransactionAmount: 1.00}, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{}, &mockAuditLogger{})
	err := svc.HandleWebhookEvent(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error for amount mismatch, got nil")
	}
	ae, ok := apperror.IsAppError(err)
	if !ok || ae.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 internal error, got %v", err)
	}
}

func TestHandleWebhookEvent_TolerantToUnknownTransaction(t *testing.T) {
	repo := &mockRepository{
		getByMPFn: func(ctx context.Context, mp string) (*GiftTransaction, error) {
			return nil, apperror.NotFound("transaction not found")
		},
	}
	gw := &mockGateway{
		getFn: func(ctx context.Context, id string) (*MPPayment, error) {
			// external_reference points elsewhere — payment not ours.
			return &MPPayment{ID: 999, Status: "approved", TransactionAmount: 199.90, ExternalReference: ""}, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{}, &mockAuditLogger{})
	if err := svc.HandleWebhookEvent(context.Background(), "999"); err != nil {
		t.Fatalf("expected nil for unknown tx (idempotency), got %v", err)
	}
}

func TestHandleWebhookEvent_RecoversOrphanViaExternalReference(t *testing.T) {
	var linkedMPID string
	var transitioned bool
	repo := &mockRepository{
		getByMPFn: func(ctx context.Context, mp string) (*GiftTransaction, error) {
			return nil, apperror.NotFound("transaction not found")
		},
		getByIDFn: func(ctx context.Context, id int64) (*GiftTransaction, error) {
			if id != 42 {
				t.Errorf("expected GetByID(42), got %d", id)
			}
			tx := sampleTx()
			tx.ID = 42
			return tx, nil
		},
		updateAfterFn: func(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
			linkedMPID = mpPaymentID
			tx := sampleTx()
			tx.ID = 42
			tx.MPPaymentID = &mpPaymentID
			tx.Status = status
			return tx, nil
		},
		updateStatusFn: func(ctx context.Context, mpPaymentID, newStatus string, allowedFrom []string) (int64, error) {
			transitioned = true
			return 1, nil
		},
	}
	gw := &mockGateway{
		getFn: func(ctx context.Context, id string) (*MPPayment, error) {
			return &MPPayment{ID: 999, Status: "approved", TransactionAmount: 199.90, ExternalReference: "gift_tx:42"}, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{}, &mockAuditLogger{})
	if err := svc.HandleWebhookEvent(context.Background(), "999"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if linkedMPID != "999" {
		t.Errorf("expected mp_payment_id linked to '999', got %q", linkedMPID)
	}
	if !transitioned {
		t.Error("expected status transition after recovery")
	}
}

func TestParseExternalReference(t *testing.T) {
	cases := map[string]struct {
		id int64
		ok bool
	}{
		"gift_tx:42":    {42, true},
		"gift_tx:0":     {0, false},
		"gift_tx:-5":    {0, false},
		"gift_tx:abc":   {0, false},
		"order:42":      {0, false},
		"":              {0, false},
		"gift_tx:":      {0, false},
		"gift_tx:99999": {99999, true},
	}
	for in, want := range cases {
		gotID, gotOk := parseExternalReference(in)
		if gotID != want.id || gotOk != want.ok {
			t.Errorf("parseExternalReference(%q) = (%d, %v), want (%d, %v)", in, gotID, gotOk, want.id, want.ok)
		}
	}
}

func TestHandleWebhookEvent_RowsAffectedZeroIsIdempotent(t *testing.T) {
	repo := &mockRepository{
		getByMPFn: func(ctx context.Context, mp string) (*GiftTransaction, error) {
			tx := sampleTx()
			tx.Status = StatusApproved // already approved — replay
			return tx, nil
		},
		updateStatusFn: func(ctx context.Context, mpPaymentID, newStatus string, allowedFrom []string) (int64, error) {
			return 0, nil
		},
	}
	gw := &mockGateway{
		getFn: func(ctx context.Context, id string) (*MPPayment, error) {
			return &MPPayment{ID: 999, Status: "approved", TransactionAmount: 199.90}, nil
		},
	}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{}, &mockAuditLogger{})
	if err := svc.HandleWebhookEvent(context.Background(), "999"); err != nil {
		t.Fatalf("expected nil on replay, got %v", err)
	}
}

func TestMapMPStatus(t *testing.T) {
	cases := map[string]string{
		"approved":      StatusApproved,
		"rejected":      StatusRejected,
		"cancelled":     StatusCancelled,
		"refunded":      StatusRefunded,
		"charged_back":  StatusRefunded,
		"pending":       StatusPending,
		"in_process":    StatusPending,
		"authorized":    StatusPending,
		"weird_unknown": "",
	}
	for in, want := range cases {
		if got := mapMPStatus(in); got != want {
			t.Errorf("mapMPStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAllowedFromStatuses_TerminalStatesAreSticky(t *testing.T) {
	if got := allowedFromStatuses(StatusApproved); !equalStrings(got, []string{StatusPending}) {
		t.Errorf("approved should accept only from pending, got %v", got)
	}
	if got := allowedFromStatuses(StatusRefunded); !equalStrings(got, []string{StatusApproved}) {
		t.Errorf("refunded should accept only from approved, got %v", got)
	}
	if got := allowedFromStatuses(StatusPending); got != nil {
		t.Errorf("pending should never be a target, got %v", got)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestCreatePurchase_ServiceUnavailableWhenMPDisabled(t *testing.T) {
	svc := NewService(&mockRepository{}, &mockTxRunner{}, nil, &mockGiftFinder{}, &mockAuditLogger{})
	_, err := svc.CreatePurchase(context.Background(), 1, 42, validCardInput())
	assertAppError(t, err, http.StatusServiceUnavailable, "Pagamentos indisponíveis")
}

func TestResolveMPOutcome_PicksCorrectStatusFromError(t *testing.T) {
	cases := []struct {
		name       string
		mpResp     *MPPayment
		mpErr      error
		wantStatus string
	}{
		{
			name:       "success approved",
			mpResp:     &MPPayment{ID: 1, Status: "approved"},
			wantStatus: StatusApproved,
		},
		{
			name:       "validation error -> rejected",
			mpErr:      apperror.Validation("declined"),
			wantStatus: StatusRejected,
		},
		{
			name:       "service unavailable -> pending",
			mpErr:      apperror.ServiceUnavailable("timeout"),
			wantStatus: StatusPending,
		},
		{
			name:       "generic error -> pending",
			mpErr:      errors.New("unexpected"),
			wantStatus: StatusPending,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, _, _, _ := resolveMPOutcome(c.mpResp, c.mpErr)
			if s != c.wantStatus {
				t.Errorf("got status %q, want %q", s, c.wantStatus)
			}
		})
	}
}

func TestCreatePurchase_RecordsAuditOnSuccess(t *testing.T) {
	repo := &mockRepository{
		createFn: func(ctx context.Context, input CreateGiftTransactionInput) (*GiftTransaction, error) {
			return sampleTx(), nil
		},
		updateAfterFn: func(ctx context.Context, id int64, mpPaymentID, status string) (*GiftTransaction, error) {
			tx := sampleTx()
			tx.MPPaymentID = &mpPaymentID
			tx.Status = status
			return tx, nil
		},
	}
	gw := &mockGateway{
		createFn: func(ctx context.Context, req CreatePaymentRequest, idempotencyKey string) (*MPPayment, error) {
			return &MPPayment{ID: 555, Status: "approved", TransactionAmount: 199.90}, nil
		},
	}
	audit := &mockAuditLogger{}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{
		getFn: func(ctx context.Context, id int64) (*GiftSnapshot, error) { return sampleGift(), nil },
	}, audit)
	if _, err := svc.CreatePurchase(context.Background(), 1, 42, validCardInput()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(audit.calls) != 1 {
		t.Fatalf("expected 1 audit call, got %d", len(audit.calls))
	}
	c := audit.calls[0]
	if c.action != auditPurchaseCreated {
		t.Errorf("expected action %q, got %q", auditPurchaseCreated, c.action)
	}
	if c.userID != 42 {
		t.Errorf("expected userID 42, got %d", c.userID)
	}
	if c.details["mp_payment_id"] != "555" {
		t.Errorf("expected mp_payment_id '555', got %v", c.details["mp_payment_id"])
	}
}

func TestHandleWebhookEvent_RecordsAuditOnAmountMismatch(t *testing.T) {
	repo := &mockRepository{
		getByMPFn: func(ctx context.Context, mp string) (*GiftTransaction, error) {
			return sampleTx(), nil
		},
	}
	gw := &mockGateway{
		getFn: func(ctx context.Context, id string) (*MPPayment, error) {
			return &MPPayment{ID: 999, Status: "approved", TransactionAmount: 1.00}, nil
		},
	}
	audit := &mockAuditLogger{}
	svc := NewService(repo, &mockTxRunner{}, gw, &mockGiftFinder{}, audit)
	if err := svc.HandleWebhookEvent(context.Background(), "999"); err == nil {
		t.Fatal("expected amount mismatch error")
	}
	if len(audit.calls) != 1 || audit.calls[0].action != auditWebhookAmountMismatch {
		t.Errorf("expected single %q audit, got %v", auditWebhookAmountMismatch, audit.calls)
	}
}
