package main

import (
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui/layout"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

func (a *app) getGoals(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	a.render(w, r, http.StatusOK, layout.App, page.Goals, data)
}

func (a *app) getAddGoal(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	a.render(w, r, http.StatusOK, layout.App, page.AddGoal, data)
}

func (a *app) getEditGoal(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	a.render(w, r, http.StatusOK, layout.App, page.EditGoal, data)
}

func (a *app) getShareGoals(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	a.render(w, r, http.StatusOK, layout.App, page.ShareGoals, data)
}

func (a *app) getSettings(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	a.render(w, r, http.StatusOK, layout.Settings, page.Settings, data)
}

func (a *app) getShare(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	a.render(w, r, http.StatusOK, layout.Share, page.Share, data)
}
