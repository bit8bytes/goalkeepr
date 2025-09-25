package main

import (
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui"
)

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", ui.ServeStaticFiles())

	mux.HandleFunc("/", app.getLanding)

	mux.Handle("/signin", http.RedirectHandler("http://localhost:8080/signin", http.StatusSeeOther))
	mux.Handle("/signup", http.RedirectHandler("http://localhost:8080/signup", http.StatusSeeOther))

	return mux
}
