package main

import (
	"database/sql"
	"net/http"

	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"github.com/bit8bytes/toolbox/vcs"
)

func (app *app) getNotFound(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusNotFound, page.NotFound, data)
}

func (app *app) getPrivacy(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, page.Privacy, data)
}

func (app *app) getImprint(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, page.Imprint, data)
}

func (app *app) getLanding(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, page.Landing, data)
}

func (app *app) getShare(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusNotFound, page.NotFound, data)
		return
	}

	userID, err := app.services.share.GetUserIDByPublicID(r.Context(), publicID)
	if err != nil {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusNotFound, page.NotFound, data)
		return
	}

	goalList, err := app.services.goals.GetAllShared(r.Context(), userID)
	if err != nil {
		app.renderError(w, r, err, "Error loading shared goals.")
		return
	}

	goalViews := make([]goals.View, len(goalList))
	for i, goal := range goalList {
		goalViews[i] = goal.ToView()
	}

	b, err := app.services.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading page branding.")
		return
	}

	data := app.newTemplateData(r)
	data.Data = map[string]any{
		"Goals":    goalViews,
		"Branding": b.ToView(),
	}

	app.render(w, r, http.StatusOK, page.Share, data)
}

func (app *app) getHealthz(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"status": "available",
		"system_info": map[string]string{
			"env":     app.config.Env.String(),
			"version": vcs.Version(),
		},
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
