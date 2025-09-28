package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/sanitize"
	"github.com/bit8bytes/goalkeepr/internal/users"
	"github.com/bit8bytes/goalkeepr/ui/layout"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"github.com/bit8bytes/toolbox/validator"
)

func (app *app) getNotFound(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	app.render(w, r, http.StatusNotFound, layout.Center, page.NotFound, data)
}

func (app *app) getLanding(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	app.render(w, r, http.StatusOK, layout.Landing, page.Landing, data)
}

type signUpForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	RepeatPassword      string `form:"repeat_password"`
	validator.Validator `form:"-"`
}

func (app *app) getSignUp(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	data.Form = &signUpForm{}
	app.render(w, r, http.StatusOK, layout.Auth, page.SignUp, data)
}

func (app *app) postSignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error parsing form.")
		return
	}

	rawEmail := r.PostForm.Get("email")
	rawPassword := r.PostForm.Get("password")
	rawRepeatPassword := r.PostForm.Get("repeat_password")
	rawWebsite := r.PostForm.Get("website")

	// Honeypot for bot protection
	if sanitize.Text(rawWebsite) != "" {
		time.Sleep(3 * time.Second)
		return
	}

	form := &signUpForm{
		Email:          sanitize.Email(rawEmail),
		Password:       sanitize.Password(rawPassword),
		RepeatPassword: sanitize.Password(rawRepeatPassword),
	}

	validateEmail(form, form.Email)
	validatePassword(form, form.Password)
	validateRepeatPassword(form, form.Password, form.RepeatPassword)

	if !form.Valid() {
		data := newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.Auth, page.SignUp, data)
		return
	}

	user := &users.User{Email: form.Email}

	if err := user.Password.Set(form.Password); err != nil {
		app.renderError(w, r, err, "Error setting up your account.")
		return
	}

	userID, err := app.modules.users.Add(r.Context(), user)
	if err != nil {
		data := newTemplateData(r)
		form.AddError("email", "This email cannot be used.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.Auth, page.SignUp, data)
		return
	}

	app.sessionManager.Put(r.Context(), UserIDSessionKey, userID)
	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

type signInForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	Website             string `form:"website"`
	validator.Validator `form:"-"`
}

func (app *app) getSignIn(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	data.Form = new(signInForm)
	app.render(w, r, http.StatusOK, layout.Auth, page.SignIn, data)
}

