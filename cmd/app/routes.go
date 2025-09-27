package main

import (
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui"
)

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", app.getNotFound)
	mux.Handle("/static/", ui.ServeStaticFiles())

	mux.HandleFunc("GET /signup", app.getSignUp)
	mux.HandleFunc("POST /signup", app.postSignUp)
	mux.HandleFunc("GET /signin", app.getSignIn)
	mux.HandleFunc("POST /signin", app.postSignIn)
	mux.HandleFunc("POST /signout", app.postSignOut)

	mux.HandleFunc("GET /s/{id}", app.getShare)

	mux.Handle("GET /{$}", http.RedirectHandler("/goals", http.StatusSeeOther))
	mux.Handle("GET /goals", app.withAuth(app.getGoals))
	mux.Handle("GET /goals/add/{$}", app.withAuth(app.getAddGoal))
	mux.Handle("POST /goals/add/{$}", app.withAuth(app.postAddGoal))
	mux.Handle("GET /goals/share/{$}", app.withAuth(app.getShareGoals))
	mux.Handle("DELETE /goals/share/{id}", app.withAuth(app.deleteShare))
	mux.Handle("POST /goals/share/create", app.withAuth(app.postCreateShare))
	mux.Handle("GET /goals/{id}", app.withAuth(app.getEditGoal))
	mux.Handle("POST /goals/{id}", app.withAuth(app.postEditGoal))
	mux.Handle("DELETE /goals/{id}", app.withAuth(app.deleteEditGoal))

	mux.Handle("GET /settings", app.withAuth(app.getSettings))
	mux.Handle("POST /settings/branding", app.withAuth(app.postBranding))
	mux.Handle("DELETE /settings/delete-user", app.withAuth(app.deleteUser))

	return app.sessionManager.LoadAndSave(mux)
}
