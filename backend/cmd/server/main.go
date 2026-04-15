package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/time/rate"

	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/config"
	"github.com/ferjunior7/parasempre/backend/internal/database"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
	"github.com/ferjunior7/parasempre/backend/internal/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", "error", err)
		os.Exit(1)
	}

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer connectCancel()

	pool, err := database.Connect(connectCtx, cfg.DB)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("connected to database")

	guestRepo := guest.NewPostgresRepository(pool)
	userRepo := user.NewPostgresRepository(pool)
	otpRepo := auth.NewPostgresOTPRepository(pool)
	txRunner := database.NewTxRunner(pool)

	jwtExpiry, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil {
		jwtExpiry = 720 * time.Hour
	}
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, jwtExpiry)

	userSvc := user.NewServiceWithTx(userRepo, guestRepo)
	guestSvc := guest.NewService(guestRepo, userSvc, txRunner)
	guestHandler := guest.NewHandler(guestSvc)
	userHandler := user.NewHandler(userSvc, cfg.AppEnv)

	var whatsappSender auth.WhatsAppSender
	if cfg.EvoAPIURL != "" && cfg.EvoAPIKey != "" {
		whatsappSender = auth.NewEvoAPISender(cfg.EvoAPIURL, cfg.EvoAPIKey, cfg.EvoAPIInstance)
	} else {
		whatsappSender = &logSender{}
	}
	otpSvc := auth.NewOTPService(otpRepo, whatsappSender)
	authHandler := auth.NewHandler(otpSvc, jwtSvc, userSvc, userSvc, userSvc)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()
	userSvc.SeedCouple(seedCtx,
		user.CoupleData{URACF: cfg.Couple.Groom.URACF, Phone: cfg.Couple.Groom.Phone},
		user.CoupleData{URACF: cfg.Couple.Bride.URACF, Phone: cfg.Couple.Bride.Phone},
	)

	otpMinuteLimiter := middleware.NewRateLimiter(rate.Every(time.Minute), 1)
	defer otpMinuteLimiter.Close()
	otpHourLimiter := middleware.NewRateLimiter(rate.Every(time.Hour/10), 10)
	defer otpHourLimiter.Close()

	mux := http.NewServeMux()
	registerRoutes(mux, routeDeps{
		auth:             authHandler,
		guest:            guestHandler,
		user:             userHandler,
		jwt:              jwtSvc,
		appEnv:           cfg.AppEnv,
		otpMinuteLimiter: otpMinuteLimiter,
		otpHourLimiter:   otpHourLimiter,
	})

	handler := middleware.Chain(mux,
		middleware.Recovery,
		middleware.Logger,
		middleware.CORS(cfg.CORSOrigin),
	)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		slog.Info("server starting on port 8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
	slog.Info("server stopped")
}

type logSender struct{}

func (s *logSender) SendMessage(phone, message string) error {
	slog.Info("OTP (dev mode)", "phone", phone, "message", message)
	return nil
}
