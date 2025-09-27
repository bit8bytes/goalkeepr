package main

import (
	"database/sql"
	"fmt"
	"log/slog"
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

type signUpForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	RepeatPassword      string `form:"repeat_password"`
	validator.Validator `form:"-"`
}

func (a *app) getSignUp(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	data.Form = new(signUpForm)
	a.render(w, r, http.StatusOK, layout.Auth, page.SignUp, data)
}

func (app *app) postSignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	rawEmail := r.PostForm.Get("email")
	rawPassword := r.PostForm.Get("password")
	rawRepeatPassword := r.PostForm.Get("repeat_password")

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

	user := &users.User{
		Email: form.Email,
	}

	if err := user.Password.Set(form.Password); err != nil {
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
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
	validator.Validator `form:"-"`
}

func (a *app) getSignIn(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	data.Form = new(signInForm)
	a.render(w, r, http.StatusOK, layout.Auth, page.SignIn, data)
}

func (app *app) postSignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	rawEmail := r.PostForm.Get("email")
	rawPassword := r.PostForm.Get("password")

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
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	if !match {
		app.logger.Warn("Password didn't match")
		data := newTemplateData(r)
		form := signInForm{Email: form.Email}
		form.AddError("email", "Invalid email or password.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.Auth, page.SignIn, data)
		return
	}

	app.sessionManager.Put(r.Context(), UserIDSessionKey, user.ID)
	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

func (app *app) postSignOut(w http.ResponseWriter, r *http.Request) {
	app.sessionManager.Remove(r.Context(), UserIDSessionKey)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/signin")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func (app *app) getGoals(w http.ResponseWriter, r *http.Request) {
	g, err := app.modules.goals.GetAll(r.Context(), getUserID(r))
	if err != nil {
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	b, err := app.modules.branding.GetByUserID(r.Context(), getUserID(r))
	if err != nil && err != sql.ErrNoRows {
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	brd := &branding.Branding{}
	if b != nil {
		brd = b
	}

	group := struct {
		Goals    []goals.Goal
		Branding branding.Branding
	}{
		Goals:    g,
		Branding: *brd,
	}

	data := newTemplateData(r)
	data.Data = group
	app.render(w, r, http.StatusOK, layout.App, page.Goals, data)
}

type addGoalForm struct {
	Goal                string `form:"goal"`
	Due                 string `form:"due"`
	VisibleToPublic     bool   `form:"visible"`
	validator.Validator `form:"-"`
}

func (a *app) getAddGoal(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	data.Form = addGoalForm{Due: time.Now().Format(HTMLDateFormat)}
	a.render(w, r, http.StatusOK, layout.App, page.AddGoal, data)
}

func (a *app) postAddGoal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	rawGoal := r.PostForm.Get("goal")
	rawDue := r.PostForm.Get("due")
	visibleToPublic := r.PostForm.Get("visible") == "on"

	form := addGoalForm{
		Goal:            sanitize.Text(rawGoal),
		Due:             sanitize.Date(rawDue),
		VisibleToPublic: visibleToPublic,
	}

	validateAddGoal(&form)

	if !form.Valid() {
		data := newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, layout.App, page.AddGoal, data)
		return
	}

	dueTime, err := time.Parse("2006-01-02", form.Due)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	goal := &goals.Goal{
		UserID:          getUserID(r),
		Goal:            form.Goal,
		Due:             dueTime,
		VisibleToPublic: form.VisibleToPublic,
		Achieved:        false,
	}

	if err := a.modules.goals.Add(r.Context(), goal); err != nil {
		a.logger.Error("failed to add goal", slog.String("err", err.Error()))
		data := newTemplateData(r)
		a.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

type editGoalForm struct {
	Goal                string `form:"goal"`
	Due                 string `form:"due"`
	Achieved            bool   `form:"achieved"`
	VisibleToPublic     bool   `form:"visible"`
	validator.Validator `form:"-"`
}

func (app *app) getEditGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	goal, err := app.modules.goals.Get(r.Context(), goalID, getUserID(r))
	if err != nil {
		if err == sql.ErrNoRows {
			data := newTemplateData(r)
			app.render(w, r, http.StatusNotFound, layout.Center, page.NotFound, data)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := newTemplateData(r)
	data.Form = new(editGoalForm)
	data.Data = goal
	app.render(w, r, http.StatusOK, layout.App, page.EditGoal, data)
}

func (app *app) postEditGoal(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	rawGoal := r.PostForm.Get("goal")
	rawDue := r.PostForm.Get("due")
	visibleToPublic := r.PostForm.Get("visible") == "on"
	achieved := r.PostForm.Get("achieved") == "on"

	form := editGoalForm{
		Goal:            sanitize.Text(rawGoal),
		Due:             sanitize.Date(rawDue),
		VisibleToPublic: visibleToPublic,
		Achieved:        achieved,
	}

	validateEditGoal(&form)

	if !form.Valid() {
		data := newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, layout.App, page.AddGoal, data)
		return
	}

	dueTime, err := time.Parse("2006-01-02", form.Due)
	if err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
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
		app.logger.Error("failed to update goal", slog.String("err", err.Error()))
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/goals/%v", goalID), http.StatusSeeOther)
}

func (app *app) deleteEditGoal(w http.ResponseWriter, r *http.Request) {
	goalID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rowsAffected, err := app.modules.goals.Delete(r.Context(), goalID, getUserID(r))
	if err != nil {
		app.logger.Error("failed to delete goal", slog.String("err", err.Error()))
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
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
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
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
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
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
	id := r.PathValue("id")
	shareID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid share ID", http.StatusBadRequest)
		return
	}

	err = app.modules.share.Delete(r.Context(), shareID)
	if err != nil {
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/goals/share/", http.StatusSeeOther)
}

type editSettingsForm struct {
	Account  editAccountForm
	Branding editBrandingForm
}

type editAccountForm struct {
	Email               string `form:"email"`
	validator.Validator `form:"-"`
}

type editBrandingForm struct {
	Title               string `form:"title"`
	Description         string `form:"description"`
	validator.Validator `form:"-"`
}

func (app *app) getSettings(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	if userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := app.modules.users.GetByID(ctx, userID)
	if err != nil {
		app.logger.Error("failed to get user", slog.String("err", err.Error()), slog.Int("userID", userID))
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	branding, err := app.modules.branding.GetByUserID(ctx, userID)
	if err != nil && err != sql.ErrNoRows {
		app.logger.Error("failed to get branding", slog.String("err", err.Error()), slog.Int("userID", userID))
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	var brandingForm editBrandingForm
	if branding != nil {
		brandingForm.Title = branding.Title
		brandingForm.Description = branding.Description
	}

	form := editSettingsForm{
		Account:  editAccountForm{Email: user.Email},
		Branding: brandingForm,
	}

	data := newTemplateData(r)
	data.Data = form
	app.render(w, r, http.StatusOK, layout.Settings, page.Settings, data)
}

func (app *app) postBranding(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}

	rawTitle := r.PostForm.Get("title")
	rawDescription := r.PostForm.Get("description")

	form := editBrandingForm{
		Title:       sanitize.Text(rawTitle),
		Description: sanitize.Text(rawDescription),
	}

	validateEditBranding(&form)

	if !form.Valid() {
		data := newTemplateData(r)
		data.Form = form
		data.Data = editSettingsForm{}
		app.render(w, r, http.StatusUnprocessableEntity, layout.Settings, page.Settings, data)
		return
	}

	brandingData := &branding.Branding{
		UserID:      getUserID(r),
		Title:       form.Title,
		Description: form.Description,
	}

	if err := app.modules.branding.CreateOrUpdate(r.Context(), brandingData); err != nil {
		app.logger.Error("failed to update branding", slog.String("err", err.Error()))
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (app *app) deleteUser(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	err := app.modules.users.DeleteByID(r.Context(), userID)
	if err != nil {
		app.logger.Error("failed to delete user", slog.String("err", err.Error()))
		data := newTemplateData(r)
		app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	app.sessionManager.Remove(r.Context(), UserIDSessionKey)

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/signup")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/signup", http.StatusSeeOther)
}

func (a *app) getShare(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		http.Error(w, "Share ID required", http.StatusBadRequest)
		return
	}

	userID, err := a.modules.share.GetUserIDByPublicID(r.Context(), publicID)
	if err != nil {
		data := newTemplateData(r)
		a.render(w, r, http.StatusNotFound, layout.Center, page.NotFound, data)
		return
	}

	goals, err := a.modules.goals.GetAllShared(r.Context(), userID)
	if err != nil {
		data := newTemplateData(r)
		a.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	branding, err := a.modules.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		data := newTemplateData(r)
		a.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
		return
	}

	data := newTemplateData(r)
	data.Data = map[string]any{
		"Goals":    goals,
		"Branding": branding,
	}
	a.render(w, r, http.StatusOK, layout.Share, page.Share, data)
}
