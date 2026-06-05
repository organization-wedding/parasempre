package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
	"github.com/ferjunior7/parasempre/backend/internal/httputil"
	"github.com/ferjunior7/parasempre/backend/internal/validate"
)

type UserFinder interface {
	FindOrCreateByPhone(ctx context.Context, phone string) (userID int64, uracf string, role string, err error)
	FindByURACF(ctx context.Context, uracf string) (int64, string, string, error)
}

type PhoneChecker interface {
	PhoneExists(ctx context.Context, phone string) (bool, error)
}

type LoginRecorder interface {
	RecordLogin(ctx context.Context, userID int64)
}

type Handler struct {
	otpSvc        *OTPService
	jwtSvc        *JWTService
	userFinder    UserFinder
	phoneCheck    PhoneChecker
	loginRecorder LoginRecorder
}

func NewHandler(otpSvc *OTPService, jwtSvc *JWTService, userFinder UserFinder, phoneCheck PhoneChecker, loginRecorder LoginRecorder) *Handler {
	return &Handler{
		otpSvc:        otpSvc,
		jwtSvc:        jwtSvc,
		userFinder:    userFinder,
		phoneCheck:    phoneCheck,
		loginRecorder: loginRecorder,
	}
}

func (h *Handler) HandleSendOTP(w http.ResponseWriter, r *http.Request) {
	var input SendOTPInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid OTP payload", err))
		return
	}
	if err := validate.Struct(input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid OTP input", err))
		return
	}

	exists, err := h.phoneCheck.PhoneExists(r.Context(), input.Phone)
	if err != nil {
		slog.Error("auth: phone check failed", "phone", input.Phone, "error", err)
		httputil.WriteError(w, r, apperror.Internal("failed to verify phone", err))
		return
	}
	if !exists {
		httputil.WriteError(w, r, apperror.NotFound("no guest found with this phone"))
		return
	}

	if err := h.otpSvc.SendOTP(r.Context(), input.Phone); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to send OTP", err))
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"message": "OTP sent"})
}

func (h *Handler) HandleVerifyOTP(w http.ResponseWriter, r *http.Request) {
	var input VerifyOTPInput
	if err := httputil.DecodeJSON(r, &input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid verification payload", err))
		return
	}
	if err := validate.Struct(input); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("invalid verification input", err))
		return
	}

	if err := h.otpSvc.VerifyOTP(r.Context(), input.Phone, input.Code); err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to verify OTP", err))
		return
	}

	userID, uracf, role, err := h.userFinder.FindOrCreateByPhone(r.Context(), input.Phone)
	if err != nil {
		httputil.WriteError(w, r, apperror.WrapIfNotApp("failed to find or create user", err))
		return
	}

	token, err := h.jwtSvc.Generate(userID, uracf, role)
	if err != nil {
		slog.Error("auth: jwt generation failed", "user_id", userID, "error", err)
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

const loginRecordTimeout = 5 * time.Second

func (h *Handler) recordLoginAsync(userID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), loginRecordTimeout)
	defer cancel()
	h.loginRecorder.RecordLogin(ctx, userID)
}
