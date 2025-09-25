package main

import (
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui"
)

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", ui.ServeStaticFiles())

	mux.Handle("/", http.RedirectHandler("/goals", http.StatusSeeOther))
	mux.HandleFunc("/goals", app.getGoals)
	mux.HandleFunc("/goals/add", app.getAddGoal)
	mux.HandleFunc("/goals/{id}", app.getEditGoal)
	mux.HandleFunc("/goals/share", app.getShareGoals)

	mux.HandleFunc("/settings", app.getSettings)

	return mux
}
