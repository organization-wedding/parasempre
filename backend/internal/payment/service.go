package payment

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/database"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

const externalRefPrefix = "gift_tx:"

type GiftFinder interface {
	GetByID(ctx context.Context, id int64) (*GiftSnapshot, error)
}

type GiftSnapshot struct {
	ID         int64
	Name       string
	PriceCents int64
	Status     string
}

// AuditLogger records user-attributed payment events into audit_log. Errors
// are surfaced to the caller; the service uses fire-and-forget (logs on fail)
// so audit issues never break the user flow.
type AuditLogger interface {
	LogAction(ctx context.Context, userID int64, action string, details map[string]any) error
}

type Service struct {
	repo     TxAwareRepository
	txRunner database.TxRunner
	mp       PaymentGateway
	gifts    GiftFinder
	audit    AuditLogger
}

func NewService(repo TxAwareRepository, txRunner database.TxRunner, mp PaymentGateway, gifts GiftFinder, audit AuditLogger) *Service {
	return &Service{repo: repo, txRunner: txRunner, mp: mp, gifts: gifts, audit: audit}
}

const (
	auditPurchaseCreated       = "payment.purchase_created"
	auditPurchaseFailed        = "payment.purchase_failed"
	auditWebhookStatusChanged  = "payment.webhook_status_changed"
	auditWebhookAmountMismatch = "payment.webhook_amount_mismatch"
	auditOrphanRecovered       = "payment.orphan_recovered"
)

func (s *Service) recordAudit(ctx context.Context, userID int64, action string, details map[string]any) {
	if s.audit == nil || userID == 0 {
		return
	}
	if err := s.audit.LogAction(ctx, userID, action, details); err != nil {
		slog.Error("payment.service audit failed", "action", action, "user_id", userID, "error", err)
	}
}

func (s *Service) CreatePurchase(ctx context.Context, giftID, userID int64, input CreatePurchaseInput) (*PurchaseResponse, error) {
	if s.mp == nil {
		return nil, apperror.ServiceUnavailable("Pagamentos indisponíveis neste ambiente.")
	}
	if userID == 0 {
		return nil, apperror.Unauthorized("autenticação obrigatória")
	}

	if err := validate.Struct(input); err != nil {
		return nil, err
	}

	method := paymentMethodFromMPID(input.PaymentMethodID)
	if method == PaymentMethodCreditCard && (input.Token == nil || *input.Token == "") {
		return nil, apperror.Validation("token é obrigatório para pagamento com cartão")
	}

	g, err := s.gifts.GetByID(ctx, giftID)
	if err != nil {
		return nil, apperror.WrapIfNotApp("falha ao carregar presente", err)
	}
	if g == nil || g.Status != "active" {
		return nil, apperror.NotFound("presente não encontrado")
	}

	idempotencyKey := input.IdempotencyKey

	var txRow *GiftTransaction
	if err := s.txRunnerCreate(ctx, &txRow, g, userID, method, idempotencyKey); err != nil {
		return nil, err
	}

	mpReq := buildMPRequest(g, txRow, input)
	mpResp, mpErr := s.mp.CreatePayment(ctx, mpReq, idempotencyKey)

	finalStatus, statusDetail, mpPaymentID, mpErrToReturn := resolveMPOutcome(mpResp, mpErr)

	updated, updErr := s.repo.UpdateAfterCreate(ctx, txRow.ID, mpPaymentID, finalStatus)
	if updErr != nil {
		slog.Error("payment.service create_purchase: persist update failed",
			"tx_id", txRow.ID,
			"mp_payment_id", mpPaymentID,
			"final_status", finalStatus,
			"error", updErr,
		)
		if mpErrToReturn != nil {
			return nil, mpErrToReturn
		}
		return nil, apperror.WrapIfNotApp("falha ao persistir resultado do pagamento", updErr)
	}

	if mpErrToReturn != nil {
		s.recordAudit(ctx, userID, auditPurchaseFailed, map[string]any{
			"tx_id":        txRow.ID,
			"gift_id":      g.ID,
			"method":       method,
			"amount_cents": g.PriceCents,
			"final_status": finalStatus,
			"mp_error":     mpErrToReturn.Error(),
		})
		return nil, mpErrToReturn
	}

	s.recordAudit(ctx, userID, auditPurchaseCreated, map[string]any{
		"tx_id":         updated.ID,
		"gift_id":       g.ID,
		"method":        method,
		"amount_cents":  g.PriceCents,
		"status":        finalStatus,
		"mp_payment_id": mpPaymentID,
	})

	resp := &PurchaseResponse{
		TransactionID: updated.ID,
		MPPaymentID:   updated.MPPaymentID,
		Status:        updated.Status,
		StatusDetail:  statusDetail,
		PaymentMethod: updated.PaymentMethod,
		AmountCents:   updated.AmountCents,
	}

	if mpResp != nil && mpResp.PointOfInteraction != nil && mpResp.PointOfInteraction.TransactionData != nil {
		td := mpResp.PointOfInteraction.TransactionData
		if td.QRCode != "" || td.QRCodeBase64 != "" || td.TicketURL != "" {
			resp.Pix = &PixData{
				QRCode:       td.QRCode,
				QRCodeBase64: td.QRCodeBase64,
				TicketURL:    td.TicketURL,
			}
		}
	}

	slog.Info("payment.service create_purchase: done",
		"tx_id", updated.ID,
		"gift_id", g.ID,
		"user_id", userID,
		"method", method,
		"mp_payment_id", mpPaymentID,
		"status", finalStatus,
	)

	return resp, nil
}

