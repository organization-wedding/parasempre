package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
)

type DevLoginHandler struct {
	jwtSvc        *JWTService
	userFinder    UserFinder
	loginRecorder LoginRecorder
	phone         string
}

func NewDevLoginHandler(jwtSvc *JWTService, userFinder UserFinder, loginRecorder LoginRecorder, phone string) *DevLoginHandler {
	return &DevLoginHandler{
		jwtSvc:        jwtSvc,
		userFinder:    userFinder,
		loginRecorder: loginRecorder,
		phone:         phone,
	}
}

func (h *DevLoginHandler) Handle(w http.ResponseWriter, r *http.Request) {
	slog.Warn("dev-login: invoked", "phone", h.phone)

	userID, uracf, role, err := h.userFinder.FindOrCreateByPhone(r.Context(), h.phone)
	if err != nil {
		slog.Error("dev-login: user lookup failed — verify SeedCouple ran or TEST_LOGIN_PHONE points to a registered phone", "phone", h.phone, "error", err)
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to find user by phone", err))
		return
	}

	token, err := h.jwtSvc.Generate(userID, uracf, role)
	if err != nil {
		slog.Error("dev-login: jwt generation failed", "user_id", userID, "error", err)
		httputil.WriteError(w, r, apperror.Internal("failed to generate token", err))
		return
	}

	go h.recordLoginAsync(userID)

	httputil.WriteJSON(w, http.StatusOK, TokenResponse{
		Token: token,
		Role:  role,
		URACF: uracf,
	})
}

func (h *DevLoginHandler) recordLoginAsync(userID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), loginRecordTimeout)
	defer cancel()
	h.loginRecorder.RecordLogin(ctx, userID)
}