func (app *app) postSignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawEmail := r.PostForm.Get("email")
	rawPassword := r.PostForm.Get("password")
	rawWebsite := r.PostForm.Get("website")

	// Honeypot for bot protection
	if sanitize.Text(rawWebsite) != "" {
		time.Sleep(3 * time.Second)
		return
	}

	form := &signInForm{
		Email:    sanitize.Email(rawEmail),
		Password: sanitize.Password(rawPassword),
	}

	validateEmail(form, form.Email)
	validatePassword(form, form.Password)

	if !form.Valid() {
		data := newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.Auth, page.SignIn, data)
		return
	}

	user, err := app.modules.users.GetByEmail(r.Context(), form.Email)
	if err != nil {
		data := newTemplateData(r)
		form.AddError("email", "Invalid email or password.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.Auth, page.SignIn, data)
		return
	}

	match, err := user.Password.Matches(form.Password)
	if err != nil {
		app.renderError(w, r, err, "Error validating password.")
		return
	}

	if !match {
		data := newTemplateData(r)
		form := signInForm{Email: form.Email} // Return only the email
		form.AddError("email", "Invalid email or password.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.Auth, page.SignIn, data)
		return
	}

	app.sessionManager.Put(r.Context(), UserIDSessionKey, user.ID)
	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

func (app *app) postSignOut(w http.ResponseWriter, r *http.Request) {
	// Remove server session & client side cookie
	app.sessionManager.Remove(r.Context(), UserIDSessionKey)
	http.SetCookie(w, &http.Cookie{Name: GoalkeeprCookieName, Value: "", Path: "/", MaxAge: -1})

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *app) getGoals(w http.ResponseWriter, r *http.Request) {
	goals, err := app.modules.goals.GetAll(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error loading your goals.")
		return
	}

	branding, err := app.modules.branding.GetByUserID(r.Context(), getUserID(r))
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading your branding settings.")
		return
	}

	forms := map[string]any{
		"Goals":    goals,
		"Branding": branding,
	}

	data := newTemplateData(r)
	data.Data = forms
	app.render(w, r, http.StatusOK, layout.App, page.Goals, data)
}

type goalForm struct {
	ID                  int    `form:"id"`
	Goal                string `form:"goal"`
	Due                 string `form:"due"`
	Achieved            bool   `form:"achieved"`
	VisibleToPublic     bool   `form:"visible"`
	validator.Validator `form:"-"`
}

func (app *app) getAddGoal(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	data.Form = goalForm{Due: time.Now().Format(HTMLDateFormat)}
	app.render(w, r, http.StatusOK, layout.App, page.AddGoal, data)
}

func (app *app) postAddGoal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawGoal := r.PostForm.Get("goal")
	rawDue := r.PostForm.Get("due")
	visibleToPublic := r.PostForm.Get("visible") == "on"

	form := &goalForm{
		Goal:            sanitize.Text(rawGoal),
		Due:             sanitize.Date(rawDue),
		VisibleToPublic: visibleToPublic,
	}

	validateAddGoal(form)

	if !form.Valid() {
		data := newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.App, page.AddGoal, data)
		return
	}

	dueTime, err := time.Parse("2006-01-02", form.Due)
	if err != nil {
		app.renderError(w, r, err, "Invalid date format.")
		return
	}

	goal := &goals.Goal{
		UserID:          getUserID(r),
		Goal:            form.Goal,
		Due:             dueTime,
		VisibleToPublic: form.VisibleToPublic,
		Achieved:        false,
	}

	if err := app.modules.goals.Add(r.Context(), goal); err != nil {
		app.renderError(w, r, err, "Error saving your goal.")
		return
	}

	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

// type goalForm struct {
// 	ID                  int    `form:"-"`
// 	Goal                string `form:"goal"`
// 	Due                 string `form:"due"`

// 	VisibleToPublic     bool   `form:"visible"`
// 	validator.Validator `form:"-"`
// }

func (app *app) getEditGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	data := newTemplateData(r)
	goal, err := app.modules.goals.Get(r.Context(), goalID, getUserID(r))
	if err != nil {
		if err == sql.ErrNoRows {
			app.render(w, r, http.StatusNotFound, layout.Center, page.NotFound, data)
			return
		}
		app.renderError(w, r, err, "Couldn't get your goals.")
		return
	}

	editGoalForm := &goalForm{
		ID:              goal.ID,
		Goal:            goal.Goal,
		Due:             goal.Due.Format(HTMLDateFormat),
		Achieved:        goal.Achieved,
		VisibleToPublic: goal.VisibleToPublic,
	}

	data.Form = editGoalForm
	app.render(w, r, http.StatusOK, layout.App, page.EditGoal, data)
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

	form := &goalForm{
		ID:              goalID,
		Goal:            sanitize.Text(rawGoal),
		Due:             sanitize.Date(rawDue),
		VisibleToPublic: visibleToPublic,
		Achieved:        achieved,
	}

	validateEditGoal(form)

	if !form.Valid() {
		data := newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.App, page.EditGoal, data)
		return
	}

	dueTime, err := time.Parse("2006-01-02", form.Due)
	if err != nil {
		app.renderError(w, r, err, "Invalid date format.")
		return
	}

	goal := &goals.Goal{
		ID:              goalID,
		UserID:          getUserID(r),
		Goal:            form.Goal,
		Due:             dueTime,
		VisibleToPublic: form.VisibleToPublic,
		Achieved:        form.Achieved,
	}

	if _, err = app.modules.goals.Update(r.Context(), goal); err != nil {
		app.renderError(w, r, err, "Error updating your goal.")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/goals/%v", goalID), http.StatusSeeOther)
}

