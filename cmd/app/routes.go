package main

import (
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui"
)

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", ui.ServeStaticFiles())

	mux.HandleFunc("/signup", app.getSignUp)
	mux.HandleFunc("POST /signup", app.postSignUp)
	mux.HandleFunc("/signin", app.getSignIn)
	mux.HandleFunc("POST /signin", app.postSignIn)
	mux.HandleFunc("POST /signout", app.postSignOut)

	mux.HandleFunc("/s/{id}", app.getShare)

	mux.Handle("/{$}", http.RedirectHandler("/goals", http.StatusSeeOther))
	mux.Handle("/goals", app.withAuth(app.getGoals))
	mux.Handle("/goals/add", app.withAuth(app.getAddGoal))
	mux.Handle("POST /goals/add", app.withAuth(app.postAddGoal))
	mux.Handle("/goals/{id}", app.withAuth(app.getEditGoal))
	mux.Handle("/goals/share", app.withAuth(app.getShareGoals))

	mux.Handle("/settings", app.withAuth(app.getSettings))

	return app.sessionManager.LoadAndSave(mux)
}
