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
	successCriteria "github.com/bit8bytes/goalkeepr/internal/success_criteria"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

func (app *app) getGoals(w http.ResponseWriter, r *http.Request) {
	goalList, err := app.services.goals.GetAll(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error loading your goals.")
		return
	}

	// TODO: Return goals with success criteria in one criteria
	goalViews := make([]goals.View, len(goalList))
	for i, goal := range goalList {
		goalView := goal.ToView()

		// Fetch success criteria for this goal
		criteria, err := app.services.successCriteria.GetAllByGoal(r.Context(), int(goal.ID), getUserID(r))
		if err != nil && err != sql.ErrNoRows {
			app.renderError(w, r, err, "Error loading success criteria.")
			return
		}

		// Count total and completed criteria
		goalView.TotalCriteriaCount = len(criteria)
		completedCount := 0
		for _, c := range criteria {
			if c.Completed.Valid && c.Completed.Int64 == 1 {
				completedCount++
			}
		}
		goalView.CompletedCriteriaCount = completedCount

		goalViews[i] = goalView
	}

	branding, err := app.services.branding.GetByUserID(r.Context(), getUserID(r))
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading your branding settings.")
		return
	}

	// Calculate default due date for new goals (3 months after latest goal)
	defaultDue := time.Now()
	if len(goalList) > 0 {
		latestGoal := goalList[len(goalList)-1]
		if latestGoal.Due.Valid {
			latestDueTime := time.Unix(latestGoal.Due.Int64, 0)
			defaultDue = latestDueTime.AddDate(0, 3, 0)
		}
	}

	forms := map[string]any{
		"Goals":      goalViews,
		"Branding":   branding.ToView(),
		"Now":        time.Now(),
		"DefaultDue": defaultDue.Format(HTMLDateFormat),
	}

	data := app.newTemplateData(r)
	data.Data = forms
	app.render(w, r, http.StatusOK, page.Goals, data)
}

func (app *app) getAddGoal(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	// Use default_due from query param if provided, otherwise use today
	defaultDue := r.URL.Query().Get("default_due")
	if defaultDue == "" {
		defaultDue = time.Now().Format(HTMLDateFormat)
	}

	data.Form = goals.Form{Due: defaultDue}
	app.render(w, r, http.StatusOK, page.AddGoal, data)
}

func (app *app) postAddGoal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	form := &goals.Form{
		Goal:            sanitize.Text(r.PostForm.Get("goal")),
		Due:             sanitize.Date(r.PostForm.Get("due")),
		VisibleToPublic: r.PostForm.Get("visible") == "on",
	}
	form.Validate()

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.AddGoal, data)
		return
	}

	goalID, err := app.services.goals.Add(r.Context(), getUserID(r), form)
	if err != nil {
		app.renderError(w, r, err, "Error saving your goal.")
		return
	}

	// Redirect to edit page to add success criteria
	http.Redirect(w, r, fmt.Sprintf("/goals/%d", goalID), http.StatusSeeOther)
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

	// Load success criteria for this goal
	criteria, err := app.services.successCriteria.GetAllByGoal(r.Context(), goalID, getUserID(r))
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading success criteria.")
		return
	}

	criteriaViews := make([]successCriteria.View, len(criteria))
	for i, c := range criteria {
		criteriaViews[i] = c.ToView()
	}

	data.Form = editGoalForm
	data.Data = map[string]any{
		"SuccessCriteria": criteriaViews,
		"GoalID":          goalID,
	}
	data.Flash = app.flash(r.Context())
	app.render(w, r, http.StatusOK, page.EditGoal, data)
}

func (app *app) postEditGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawGoal := r.PostForm.Get("goal")
	rawDue := r.PostForm.Get("due")
	visibleToPublic := r.PostForm.Get("visible") == "on"
	achieved := r.PostForm.Get("achieved") == "on"

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

	// TODO: Refactor and use internal/htmx
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

func (app *app) postAddSuccessCriteria(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawDescription := r.PostForm.Get("new-criteria")
	if rawDescription == "" {
		rawDescription = r.PostForm.Get("description")
	}
	rawPosition := r.PostForm.Get("position")

	position := 0
	if rawPosition != "" {
		position, _ = strconv.Atoi(rawPosition)
	}

	form := &successCriteria.Form{
		GoalID:      goalID,
		Description: sanitize.Text(rawDescription),
		Position:    position,
	}

	form.Validate()

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.EditGoal, data)
		return
	}

	if err := app.services.successCriteria.Add(r.Context(), goalID, getUserID(r), form); err != nil {
		app.renderError(w, r, err, "Error saving success criteria.")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/goals/%d", goalID), http.StatusSeeOther)
}

func (app *app) postToggleSuccessCriteria(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	criteriaID, err := strconv.Atoi(r.PathValue("criteriaId"))
	if err != nil {
		app.renderError(w, r, err, "Invalid criteria ID.")
		return
	}

	if _, err := app.services.successCriteria.Toggle(r.Context(), criteriaID, getUserID(r)); err != nil {
		app.renderError(w, r, err, "Error toggling success criteria.")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/goals/%d", goalID), http.StatusSeeOther)
}

func (app *app) deleteSuccessCriteria(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	// Support _method override for DELETE
	method := r.PostForm.Get("_method")
	if method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	criteriaID, err := strconv.Atoi(r.PathValue("criteriaId"))
	if err != nil {
		app.renderError(w, r, err, "Invalid criteria ID.")
		return
	}

	if _, err := app.services.successCriteria.Delete(r.Context(), criteriaID, getUserID(r)); err != nil {
		app.renderError(w, r, err, "Error deleting success criteria.")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/goals/%d", goalID), http.StatusSeeOther)
}

func (app *app) postUpdateSuccessCriteria(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	userID := getUserID(r)

	// Process checkboxes for existing criteria
	criteria, err := app.services.successCriteria.GetAllByGoal(r.Context(), goalID, userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading success criteria.")
		return
	}

	for _, c := range criteria {
		criteriaIDStr := strconv.Itoa(int(c.ID))

		// Check if this criterion should be deleted
		if r.PostForm.Get(fmt.Sprintf("delete_%s", criteriaIDStr)) == "1" {
			if _, err := app.services.successCriteria.Delete(r.Context(), int(c.ID), userID); err != nil {
				app.renderError(w, r, err, "Error deleting success criteria.")
				return
			}
			continue
		}

		// Check if completion status changed
		isChecked := r.PostForm.Get(fmt.Sprintf("criteria_%s", criteriaIDStr)) == "on"
		wasCompleted := c.Completed.Valid && c.Completed.Int64 == 1

		if isChecked != wasCompleted {
			if _, err := app.services.successCriteria.Toggle(r.Context(), int(c.ID), userID); err != nil {
				app.renderError(w, r, err, "Error updating success criteria.")
				return
			}
		}
	}

	// Add new criterion if provided
	newCriterion := sanitize.Text(r.PostForm.Get("new_criterion"))
	if newCriterion != "" {
		form := &successCriteria.Form{
			GoalID:      goalID,
			Description: newCriterion,
			Position:    len(criteria),
		}
		form.Validate()

		if form.Valid() {
			if err := app.services.successCriteria.Add(r.Context(), goalID, userID, form); err != nil {
				app.renderError(w, r, err, "Error saving success criteria.")
				return
			}
		}
	}

	app.putFlash(r.Context(), "Success criteria updated!")
	http.Redirect(w, r, fmt.Sprintf("/goals/%d", goalID), http.StatusSeeOther)
}
