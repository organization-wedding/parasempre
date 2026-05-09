package main

import (
	"context"
	"errors"
	"fmt"
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
	"github.com/ferjunior7/parasempre/backend/internal/gift"
	"github.com/ferjunior7/parasempre/backend/internal/giftmessage"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
	"github.com/ferjunior7/parasempre/backend/internal/payment"
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
	giftRepo := gift.NewPostgresRepository(pool)
	userRepo := user.NewPostgresRepository(pool)
	otpRepo := auth.NewPostgresOTPRepository(pool)
	paymentRepo := payment.NewPostgresRepository(pool)
	giftMessageRepo := giftmessage.NewPostgresRepository(pool)
	txRunner := database.NewTxRunner(pool)

	jwtExpiry, err := time.ParseDuration(cfg.JWTExpiry)
	if err != nil {
		jwtExpiry = 720 * time.Hour
	}
	jwtSvc := auth.NewJWTService(cfg.JWTSecret, jwtExpiry)

	userSvc := user.NewServiceWithTx(userRepo, guestRepo)
	guestSvc := guest.NewService(guestRepo, userSvc, txRunner)
	guestHandler := guest.NewHandler(guestSvc)

	var firecrawlClient gift.ProductScraper
	const defaultFirecrawlURL = "https://api.firecrawl.dev"
	if cfg.FirecrawlAPIKey != "" || cfg.FirecrawlURL != defaultFirecrawlURL {
		firecrawlClient = gift.NewFirecrawlClient(cfg.FirecrawlAPIKey, cfg.FirecrawlURL)
		slog.Info("firecrawl: enabled", "url", cfg.FirecrawlURL, "auth", cfg.FirecrawlAPIKey != "")
	} else {
		slog.Warn("firecrawl: disabled (set FIRECRAWL_API_KEY for cloud or FIRECRAWL_URL for self-host)")
	}
	giftSvc := gift.NewService(giftRepo, txRunner, firecrawlClient)
	giftHandler := gift.NewHandler(giftSvc)
	userHandler := user.NewHandler(userSvc, cfg.AppEnv)

	var paymentHandler *payment.Handler
	var purchaseLimiterMW, webhookLimiterMW func(http.Handler) http.Handler
	if cfg.MercadoPagoAccessToken != "" && cfg.MercadoPagoWebhookSecret != "" {
		mpClient := payment.NewMercadoPagoClient(
			cfg.MercadoPagoAccessToken,
			cfg.MercadoPagoBaseURL,
			cfg.MercadoPagoWebhookSecret,
			"",
		)
		paymentSvc := payment.NewService(paymentRepo, txRunner, mpClient, giftFinderAdapter{repo: giftRepo}, userRepo)
		paymentHandler = payment.NewHandler(paymentSvc, mpClient)

		purchaseLimiter := middleware.NewRateLimiter(rate.Every(12*time.Second), 5)
		webhookLimiter := middleware.NewRateLimiter(rate.Limit(30), 60)
		purchaseLimiterMW = purchaseLimiter.MiddlewareWithKey(func(r *http.Request) string {
			if uid := middleware.UserIDFromContext(r.Context()); uid != 0 {
				return fmt.Sprintf("user:%d", uid)
			}
			return middleware.IPKey(r)
		})
		webhookLimiterMW = webhookLimiter.Middleware()

		slog.Info("mercado pago: enabled", "base_url", cfg.MercadoPagoBaseURL)
	} else {
		slog.Warn("mercado pago: disabled (set MERCADO_PAGO_ACCESS_TOKEN and MERCADO_PAGO_WEBHOOK_SECRET to enable)")
	}

	var giftMessageHandler *giftmessage.Handler
	var messageLimiterMW func(http.Handler) http.Handler
	{
		var storage giftmessage.Storage
		if cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != "" {
			storage = giftmessage.NewSupabaseStorage(cfg.SupabaseURL, cfg.SupabaseStorageBucket, cfg.SupabaseServiceRoleKey)
			slog.Info("supabase storage: enabled", "bucket", cfg.SupabaseStorageBucket)
		} else {
			slog.Warn("supabase storage: disabled (set SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY to enable media uploads)")
		}
		ttl := time.Duration(cfg.GiftMessageSignedURLTTLSecs) * time.Second
		txFinder := payment.NewMessageTxFinder(paymentRepo)
		giftMessageSvc := giftmessage.NewService(giftMessageRepo, txFinder, storage, userRepo, ttl)
		giftMessageHandler = giftmessage.NewHandler(giftMessageSvc)

		messageLimiter := middleware.NewRateLimiter(rate.Every(12*time.Second), 5)
		messageLimiterMW = messageLimiter.MiddlewareWithKey(func(r *http.Request) string {
			if uid := middleware.UserIDFromContext(r.Context()); uid != 0 {
				return fmt.Sprintf("user:%d", uid)
			}
			return middleware.IPKey(r)
		})
	}

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

	mux := http.NewServeMux()
	registerRoutes(mux, routeDeps{
		auth:            authHandler,
		guest:           guestHandler,
		gift:            giftHandler,
		user:            userHandler,
		payment:         paymentHandler,
		giftMessage:     giftMessageHandler,
		jwt:             jwtSvc,
		appEnv:          cfg.AppEnv,
		purchaseLimiter: purchaseLimiterMW,
		webhookLimiter:  webhookLimiterMW,
		messageLimiter:  messageLimiterMW,
	})

	handler := middleware.Chain(mux,
		middleware.Recovery,
		middleware.Logger,
		middleware.SecurityHeaders(cfg.AppEnv),
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

type giftFinderAdapter struct {
	repo *gift.PostgresRepository
}

func (a giftFinderAdapter) GetByID(ctx context.Context, id int64) (*payment.GiftSnapshot, error) {
	g, err := a.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &payment.GiftSnapshot{
		ID:         g.ID,
		Name:       g.Name,
		PriceCents: g.PriceCents,
		Status:     g.Status,
	}, nil
}
