package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
)

type DevLoginHandler struct {
	jwtSvc        *JWTService
	userFinder    UserFinder
	loginRecorder LoginRecorder
}

func NewDevLoginHandler(jwtSvc *JWTService, userFinder UserFinder, loginRecorder LoginRecorder) *DevLoginHandler {
	return &DevLoginHandler{
		jwtSvc:        jwtSvc,
		userFinder:    userFinder,
		loginRecorder: loginRecorder,
	}
}

func (h *DevLoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	uracf := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("uracf")))
	if uracf == "" {
		httputil.WriteError(w, r, apperror.Validation("uracf query param is required"))
		return
	}

	userID, resolvedURACF, role, err := h.userFinder.FindByURACF(r.Context(), uracf)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to find user by uracf", err))
		return
	}

	slog.Info("dev-login: authenticated", "user_id", userID, "role", role)

	token, err := h.jwtSvc.Generate(userID, resolvedURACF, role)
	if err != nil {
		slog.Error("dev-login: jwt generation failed", "user_id", userID, "error", err)
		httputil.WriteError(w, r, apperror.Internal("failed to generate token", err))
		return
	}

	go h.recordLoginAsync(userID)

	httputil.WriteJSON(w, http.StatusOK, TokenResponse{
		Token: token,
		Role:  role,
		URACF: resolvedURACF,
	})
}

func (h *DevLoginHandler) recordLoginAsync(userID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), loginRecordTimeout)
	defer cancel()
	h.loginRecorder.RecordLogin(ctx, userID)
}
