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
		writeError(w, http.StatusBadRequest, "Dados inválidos na requisição")
		return
	}

	u, err := h.svc.Register(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrAlreadyRegistered) {
			slog.Warn("register: user already registered", "phone", input.Phone)
			writeError(w, http.StatusConflict, "Este convidado já possui cadastro")
			return
		}
		if errors.Is(err, ErrGuestNotFound) {
			slog.Warn("register: guest not found", "phone", input.Phone)
			writeError(w, http.StatusNotFound, "Nenhum convidado encontrado com este telefone")
			return
		}
		if errors.Is(err, ErrURACFTaken) {
			slog.Warn("register: uracf already in use", "uracf", input.URACF)
			writeError(w, http.StatusConflict, "Este identificador de acesso já está em uso")
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
		writeError(w, http.StatusInternalServerError, "Não foi possível carregar a lista de usuários")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	uracf := r.Header.Get("user-racf")
	if uracf == "" {
		writeError(w, http.StatusUnauthorized, "Autenticação necessária")
		return
	}

	u, err := h.svc.GetMe(r.Context(), uracf)
	if err != nil {
		slog.Error("me: lookup failed", "uracf", uracf, "error", err)
		writeError(w, http.StatusInternalServerError, "Não foi possível carregar os dados do usuário")
		return
	}
	if u == nil {
		writeError(w, http.StatusNotFound, "Usuário não encontrado")
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
