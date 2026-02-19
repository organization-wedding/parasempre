package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users", h.handleRegister)
	mux.HandleFunc("GET /api/users/check", h.handleCheck)
}

func (h *Handler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		slog.Error("register: invalid request body", "error", err)
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	u, err := h.svc.Register(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrAlreadyRegistered) {
			slog.Warn("register: user already registered", "phone", input.Phone)
			writeError(w, http.StatusConflict, "user already registered for this guest")
			return
		}
		if errors.Is(err, ErrGuestNotFound) {
			slog.Warn("register: guest not found", "phone", input.Phone)
			writeError(w, http.StatusNotFound, "no guest found with this phone")
			return
		}
		if errors.Is(err, ErrURACFTaken) {
			slog.Warn("register: uracf already in use", "uracf", input.URACF)
			writeError(w, http.StatusConflict, "uracf already in use")
			return
		}
		slog.Error("register: failed to register user", "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (h *Handler) handleCheck(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")

	resp, err := h.svc.CheckByPhone(r.Context(), phone)
	if err != nil {
		slog.Warn("check: failed to check phone", "phone", phone, "error", err)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
