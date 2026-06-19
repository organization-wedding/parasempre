package guest

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

func validatePathPhone(phone string) error {
	if !validate.BRPhoneRegex.MatchString(phone) {
		return apperror.Validation("invalid phone number format")
	}
	return nil
}

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

	filters := ListFilters{
		Search:       strings.TrimSpace(q.Get("search")),
		Relationship: q.Get("relationship"),
		Attending:    q.Get("attending"),
	}

	result, err := h.svc.List(r.Context(), page, limit, filters)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to list guests", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) HandleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.Stats(r.Context())
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to compute guest stats", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handler) HandleListMyFamily(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	guests, err := h.svc.ListMyFamily(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to list family guests", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guests)
}

func (h *Handler) HandleBatchConfirm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	var input BatchConfirmInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid batch payload", err))
		return
	}

	guests, err := h.svc.SetConfirmedBatch(r.Context(), input, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to update batch confirmation", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guests)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid guest id", err))
		return
	}

	guest, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to get guest", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	var input CreateGuestInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid guest payload", err))
		return
	}

	guest, err := h.svc.Create(r.Context(), input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to create guest", err))
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, guest)
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid guest id", err))
		return
	}

	var input UpdateGuestInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid guest payload", err))
		return
	}

	guest, err := h.svc.Update(r.Context(), id, input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to update guest", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid guest id", err))
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to delete guest", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleConfirm(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid guest id", err))
		return
	}

	guest, err := h.svc.Confirm(r.Context(), id, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to confirm guest", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid guest id", err))
		return
	}

	guest, err := h.svc.Cancel(r.Context(), id, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to cancel guest", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleConfirmByPhone(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	phone := r.PathValue("phone")
	if err := validatePathPhone(phone); err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	guest, err := h.svc.ConfirmByPhone(r.Context(), phone, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to confirm guest by phone", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleCancelByPhone(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	phone := r.PathValue("phone")
	if err := validatePathPhone(phone); err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	guest, err := h.svc.CancelByPhone(r.Context(), phone, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to cancel guest by phone", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleConfirmFamily(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	familyGroupStr := r.PathValue("familyGroup")
	familyGroup, err := strconv.ParseInt(familyGroupStr, 10, 64)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid family group", err))
		return
	}

	guests, err := h.svc.ConfirmFamily(r.Context(), familyGroup, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to confirm family group", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guests)
}

func (h *Handler) HandleCancelFamily(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	familyGroupStr := r.PathValue("familyGroup")
	familyGroup, err := strconv.ParseInt(familyGroupStr, 10, 64)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid family group", err))
		return
	}

	guests, err := h.svc.CancelFamily(r.Context(), familyGroup, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to cancel family group", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guests)
}

func (h *Handler) HandleConfirmFamilyByPhone(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	phone := r.PathValue("phone")
	if err := validatePathPhone(phone); err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	guests, err := h.svc.ConfirmFamilyByPhone(r.Context(), phone, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to confirm family by phone", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guests)
}

func (h *Handler) HandleCancelFamilyByPhone(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID == 0 {
		httputil.WriteError(w, r, apperror.Unauthorized("authentication required"))
		return
	}

	phone := r.PathValue("phone")
	if err := validatePathPhone(phone); err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	guests, err := h.svc.CancelFamilyByPhone(r.Context(), phone, userID)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to cancel family by phone", err))
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guests)
}

func (h *Handler) HandleImport(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.WriteError(w, r, apperror.Validation("file is required"))
		return
	}
	defer file.Close()

	var guests []CreateGuestInput
	ext := strings.ToLower(filepath.Ext(header.Filename))

	switch ext {
	case ".csv":
		guests, err = ParseCSV(file)
	case ".xlsx":
		guests, err = ParseXLSX(file)
	default:
		httputil.WriteError(w, r, apperror.Validation("unsupported file format: use .csv or .xlsx"))
		return
	}

	if err != nil {
		slog.Error("import: failed to parse file", "extension", ext, "error", err)
		httputil.WriteError(w, r, apperror.Validation("failed to parse uploaded file"))
		return
	}

	result := h.svc.Import(r.Context(), guests, userRACF)

	status := http.StatusOK
	if result.ErrorCount > 0 && result.SuccessCount > 0 {
		status = http.StatusMultiStatus
	} else if result.ErrorCount > 0 {
		status = http.StatusBadRequest
	}

	httputil.WriteJSON(w, status, result)
}