func (s *Service) txRunnerCreate(ctx context.Context, out **GiftTransaction, g *GiftSnapshot, userID int64, method, idempotencyKey string) error {
	err := s.txRunner.RunInTx(ctx, func(tx pgx.Tx) error {
		row, err := s.repo.WithTx(tx).Create(ctx, CreateGiftTransactionInput{
			GiftID:           g.ID,
			UserID:           userID,
			PaymentMethod:    method,
			AmountCents:      g.PriceCents,
			Status:           StatusPending,
			IdempotencyKey:   idempotencyKey,
			GiftNameSnapshot: g.Name,
		})
		if err != nil {
			return err
		}
		*out = row
		return nil
	})
	if err != nil {
		return apperror.WrapIfNotApp("falha ao registrar transação", err)
	}
	return nil
}

func (s *Service) HandleWebhookEvent(ctx context.Context, dataID string) error {
	if s.mp == nil {
		return apperror.ServiceUnavailable("Pagamentos indisponíveis neste ambiente.")
	}
	if dataID == "" {
		return apperror.Validation("missing data.id")
	}

	mpPayment, err := s.mp.GetPayment(ctx, dataID)
	if err != nil {
		return apperror.WrapIfNotApp("failed to fetch authoritative MP state", err)
	}

	row, err := s.repo.GetByMPPaymentID(ctx, dataID)
	if err != nil {
		var ae *apperror.AppError
		if !errors.As(err, &ae) || ae.Code != http.StatusNotFound {
			return apperror.WrapIfNotApp("failed to load transaction", err)
		}
		recovered, recErr := s.recoverByExternalReference(ctx, dataID, mpPayment.ExternalReference)
		if recErr != nil {
			return recErr
		}
		if recovered == nil {
			slog.Warn("payment.service webhook: transaction not recoverable",
				"mp_payment_id", dataID,
				"external_reference", mpPayment.ExternalReference,
			)
			return nil
		}
		row = recovered
	}

	if CentsFromAmount(mpPayment.TransactionAmount) != row.AmountCents {
		slog.Error("payment.service webhook: amount mismatch",
			"mp_payment_id", dataID,
			"expected_cents", row.AmountCents,
			"got_amount", mpPayment.TransactionAmount,
		)
		s.recordAudit(ctx, row.UserID, auditWebhookAmountMismatch, map[string]any{
			"tx_id":          row.ID,
			"mp_payment_id":  dataID,
			"expected_cents": row.AmountCents,
			"got_amount":     mpPayment.TransactionAmount,
		})
		return apperror.Internal("payment amount mismatch", nil)
	}

	newStatus := mapMPStatus(mpPayment.Status)
	allowedFrom := allowedFromStatuses(newStatus)
	if len(allowedFrom) == 0 {
		slog.Warn("payment.service webhook: unmappable status", "mp_status", mpPayment.Status)
		return nil
	}

	rows, err := s.repo.UpdateStatus(ctx, dataID, newStatus, allowedFrom)
	if err != nil {
		return apperror.WrapIfNotApp("failed to update transaction status", err)
	}
	if rows == 0 {
		slog.Info("payment.service webhook: status transition rejected (replay or terminal)",
			"mp_payment_id", dataID,
			"current_db_status", row.Status,
			"target_status", newStatus,
		)
	} else {
		slog.Info("payment.service webhook: status updated",
			"mp_payment_id", dataID,
			"from", row.Status,
			"to", newStatus,
		)
		s.recordAudit(ctx, row.UserID, auditWebhookStatusChanged, map[string]any{
			"tx_id":         row.ID,
			"mp_payment_id": dataID,
			"from":          row.Status,
			"to":            newStatus,
		})
	}
	return nil
}

