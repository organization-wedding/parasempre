package main

import (
	"net/http"

	"github.com/ferjunior7/parasempre/backend/internal/auth"
	"github.com/ferjunior7/parasempre/backend/internal/guest"
	"github.com/ferjunior7/parasempre/backend/internal/middleware"
	"github.com/ferjunior7/parasempre/backend/internal/user"
)

type routeDeps struct {
	auth   *auth.Handler
	guest  *guest.Handler
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

	guestsAdmin := newGroup(mux, authMW, coupleMW)
	guestsAdmin.handle("POST /api/guests", d.guest.HandleCreate)
	guestsAdmin.handle("PUT /api/guests/{id}", d.guest.HandleUpdate)
	guestsAdmin.handle("DELETE /api/guests/{id}", d.guest.HandleDelete)
	guestsAdmin.handle("POST /api/guests/import", d.guest.HandleImport)

	mux.HandleFunc("GET /api/users/check", d.user.HandleCheck)

	users := newGroup(mux, authMW)
	users.handle("GET /api/users/me", d.user.HandleMe)

	usersDev := newGroup(mux, middleware.DevOnly(d.appEnv))
	usersDev.handle("GET /api/users", d.user.HandleList)
	usersDev.handle("POST /api/users", d.user.HandleRegister)
}
