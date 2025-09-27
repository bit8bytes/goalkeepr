package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/sanitize"
	"github.com/bit8bytes/goalkeepr/internal/users"
	"github.com/bit8bytes/goalkeepr/ui/layout"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"github.com/bit8bytes/toolbox/validator"
)

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
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	user := &users.User{
		Email: form.Email,
	}

	user, err := app.modules.users.Get(r.Context(), user)
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

func (a *app) getGoals(w http.ResponseWriter, r *http.Request) {
	data := newTemplateData(r)
	a.render(w, r, http.StatusOK, layout.App, page.Goals, data)
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
		UserID:          1,
		Goal:            form.Goal,
		Due:             dueTime,
		VisibleToPublic: form.VisibleToPublic,
		Achieved:        false,
	}

	if err := a.modules.goals.Add(r.Context(), goal); err != nil {
		a.logger.Error("failed to add goal", slog.String("err", err.Error()))
		data := newTemplateData(r)
		data.Form = form
		a.render(w, r, http.StatusUnprocessableEntity, layout.App, page.AddGoal, data)
		return
	}

	http.Redirect(w, r, "/goals", http.StatusSeeOther)

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