func (app *app) deleteEditGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.renderError(w, r, err, "Invalid goal ID.")
		return
	}

	rowsAffected, err := app.modules.goals.Delete(r.Context(), goalID, getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error deleting your goal.")
		return
	}

	if rowsAffected == 0 {
		data := newTemplateData(r)
		app.render(w, r, http.StatusNotFound, layout.Center, page.NotFound, data)
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
	shareLinks, err := app.modules.share.GetAll(r.Context(), getUserID(r))
	if err != nil {
		app.renderError(w, r, err, "Error loading your share links.")
		return
	}

	data := newTemplateData(r)
	data.Data = map[string]any{
		"Links": shareLinks,
		"Host":  r.Host,
	}

	app.render(w, r, http.StatusOK, layout.App, page.ShareGoals, data)
}

func (app *app) postCreateShare(w http.ResponseWriter, r *http.Request) {
	share, err := app.modules.share.Create(r.Context(), getUserID(r))
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

	if err = app.modules.share.Delete(r.Context(), shareID); err != nil {
		app.renderError(w, r, err, "Error deleting share link.")
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/goals/share/", http.StatusSeeOther)
}

type accountForm struct {
	Email               string `form:"email"`
	validator.Validator `form:"-"`
}

type brandingForm struct {
	Title               string `form:"title"`
	Description         string `form:"description"`
	validator.Validator `form:"-"`
}

func (app *app) getSettings(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	user, err := app.modules.users.GetByID(r.Context(), userID)
	if err != nil {
		app.renderError(w, r, err, "Error loading user settings.")
		return
	}

	branding, err := app.modules.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading branding settings.")
		return
	}

	forms := map[string]any{
		"Account":  accountForm{Email: user.Email},
		"Branding": branding,
	}

	data := newTemplateData(r)
	data.Form = forms
	app.render(w, r, http.StatusOK, layout.Settings, page.Settings, data)
}

func (app *app) postBranding(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawTitle := r.PostForm.Get("title")
	rawDescription := r.PostForm.Get("description")

	form := &brandingForm{
		Title:       sanitize.Text(rawTitle),
		Description: sanitize.Text(rawDescription),
	}

	validateBranding(form)

	if !form.Valid() {
		// Todo: better error handling for form but this requires the users email.
		app.renderError(w, r, fmt.Errorf("title or description to long"), "Title or description to long.")
		return
	}

	branding := &branding.Branding{
		UserID:      getUserID(r),
		Title:       form.Title,
		Description: form.Description,
	}

	if err := app.modules.branding.CreateOrUpdate(r.Context(), branding); err != nil {
		app.renderError(w, r, err, "Error updating branding settings.")
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (app *app) deleteUser(w http.ResponseWriter, r *http.Request) {
	if err := app.modules.users.DeleteByID(r.Context(), getUserID(r)); err != nil {
		app.renderError(w, r, err, "Error deleting your account.")
		return
	}

	// Delete session token & client cookie
	app.sessionManager.Remove(r.Context(), UserIDSessionKey)
	http.SetCookie(w, &http.Cookie{Name: GoalkeeprCookieName, Value: "", Path: "/", MaxAge: -1})

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/signup")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/signup", http.StatusSeeOther)
}

func (app *app) getShare(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		data := newTemplateData(r)
		app.render(w, r, http.StatusNotFound, layout.Center, page.NotFound, data)
		return
	}

	userID, err := app.modules.share.GetUserIDByPublicID(r.Context(), publicID)
	if err != nil {
		data := newTemplateData(r)
		app.render(w, r, http.StatusNotFound, layout.Center, page.NotFound, data)
		return
	}

	goals, err := app.modules.goals.GetAllShared(r.Context(), userID)
	if err != nil {
		app.renderError(w, r, err, "Error loading shared goals.")
		return
	}

	branding, err := app.modules.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading page branding.")
		return
	}

	data := newTemplateData(r)
	data.Data = map[string]any{
		"Goals":    goals,
		"Branding": branding,
	}

	app.render(w, r, http.StatusOK, layout.Share, page.Share, data)
}
