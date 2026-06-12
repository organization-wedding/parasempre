package gift

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
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
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	active := statusActive
	filter := ListFilter{Status: &active}
	if s := strings.TrimSpace(q.Get("search")); s != "" {
		filter.Search = &s
	}
	if v, err := strconv.ParseInt(q.Get("price_min"), 10, 64); err == nil {
		filter.PriceMin = &v
	}
	if v, err := strconv.ParseInt(q.Get("price_max"), 10, 64); err == nil {
		filter.PriceMax = &v
	}
	if s := q.Get("sort"); s != "" {
		filter.Sort = &s
	}

	result, err := h.svc.List(r.Context(), page, limit, filter)
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

const maxCSVSize = 5 << 20 // 5MB

func (h *Handler) HandlePreviewImport(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxCSVSize)

	if err := r.ParseMultipartForm(maxCSVSize); err != nil {
		httputil.WriteError(w, r, apperror.Validation("invalid multipart form"))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		httputil.WriteError(w, r, apperror.Validation("missing \"file\" field"))
		return
	}
	defer file.Close()

	preview, err := h.svc.PreviewImport(r.Context(), file)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to preview CSV", err))
		return
	}

	httputil.WriteJSON(w, http.StatusOK, preview)
}

func (h *Handler) HandleCommitImport(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var req CommitImportRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid request body", err))
		return
	}

	resp, err := h.svc.CommitImport(r.Context(), req.Rows, userRACF)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to import gifts", err))
		return
	}

	// Sempre 200 com body auto-descritivo (created + skipped). O frontend
	// decide a mensagem a partir desses contadores.
	httputil.WriteJSON(w, http.StatusOK, resp)
}

const scrapeRequestTimeout = 28 * time.Second

func (h *Handler) HandleScrapePreview(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	var input ScrapePreviewRequest
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid scrape payload", err))
		return
	}

	if err := validate.Struct(input); err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), scrapeRequestTimeout)
	defer cancel()

	preview, err := h.svc.ScrapePreview(ctx, input.URL, userRACF)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to scrape product URL", err))
		return
	}

	httputil.WriteJSON(w, http.StatusOK, preview)
}
