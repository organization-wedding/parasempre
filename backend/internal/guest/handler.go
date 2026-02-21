package guest

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/ferjunior7/parasempre/backend/internal/apperr"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
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
		return "", apperr.Validation("autenticação necessária")
	}
	if !racfRegex.MatchString(racf) {
		return "", apperr.Validation("credencial de acesso inválida")
	}
	return strings.ToUpper(racf), nil
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	guests, err := h.svc.List(r.Context())
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guests)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		httputil.WriteError(w, r, apperr.Validation("ID de convidado inválido"))
		return
	}

	guest, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) handleCreate(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	var input CreateGuestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("create: invalid request body", "error", err)
		httputil.WriteError(w, r, apperr.Validation("Dados inválidos na requisição"))
		return
	}

	guest, err := h.svc.Create(r.Context(), input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, guest)
}

func (h *Handler) handleUpdate(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	id, err := parseID(r)
	if err != nil {
		httputil.WriteError(w, r, apperr.Validation("ID de convidado inválido"))
		return
	}

	var input UpdateGuestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("update: invalid request body", "id", id, "error", err)
		httputil.WriteError(w, r, apperr.Validation("Dados inválidos na requisição"))
		return
	}

	guest, err := h.svc.Update(r.Context(), id, input, userRACF)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, guest)
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	_ = userRACF

	id, err := parseID(r)
	if err != nil {
		httputil.WriteError(w, r, apperr.Validation("ID de convidado inválido"))
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleImport(w http.ResponseWriter, r *http.Request) {
	userRACF, err := getUserRACF(r)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		slog.Error("import: missing file", "error", err)
		httputil.WriteError(w, r, apperr.Validation("Arquivo não enviado"))
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
		slog.Warn("import: unsupported file format", "extension", ext)
		httputil.WriteError(w, r, apperr.Validation("Formato não suportado. Use .csv ou .xlsx"))
		return
	}

	if err != nil {
		slog.Error("import: failed to parse file", "extension", ext, "error", err)
		httputil.WriteError(w, r, apperr.Validation("Não foi possível processar o arquivo"))
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

	httputil.WriteJSON(w, status, map[string]any{
		"imported": created,
		"errors":   errs,
		"total":    len(guests),
	})
}
