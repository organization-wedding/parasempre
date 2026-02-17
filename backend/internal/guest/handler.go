package guest

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/guests", h.handleList)
	mux.HandleFunc("POST /api/guests", h.handleCreate)
	mux.HandleFunc("GET /api/guests/{id}", h.handleGet)
	mux.HandleFunc("PUT /api/guests/{id}", h.handleUpdate)
	mux.HandleFunc("DELETE /api/guests/{id}", h.handleDelete)
	mux.HandleFunc("POST /api/guests/import", h.handleImport)
}

var racfRegex = regexp.MustCompile(`^[A-Za-z0-9]{5}$`)

func getUserRACF(r *http.Request) (string, error) {
	racf := strings.TrimSpace(r.Header.Get("user-racf"))
	if racf == "" {
		return "", fmt.Errorf("header user-racf is required")
	}
	if !racfRegex.MatchString(racf) {
		return "", fmt.Errorf("user-racf must be exactly 5 alphanumeric characters")
	}
	return strings.ToUpper(racf), nil
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	guests, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list guests")
		return
	}
	writeJSON(w, http.StatusOK, guests)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid guest ID")
		return
	}

	guest, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "guest not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get guest")
		return
	}
	writeJSON(w, http.StatusOK, guest)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var input CreateGuestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	guest, err := h.svc.Create(r.Context(), input, userRACF)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, guest)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid guest ID")
		return
	}

	var input UpdateGuestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	guest, err := h.svc.Update(r.Context(), id, input, userRACF)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "guest not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, guest)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	_ = userRACF

	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid guest ID")
		return
	}

	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "guest not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete guest")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleImport(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
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
		writeError(w, http.StatusBadRequest, "unsupported file format: use .csv or .xlsx")
		return
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse file: "+err.Error())
		return
	}

	var created int
	var errs []string
	for i, input := range guests {
		if _, err := h.svc.Create(r.Context(), input, userRACF); err != nil {
			slog.Warn("import: failed to create guest", "row", i+2, "error", err)
			errs = append(errs, err.Error())
			continue
		}
		created++
	}

	status := http.StatusOK
	if len(errs) > 0 {
		status = http.StatusBadRequest
	}

	writeJSON(w, status, map[string]any{
		"imported": created,
		"errors":   errs,
		"total":    len(guests),
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
