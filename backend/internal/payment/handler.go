package payment

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
)

const (
	maxPurchaseBodySize = 64 << 10
	maxWebhookBodySize  = 256 << 10
	webhookTimeout      = 20 * time.Second
)

type Handler struct {
	svc *Service
	mp  PaymentGateway
}

func NewHandler(svc *Service, mp PaymentGateway) *Handler {
	return &Handler{svc: svc, mp: mp}
}

func (h *Handler) HandleCreatePurchase(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("autenticação obrigatória"))
		return
	}

	giftID, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("id de presente inválido", err))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPurchaseBodySize)

	var input CreatePurchaseInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("payload de compra inválido", err))
		return
	}

	resp, err := h.svc.CreatePurchase(r.Context(), giftID, userID, input)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("falha ao criar compra", err))
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleListMyPurchases(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("autenticação obrigatória"))
		return
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	resp, err := h.svc.ListMyPurchases(r.Context(), userID, page, limit)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleGetMyPurchase(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("autenticação obrigatória"))
		return
	}
	txID, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("id de transação inválido", err))
		return
	}
	resp, err := h.svc.GetMyPurchase(r.Context(), userID, txID)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleListAll(w http.ResponseWriter, r *http.Request) {
	var filter ListFilter
	if s := r.URL.Query().Get("status"); s != "" {
		filter.Status = &s
	}
	if g := r.URL.Query().Get("gift_id"); g != "" {
		if v, err := strconv.ParseInt(g, 10, 64); err == nil {
			filter.GiftID = &v
		}
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	resp, err := h.svc.ListAll(r.Context(), filter, page, limit)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.Summary(r.Context())
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

type webhookPayload struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Data   struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (h *Handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodySize)

	dataID := r.URL.Query().Get("data.id")
	if dataID == "" {
		dataID = r.URL.Query().Get("id")
	}

	if h.mp == nil || !h.mp.VerifyWebhookSignature(r.Header, dataID) {
		slog.Warn("payment.webhook: signature verification failed",
			"path", r.URL.Path,
			"remote", r.RemoteAddr,
			"x-request-id", r.Header.Get("x-request-id"),
		)
		httputil.WriteError(w, r, apperror.Unauthorized("invalid signature"))
		return
	}

	var payload webhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		slog.Warn("payment.webhook: failed to decode body", "error", err)
	}

	if payload.Type != "" && payload.Type != "payment" {
		slog.Info("payment.webhook: non-payment event ignored", "type", payload.Type, "action", payload.Action)
		w.WriteHeader(http.StatusOK)
		return
	}

	if dataID == "" && payload.Data.ID != "" {
		dataID = payload.Data.ID
	}
	if dataID == "" {
		slog.Warn("payment.webhook: missing data.id")
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), webhookTimeout)
	defer cancel()

	if err := h.svc.HandleWebhookEvent(ctx, dataID); err != nil {
		slog.Error("payment.webhook: service error", "data_id", dataID, "error", err)
		var ae *apperror.AppError
		if app, ok := apperror.IsAppError(err); ok {
			ae = app
		}
		if ae != nil && ae.Code >= 500 {
			httputil.WriteError(w, r, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}
