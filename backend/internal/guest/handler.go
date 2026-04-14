package guest

import (
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
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

	result, err := h.svc.List(r.Context(), page, limit)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	guest, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	var input CreateGuestInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	guest, err := h.svc.Create(r.Context(), input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, guest)
}

func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	var input UpdateGuestInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	guest, err := h.svc.Update(r.Context(), id, input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleConfirm(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	guest, err := h.svc.Confirm(r.Context(), id, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	id, err := httputil.PathID(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	guest, err := h.svc.Cancel(r.Context(), id, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleConfirmByPhone(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	phone := r.PathValue("phone")
	guest, err := h.svc.ConfirmByPhone(r.Context(), phone, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleCancelByPhone(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	phone := r.PathValue("phone")
	guest, err := h.svc.CancelByPhone(r.Context(), phone, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) HandleImport(w http.ResponseWriter, r *http.Request) {
	userRACF := middleware.UserRACFFromContext(r.Context())

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.WriteErrorMsg(w, http.StatusBadRequest, "file is required")
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
		httputil.WriteErrorMsg(w, http.StatusBadRequest, "unsupported file format: use .csv or .xlsx")
		return
	}

	if err != nil {
		slog.Error("import: failed to parse file", "extension", ext, "error", err)
		httputil.WriteErrorMsg(w, http.StatusBadRequest, "failed to parse file: "+err.Error())
		return
	}

	var successCount int
	var rowErrors []ImportRowError
	const dataRowStart = 2 // row 1 is the header in CSV/XLSX files
	for i, input := range guests {
		rowNumber := i + dataRowStart
		if _, err := h.svc.Create(r.Context(), input, userRACF); err != nil {
			slog.Warn("import: failed to create guest", "row", rowNumber, "error", err)
			rowErrors = append(rowErrors, ImportRowError{Row: rowNumber, Error: err.Error()})
			continue
		}
		successCount++
	}

	if rowErrors == nil {
		rowErrors = []ImportRowError{}
	}

	status := http.StatusOK
	if len(rowErrors) > 0 && successCount > 0 {
		status = http.StatusMultiStatus
	} else if len(rowErrors) > 0 {
		status = http.StatusBadRequest
	}

	httputil.WriteJSON(w, status, ImportResponse{
		SuccessCount: successCount,
		ErrorCount:   len(rowErrors),
		Total:        len(guests),
		Errors:       rowErrors,
	})
}
