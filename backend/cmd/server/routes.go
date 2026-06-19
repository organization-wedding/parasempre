package main

import (
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/gift"
	"github.com/ferjunior7/parasempre/backend/internal/giftmessage"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
	"github.com/ferjunior7/parasempre/backend/internal/payment"
	"github.com/ferjunior7/parasempre/backend/internal/user"
)

type routeDeps struct {
	auth            *auth.Handler
	devLogin        *auth.DevLoginHandler
	guest           *guest.Handler
	gift            *gift.Handler
	user            *user.Handler
	payment         *payment.Handler
	giftMessage     *giftmessage.Handler
	jwt             *auth.JWTService
	appEnv          string
	purchaseLimiter func(http.Handler) http.Handler
	webhookLimiter  func(http.Handler) http.Handler
	messageLimiter  func(http.Handler) http.Handler
}

type routeGroup struct {
	mux *http.ServeMux
	mws []func(http.Handler) http.Handler
}

func newGroup(mux *http.ServeMux, mws ...func(http.Handler) http.Handler) routeGroup {
	return routeGroup{mux: mux, mws: mws}
}

func (g routeGroup) handle(pattern string, fn http.HandlerFunc) {
	if len(g.mws) == 0 {
		g.mux.HandleFunc(pattern, fn)
		return
	}
	g.mux.Handle(pattern, middleware.Chain(fn, g.mws...))
}

func registerRoutes(mux *http.ServeMux, d routeDeps) {
	authMW := middleware.RequireAuth(d.jwt)
	coupleMW := middleware.RequireRole("groom", "bride")

	otp := newGroup(mux)
	otp.handle("POST /api/auth/otp/send", d.auth.HandleSendOTP)
	otp.handle("POST /api/auth/otp/verify", d.auth.HandleVerifyOTP)

	if d.devLogin != nil {
		dev := newGroup(mux, middleware.DevOnly(d.appEnv))
		dev.handle("POST /api/auth/dev-login", d.devLogin.Handle)
	}

	guests := newGroup(mux, authMW)
	guests.handle("GET /api/guests/my-family", d.guest.HandleListMyFamily)
	guests.handle("PATCH /api/guests/{id}/confirm", d.guest.HandleConfirm)
	guests.handle("PATCH /api/guests/{id}/cancel", d.guest.HandleCancel)
	guests.handle("PATCH /api/guests/phone/{phone}/confirm", d.guest.HandleConfirmByPhone)
	guests.handle("PATCH /api/guests/phone/{phone}/cancel", d.guest.HandleCancelByPhone)
	guests.handle("PATCH /api/guests/family/batch", d.guest.HandleBatchConfirm)
	guests.handle("PATCH /api/guests/family/{familyGroup}/confirm", d.guest.HandleConfirmFamily)
	guests.handle("PATCH /api/guests/family/{familyGroup}/cancel", d.guest.HandleCancelFamily)
	guests.handle("PATCH /api/guests/family/phone/{phone}/confirm", d.guest.HandleConfirmFamilyByPhone)
	guests.handle("PATCH /api/guests/family/phone/{phone}/cancel", d.guest.HandleCancelFamilyByPhone)

	guestsAdmin := newGroup(mux, authMW, coupleMW)
	// The guest list is the wedding's shared roster, managed only by the couple.
	// created_by used to scope these implicitly; gate them on role explicitly.
	guestsAdmin.handle("GET /api/guests", d.guest.HandleList)
	guestsAdmin.handle("GET /api/guests/stats", d.guest.HandleStats)
	guestsAdmin.handle("GET /api/guests/{id}", d.guest.HandleGet)
	guestsAdmin.handle("POST /api/guests", d.guest.HandleCreate)
	guestsAdmin.handle("PUT /api/guests/{id}", d.guest.HandleUpdate)
	guestsAdmin.handle("DELETE /api/guests/{id}", d.guest.HandleDelete)
	guestsAdmin.handle("POST /api/guests/import", d.guest.HandleImport)

	giftsPublic := newGroup(mux)
	giftsPublic.handle("GET /api/gifts", d.gift.HandleList)
	giftsPublic.handle("GET /api/gifts/{id}", d.gift.HandleGet)

	giftsAdmin := newGroup(mux, authMW, coupleMW)
	giftsAdmin.handle("POST /api/gifts", d.gift.HandleCreate)
	giftsAdmin.handle("PUT /api/gifts/{id}", d.gift.HandleUpdate)
	giftsAdmin.handle("DELETE /api/gifts/{id}", d.gift.HandleDelete)
	giftsAdmin.handle("POST /api/gifts/import/preview", d.gift.HandlePreviewImport)
	giftsAdmin.handle("POST /api/gifts/import/commit", d.gift.HandleCommitImport)
	giftsAdmin.handle("POST /api/gifts/scrape-preview", d.gift.HandleScrapePreview)

	if d.payment != nil {
		purchases := newGroup(mux, authMW, d.purchaseLimiter)
		purchases.handle("POST /api/gifts/{id}/purchase", d.payment.HandleCreatePurchase)

		webhooks := newGroup(mux, d.webhookLimiter)
		webhooks.handle("POST /api/webhooks/mercadopago", d.payment.HandleWebhook)

		me := newGroup(mux, authMW)
		me.handle("GET /api/me/purchases", d.payment.HandleListMyPurchases)
		me.handle("GET /api/me/purchases/{id}", d.payment.HandleGetMyPurchase)

		txAdmin := newGroup(mux, authMW, coupleMW)
		txAdmin.handle("GET /api/transactions", d.payment.HandleListAll)
		txAdmin.handle("GET /api/transactions/summary", d.payment.HandleSummary)
	}

	if d.giftMessage != nil {
		messagesPublic := newGroup(mux)
		messagesPublic.handle("GET /api/gifts/{id}/messages", d.giftMessage.HandleListByGift)

		messagesAuth := newGroup(mux, authMW, d.messageLimiter)
		messagesAuth.handle("POST /api/transactions/{id}/message", d.giftMessage.HandleCreate)

		messagesGet := newGroup(mux, authMW)
		messagesGet.handle("GET /api/transactions/{id}/message", d.giftMessage.HandleGetMine)

		messagesAdmin := newGroup(mux, authMW, coupleMW)
		messagesAdmin.handle("GET /api/admin/gift-messages", d.giftMessage.HandleAdminList)
		messagesAdmin.handle("DELETE /api/admin/gift-messages/{id}", d.giftMessage.HandleAdminDelete)
	}

	users := newGroup(mux, authMW)
	users.handle("GET /api/users/me", d.user.HandleMe)

	usersAdmin := newGroup(mux, authMW, coupleMW)
	usersAdmin.handle("GET /api/users/check", d.user.HandleCheck)
	usersAdmin.handle("PATCH /api/users/{id}", d.user.HandleUpdate)
	usersAdmin.handle("DELETE /api/users/{id}", d.user.HandleDelete)
}