func (s *Service) recoverByExternalReference(ctx context.Context, mpPaymentID, externalRef string) (*GiftTransaction, error) {
	txID, ok := parseExternalReference(externalRef)
	if !ok {
		return nil, nil
	}
	row, err := s.repo.GetByID(ctx, txID)
	if err != nil {
		var ae *apperror.AppError
		if errors.As(err, &ae) && ae.Code == http.StatusNotFound {
			return nil, nil
		}
		return nil, apperror.WrapIfNotApp("failed to load tx by external_reference", err)
	}
	if row.MPPaymentID != nil && *row.MPPaymentID == mpPaymentID {
		return row, nil
	}
	if row.MPPaymentID != nil && *row.MPPaymentID != mpPaymentID {
		slog.Warn("payment.service webhook: external_reference points to tx with different mp_payment_id",
			"tx_id", txID,
			"expected", mpPaymentID,
			"have", *row.MPPaymentID,
		)
		return nil, nil
	}
	updated, err := s.repo.UpdateAfterCreate(ctx, row.ID, mpPaymentID, row.Status)
	if err != nil {
		return nil, apperror.WrapIfNotApp("failed to link mp_payment_id", err)
	}
	slog.Info("payment.service webhook: recovered orphan tx via external_reference",
		"tx_id", row.ID,
		"mp_payment_id", mpPaymentID,
	)
	s.recordAudit(ctx, row.UserID, auditOrphanRecovered, map[string]any{
		"tx_id":         row.ID,
		"mp_payment_id": mpPaymentID,
	})
	return updated, nil
}

