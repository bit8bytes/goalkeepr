package main

import (
	"database/sql"
	"net/http"
	"time"

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

	// Group goals by date for visual grouping in timeline
	goalGroups := []GoalGroup{}
	var currentGroup *GoalGroup

	for _, goalView := range goalViews {
		// Normalize the date to start of day for comparison
		goalDate := time.Date(goalView.Due.Year(), goalView.Due.Month(), goalView.Due.Day(), 0, 0, 0, 0, time.UTC)

		// Check if we need to start a new group
		if currentGroup == nil || !currentGroup.Date.Equal(goalDate) {
			goalGroups = append(goalGroups, GoalGroup{
				Date:  goalDate,
				Goals: []goals.View{goalView},
			})
			currentGroup = &goalGroups[len(goalGroups)-1]
		} else {
			// Add to current group
			currentGroup.Goals = append(currentGroup.Goals, goalView)
		}
	}

	b, err := app.services.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading page branding.")
		return
	}

	data := app.newTemplateData(r)
	data.Data = SharePageData{
		Goals:      goalViews,
		GoalGroups: goalGroups,
		Branding:   b.ToView(),
	}

	app.render(w, r, http.StatusOK, page.Share, data)
}

type getHealthzData struct {
	Status string `json:"status"`
	System system `json:"system"`
}

type system struct {
	Env     string `json:"env"`
	Version string `json:"version"`
}

func (app *app) getHealthz(w http.ResponseWriter, r *http.Request) {
	data := getHealthzData{
		Status: "available",
		System: system{
			Env:     app.config.Env.String(),
			Version: vcs.Version(),
		},
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
