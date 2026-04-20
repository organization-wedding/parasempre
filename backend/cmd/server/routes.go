package main

import (
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/gift"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
	"github.com/ferjunior7/parasempre/backend/internal/user"
)

type routeDeps struct {
	auth   *auth.Handler
	guest  *guest.Handler
	gift   *gift.Handler
	user   *user.Handler
	jwt    *auth.JWTService
	appEnv string
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

	guests := newGroup(mux, authMW)
	guests.handle("GET /api/guests", d.guest.HandleList)
	guests.handle("GET /api/guests/{id}", d.guest.HandleGet)
	guests.handle("PATCH /api/guests/{id}/confirm", d.guest.HandleConfirm)
	guests.handle("PATCH /api/guests/{id}/cancel", d.guest.HandleCancel)
	guests.handle("PATCH /api/guests/phone/{phone}/confirm", d.guest.HandleConfirmByPhone)
	guests.handle("PATCH /api/guests/phone/{phone}/cancel", d.guest.HandleCancelByPhone)
	guests.handle("PATCH /api/guests/family/{familyGroup}/confirm", d.guest.HandleConfirmFamily)
	guests.handle("PATCH /api/guests/family/{familyGroup}/cancel", d.guest.HandleCancelFamily)
	guests.handle("PATCH /api/guests/family/phone/{phone}/confirm", d.guest.HandleConfirmFamilyByPhone)
	guests.handle("PATCH /api/guests/family/phone/{phone}/cancel", d.guest.HandleCancelFamilyByPhone)

	guestsAdmin := newGroup(mux, authMW, coupleMW)
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

	users := newGroup(mux, authMW)
	users.handle("GET /api/users/me", d.user.HandleMe)

	usersAdmin := newGroup(mux, authMW, coupleMW)
	usersAdmin.handle("GET /api/users/check", d.user.HandleCheck)
	usersAdmin.handle("PATCH /api/users/{id}", d.user.HandleUpdate)
	usersAdmin.handle("DELETE /api/users/{id}", d.user.HandleDelete)
}
