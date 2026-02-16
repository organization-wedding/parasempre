package user

import (
	"encoding/json"
	"errors"
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
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	u, err := h.svc.Register(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrAlreadyRegistered) {
			writeError(w, http.StatusConflict, "user already registered for this guest")
			return
		}
		if errors.Is(err, ErrGuestNotFound) {
			writeError(w, http.StatusNotFound, "no guest found with this phone")
			return
		}
		if errors.Is(err, ErrURACFTaken) {
			writeError(w, http.StatusConflict, "uracf already in use")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, u)
}

func (h *Handler) handleCheck(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")

	resp, err := h.svc.CheckByPhone(r.Context(), phone)
	if err != nil {
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
