package main

import (
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui"
)

func (app *app) routes() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/static/", ui.ServeStaticFiles())

	mux.HandleFunc("/", app.home)

	return mux
}
