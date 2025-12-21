package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/sanitize"
	"github.com/bit8bytes/goalkeepr/internal/share"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

func (app *app) getGoals(w http.ResponseWriter, r *http.Request) {
	goalList, err := app.services.goals.GetAll(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error loading your goals.")
		return
	}

	goalViews := make([]goals.View, len(goalList))
	for i, goal := range goalList {
		goalViews[i] = goal.ToView()
	}

	branding, err := app.services.branding.GetByUserID(r.Context(), getUserID(r))
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading your branding settings.")
		return
	}

	forms := map[string]any{
		"Goals":    goalViews,
		"Branding": branding.ToView(),
		"Now":      time.Now(),
	}

	data := app.newTemplateData(r)
	data.Data = forms
	app.render(w, r, http.StatusOK, page.Goals, data)
}

func (app *app) getAddGoal(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = goals.Form{Due: time.Now().Format(HTMLDateFormat)}
	app.render(w, r, http.StatusOK, page.AddGoal, data)
}

func (app *app) postAddGoal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawGoal := r.PostForm.Get("goal")
	rawDue := r.PostForm.Get("due")
	visibleToPublic := r.PostForm.Get("visible") == "on"

	form := &goals.Form{
		Goal:            sanitize.Text(rawGoal),
		Due:             sanitize.Date(rawDue),
		VisibleToPublic: visibleToPublic,
	}

	form.Validate()

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.AddGoal, data)
		return
	}

	if err := app.services.goals.Add(r.Context(), getUserID(r), form); err != nil {
		app.renderError(w, r, err, "Error saving your goal.")
		return
	}

	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

func (app *app) getEditGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	data := app.newTemplateData(r)
	goal, err := app.services.goals.Get(r.Context(), goalID, getUserID(r))
	if err != nil {
		if err == sql.ErrNoRows {
			app.render(w, r, http.StatusNotFound, page.NotFound, data)
			return
		}
		app.renderError(w, r, err, "Couldn't get your goals.")
		return
	}

	goalView := goal.ToView()

	editGoalForm := &goals.Form{
		ID:              int(goalView.ID),
		Goal:            goalView.Goal,
		Due:             goalView.Due.Format(HTMLDateFormat),
		Achieved:        goalView.Achieved,
		VisibleToPublic: goalView.VisibleToPublic,
	}

	data.Form = editGoalForm
	data.Flash = app.flash(r.Context())
	app.render(w, r, http.StatusOK, page.EditGoal, data)
}

func (app *app) postEditGoal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawGoal := r.PostForm.Get("goal")
	rawDue := r.PostForm.Get("due")
	visibleToPublic := r.PostForm.Get("visible") == "on"
	achieved := r.PostForm.Get("achieved") == "on"
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	form := &goals.Form{
		ID:              goalID,
		Goal:            sanitize.Text(rawGoal),
		Due:             sanitize.Date(rawDue),
		VisibleToPublic: visibleToPublic,
		Achieved:        achieved,
	}

	form.Validate()

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.EditGoal, data)
		return
	}

	if _, err = app.services.goals.Update(r.Context(), goalID, getUserID(r), form); err != nil {
		app.renderError(w, r, err, "Error updating your goal.")
		return
	}

	app.putFlash(r.Context(), "Goal saved!")
	http.Redirect(w, r, fmt.Sprintf("/goals/%v", goalID), http.StatusSeeOther)
}

func (app *app) deleteEditGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	rowsAffected, err := app.services.goals.Delete(r.Context(), goalID, getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error deleting your goal.")
		return
	}

	if rowsAffected == 0 {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusNotFound, page.NotFound, data)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/goals")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

func (app *app) getShareGoals(w http.ResponseWriter, r *http.Request) {
	shareLinks, err := app.services.share.GetAll(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error loading your share links.")
		return
	}

	shareViews := make([]share.View, len(shareLinks))
	for i, s := range shareLinks {
		shareViews[i] = s.ToView()
	}

	data := app.newTemplateData(r)
	data.Data = map[string]any{
		"Links": shareViews,
		"Host":  r.Host,
	}

	app.render(w, r, http.StatusOK, page.ShareGoals, data)
}

func (app *app) postCreateShare(w http.ResponseWriter, r *http.Request) {
	shareModel, err := app.services.share.Create(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error creating share link.")
		return
	}

	shareView := shareModel.ToView()

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "shareCreated")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<div class="flex gap-1 items-center">
			<input type="text" value="%s/s/%s" readonly class="input flex-1 bg-base-200" onclick="this.select()">
			<button class="btn" onclick="navigator.clipboard.writeText('%s/s/%s')">Copy</button>
			<button class="btn btn-error" hx-delete="/goals/share/%d" hx-target="closest .flex" hx-swap="outerHTML" hx-confirm="Delete this share link?">Delete</button>
		</div>`, r.Host, shareView.PublicID, r.Host, shareView.PublicID, shareView.ID)
		return
	}

	http.Redirect(w, r, "/goals/share/", http.StatusSeeOther)
}

func (app *app) deleteShare(w http.ResponseWriter, r *http.Request) {
	shareID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid share ID.")
		return
	}

	if err = app.services.share.Delete(r.Context(), shareID); err != nil {
		app.renderError(w, r, err, "Error deleting share link.")
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/goals/share/", http.StatusSeeOther)
}
