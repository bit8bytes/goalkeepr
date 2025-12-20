package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/sanitize"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

// GoalView is a template-friendly representation of a Goal
type GoalView struct {
	ID              int64
	UserID          int64
	Goal            string
	Due             time.Time
	VisibleToPublic bool
	Achieved        bool
}

func toGoalViews(sqlcGoals []goals.Goal) []GoalView {
	views := make([]GoalView, 0, len(sqlcGoals))
	for _, g := range sqlcGoals {
		view := GoalView{
			ID:     g.ID,
			UserID: g.UserID,
		}
		if g.Goal.Valid {
			view.Goal = g.Goal.String
		}
		if g.Due.Valid {
			view.Due = time.Unix(g.Due.Int64, 0)
		}
		if g.VisibleToPublic.Valid {
			view.VisibleToPublic = g.VisibleToPublic.Int64 == 1
		}
		if g.Achieved.Valid {
			view.Achieved = g.Achieved.Int64 == 1
		}
		views = append(views, view)
	}
	return views
}

func (app *app) getGoals(w http.ResponseWriter, r *http.Request) {
	goals, err := app.services.goals.GetAll(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error loading your goals.")
		return
	}

	b, err := app.services.branding.GetByUserID(r.Context(), getUserID(r))
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading your branding settings.")
		return
	}

	brandingData := map[string]string{
		"Title":       "",
		"Description": "",
	}
	if b != nil {
		if b.Title.Valid {
			brandingData["Title"] = b.Title.String
		}
		if b.Description.Valid {
			brandingData["Description"] = b.Description.String
		}
	}

	forms := map[string]any{
		"Goals":    goals,
		"Branding": brandingData,
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

	editGoalForm := &goals.Form{
		ID:              int(goal.ID),
		Goal:            goal.Goal,
		Due:             goal.Due.Format(HTMLDateFormat),
		Achieved:        goal.Achieved,
		VisibleToPublic: goal.VisibleToPublic,
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

	data := app.newTemplateData(r)
	data.Data = map[string]any{
		"Links": shareLinks,
		"Host":  r.Host,
	}

	app.render(w, r, http.StatusOK, page.ShareGoals, data)
}

func (app *app) postCreateShare(w http.ResponseWriter, r *http.Request) {
	share, err := app.services.share.Create(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error creating share link.")
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "shareCreated")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<div class="flex gap-1 items-center">
			<input type="text" value="%s/s/%s" readonly class="input flex-1 bg-base-200" onclick="this.select()">
			<button class="btn" onclick="navigator.clipboard.writeText('%s/s/%s')">Copy</button>
			<button class="btn btn-error" hx-delete="/goals/share/%d" hx-target="closest .flex" hx-swap="outerHTML" hx-confirm="Delete this share link?">Delete</button>
		</div>`, r.Host, share.PublicID, r.Host, share.PublicID, share.ID)
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