func parseExternalReference(s string) (int64, bool) {
	if !strings.HasPrefix(s, externalRefPrefix) {
		return 0, false
	}
	id, err := strconv.ParseInt(s[len(externalRefPrefix):], 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

func resolveMPOutcome(mpResp *MPPayment, mpErr error) (status, statusDetail, mpPaymentID string, errToReturn error) {
	if mpErr == nil && mpResp != nil {
		return mapMPStatus(mpResp.Status), mpResp.StatusDetail, strconv.FormatInt(mpResp.ID, 10), nil
	}
	var ae *apperror.AppError
	if errors.As(mpErr, &ae) {
		switch ae.Code {
		case http.StatusBadRequest, http.StatusUnprocessableEntity:
			return StatusRejected, ae.Message, "", ae
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return StatusPending, "", "", ae
		}
	}
	return StatusPending, "", "", mpErr
}

func mapMPStatus(mpStatus string) string {
	switch mpStatus {
	case "approved":
		return StatusApproved
	case "rejected":
		return StatusRejected
	case "cancelled":
		return StatusCancelled
	case "refunded", "charged_back":
		return StatusRefunded
	case "pending", "in_process", "authorized":
		return StatusPending
	default:
		return ""
	}
}

func allowedFromStatuses(newStatus string) []string {
	switch newStatus {
	case StatusApproved, StatusRejected, StatusCancelled:
		return []string{StatusPending}
	case StatusRefunded:
		return []string{StatusApproved}
	}
	return nil
}

func paymentMethodFromMPID(mpID string) string {
	if mpID == MPMethodPix {
		return PaymentMethodPix
	}
	return PaymentMethodCreditCard
}

func buildMPRequest(g *GiftSnapshot, row *GiftTransaction, input CreatePurchaseInput) CreatePaymentRequest {
	req := CreatePaymentRequest{
		TransactionAmount: AmountFromCents(g.PriceCents),
		PaymentMethodID:   input.PaymentMethodID,
		Description:       fmt.Sprintf("Presente: %s", g.Name),
		ExternalReference: fmt.Sprintf("%s%d", externalRefPrefix, row.ID),
		Payer: MPPayer{
			Email: input.Payer.Email,
			Identification: MPIdentification{
				Type:   "CPF",
				Number: validate.StripCPF(input.Payer.Identification.Number),
			},
		},
	}
	if input.Token != nil {
		req.Token = *input.Token
	}
	if input.IssuerID != nil {
		req.IssuerID = *input.IssuerID
	}
	if input.Installments != nil && *input.Installments > 0 {
		req.Installments = *input.Installments
	} else if input.PaymentMethodID != MPMethodPix {
		req.Installments = 1
	}
	return req
}

func (s *Service) ListMyPurchases(ctx context.Context, userID int64, page, limit int) (*PagedTransactions[PublicTransaction], error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	offset := (page - 1) * limit

	rows, total, err := s.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, apperror.WrapIfNotApp("falha ao listar compras", err)
	}

	data := make([]PublicTransaction, len(rows))
	for i, row := range rows {
		data[i] = row.ToPublic()
	}
	return &PagedTransactions[PublicTransaction]{Data: data, Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) GetMyPurchase(ctx context.Context, userID, txID int64) (*PublicTransaction, error) {
	row, err := s.repo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	if row.UserID != userID {
		return nil, apperror.NotFound("transação não encontrada")
	}

	pub := row.ToPublic()

	if row.Status == StatusPending && row.PaymentMethod == PaymentMethodPix && row.MPPaymentID != nil {
		mpPayment, mpErr := s.mp.GetPayment(ctx, *row.MPPaymentID)
		if mpErr == nil && mpPayment.PointOfInteraction != nil && mpPayment.PointOfInteraction.TransactionData != nil {
			td := mpPayment.PointOfInteraction.TransactionData
			if td.QRCode != "" || td.QRCodeBase64 != "" {
				pub.Pix = &PixData{
					QRCode:       td.QRCode,
					QRCodeBase64: td.QRCodeBase64,
					TicketURL:    td.TicketURL,
				}
			}
		}
	}

	return &pub, nil
}

func (s *Service) ListAll(ctx context.Context, filter ListFilter, page, limit int) (*PagedTransactions[AdminTransaction], error) {
	if filter.Status != nil && !knownStatuses[*filter.Status] {
		return nil, apperror.Validation("status inválido")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	rows, total, err := s.repo.ListAll(ctx, filter, limit, offset)
	if err != nil {
		return nil, apperror.WrapIfNotApp("falha ao listar transações", err)
	}

	data := make([]AdminTransaction, len(rows))
	for i, row := range rows {
		pub := row.GiftTransaction.ToPublic()
		data[i] = AdminTransaction{
			PublicTransaction: pub,
			UserID:            row.GiftTransaction.UserID,
			UserURACF:         row.UserURACF,
			UserPhone:         row.UserPhone,
			MPPaymentID:       row.GiftTransaction.MPPaymentID,
		}
	}
	return &PagedTransactions[AdminTransaction]{Data: data, Page: page, Limit: limit, Total: total}, nil
}

func (s *Service) Summary(ctx context.Context) (*AdminSummary, error) {
	summary, err := s.repo.Summary(ctx)
	if err != nil {
		return nil, apperror.WrapIfNotApp("falha ao carregar resumo", err)
	}
	return summary, nil
}
