package gift

import (
	"net/http"
	"strconv"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
)

const (
	maxBodySize  = 1 << 20 // 1MB
	statusActive = "active"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	active := statusActive
	result, err := h.svc.List(r.Context(), page, limit, &active)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to list gifts", err))
		return
	}

	public := make([]PublicGift, len(result.Data))
	for i, g := range result.Data {
		public[i] = g.ToPublic()
	}
	httputil.WriteJSON(w, http.StatusOK, PublicPagedResponse{
		Data:  public,
		Page:  result.Page,
		Limit: result.Limit,
		Total: result.Total,
	})
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid gift id", err))
		return
	}

	g, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to get gift", err))
		return
	}
	if g.Status != statusActive {
		httputil.WriteError(w, r, apperror.NotFound("gift not found"))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, g.ToPublic())
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var input CreateGiftInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid gift payload", err))
		return
	}

	g, err := h.svc.Create(r.Context(), input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to create gift", err))
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, g)
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid gift id", err))
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var input UpdateGiftInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid gift payload", err))
		return
	}

	g, err := h.svc.Update(r.Context(), id, input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to update gift", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, g)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid gift id", err))
		return
	}

	if err := h.svc.Delete(r.Context(), id, userRACF); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to delete gift", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
