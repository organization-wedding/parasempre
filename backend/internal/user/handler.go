package user

import (
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
)

type Handler struct {
	svc    *Service
	appEnv string
}

func NewHandler(svc *Service, appEnv string) *Handler {
	return &Handler{svc: svc, appEnv: appEnv}
}

func (h *Handler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.HandleError(w, err)
		return
	}

	u, err := h.svc.Register(r.Context(), input)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, u)
}

func (h *Handler) HandleCheck(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")

	resp, err := h.svc.CheckByPhone(r.Context(), phone)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		httputil.HandleError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, users)
}

func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	uracf := middleware.UserRACFFromContext(r.Context())

	u, err := h.svc.GetMe(r.Context(), uracf)
	if err != nil {
		httputil.HandleError(w, err)
		return
	}
	if u == nil {
		httputil.WriteError(w, http.StatusNotFound, "user not found")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"role": u.Role})
}
