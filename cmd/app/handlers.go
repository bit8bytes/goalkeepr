package main

import (
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui/layout"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

func (app *app) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, layout.App, page.Home, nil)
}
