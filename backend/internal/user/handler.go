package user

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/apperr"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
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
		httputil.WriteError(w, r, apperr.Validation("Dados inválidos na requisição"))
		return
	}

	u, err := h.svc.Register(r.Context(), input)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, u)
}

func (h *Handler) handleCheck(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")

	resp, err := h.svc.CheckByPhone(r.Context(), phone)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) handleList(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) handleMe(w http.ResponseWriter, r *http.Request) {
	uracf := r.Header.Get("user-racf")
	if uracf == "" {
		httputil.WriteError(w, r, apperr.Validation("Autenticação necessária"))
		return
	}

	u, err := h.svc.GetMe(r.Context(), uracf)
	if err != nil {
		httputil.WriteError(w, r, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"role": u.Role})
}
