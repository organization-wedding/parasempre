package user

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc    *Service
	appEnv string
}

func NewHandler(svc *Service, appEnv string) *Handler {
	return &Handler{svc: svc, appEnv: appEnv}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users", h.handleRegister)
	mux.HandleFunc("GET /api/users/check", h.handleCheck)
	mux.HandleFunc("GET /api/users/me", h.handleMe)
	if h.appEnv != "production" {
		mux.HandleFunc("GET /api/users", h.handleList)
	}
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

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		slog.Error("list users: failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	uracf := r.Header.Get("user-racf")
	if uracf == "" {
		writeError(w, http.StatusUnauthorized, "user-racf header required")
		return
	}

	u, err := h.svc.GetMe(r.Context(), uracf)
	if err != nil {
		slog.Error("me: lookup failed", "uracf", uracf, "error", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if u == nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"role": u.Role})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
