package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"time"

	"github.com/ferjunior7/parasempre/backend/internal/apperror"
)

const otpTTL = 5 * time.Minute

type OTPService struct {
	repo   OTPRepository
	sender WhatsAppSender
}

func NewOTPService(repo OTPRepository, sender WhatsAppSender) *OTPService {
	return &OTPService{repo: repo, sender: sender}
}

func (s *OTPService) SendOTP(ctx context.Context, phone string) error {
	wait, err := s.repo.SendCooldown(ctx, phone)
	if err != nil {
		slog.Error("otp: failed to check cooldown", "phone", phone, "error", err)
		return apperror.Internal("failed to check OTP rate limit", err)
	}
	if wait > 0 {
		return apperror.RateLimited(
			fmt.Sprintf("aguarde %d segundos para solicitar um novo código", int(wait.Seconds())),
			wait,
		)
	}

	code, err := generateCode()
	if err != nil {
		return apperror.Internal("failed to generate OTP", err)
	}

	expiresAt := time.Now().Add(otpTTL)
	if err := s.repo.Create(ctx, phone, code, expiresAt); err != nil {
		slog.Error("otp: failed to save code", "phone", phone, "error", err)
		return apperror.Internal("failed to save OTP", err)
	}

	msg := fmt.Sprintf("*ParaSempre* - Gerenciamento de Convidados\n\nSeu codigo de verificacao: *%s*\n\nUse este codigo para acessar sua conta. Ele e valido por 5 minutos.\n\nPor seguranca, nao compartilhe este codigo com ninguem. A equipe ParaSempre nunca solicitara seu codigo.", code)
	if err := s.sender.SendMessage(phone, msg); err != nil {
		slog.Error("otp: failed to send message", "phone", phone, "error", err)
		return apperror.Internal("failed to send OTP via WhatsApp", err)
	}

	slog.Info("otp: code sent", "phone", phone)
	return nil
}

func (s *OTPService) VerifyOTP(ctx context.Context, phone, code string) error {
	verified, err := s.repo.VerifyAndMarkUsed(ctx, phone, code)
	if err != nil {
		slog.Error("otp: verification failed", "phone", phone, "error", err)
		return apperror.Internal("failed to verify OTP", err)
	}
	if !verified {
		return apperror.Unauthorized("invalid or expired code")
	}

	return nil
}

func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
