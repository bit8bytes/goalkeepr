package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/sanitize"
	"github.com/bit8bytes/goalkeepr/internal/users"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

func (app *app) getSignUp(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), string(users.Key))
	if userID != 0 {
		http.Redirect(w, r, "/goals", http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.Form = &users.SignUpForm{}
	app.render(w, r, http.StatusOK, page.SignUp, data)
}

func (app *app) postSignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.logger.WarnContext(r.Context(), "error parsing form", slog.String("msg", err.Error()))
		data := app.newTemplateData(r)
		form := &users.SignUpForm{} // Needs to initialized. The other returns already have it.
		form.AddError("email", "This email cannot be used.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.SignUp, data)
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

	form := &users.SignUpForm{
		Email:          sanitize.Email(rawEmail),
		Password:       sanitize.Password(rawPassword),
		RepeatPassword: sanitize.Password(rawRepeatPassword),
	}

	form.Validate()

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.SignUp, data)
		return
	}

	user := &users.User{Email: form.Email}

	if err := user.SetPassword(form.Password); err != nil {
		app.logger.WarnContext(r.Context(), "error setting user password", slog.String("msg", err.Error()))
		data := app.newTemplateData(r)
		form.AddError("email", "This email cannot be used.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.SignUp, data)
		return
	}

	userID, err := app.services.users.Add(r.Context(), user)
	if err != nil {
		app.logger.WarnContext(r.Context(), "error creating new user", slog.String("msg", err.Error()))
		data := app.newTemplateData(r)
		form.AddError("email", "This email cannot be used.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.SignUp, data)
		return
	}

	app.sessionManager.Put(r.Context(), string(users.Key), userID)
	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

func (app *app) getSignIn(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetInt(r.Context(), string(users.Key))
	if userID != 0 {
		http.Redirect(w, r, "/goals", http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.Form = new(users.SignInForm)
	app.render(w, r, http.StatusOK, page.SignIn, data)
}

func (app *app) postSignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.logger.WarnContext(r.Context(), "error parsing form", slog.String("msg", err.Error()))
		data := app.newTemplateData(r)
		form := &users.SignInForm{} // Needs to initialized. The other returns already have it.
		form.AddError("email", "Invalid email or password.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.SignIn, data)
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

	form := &users.SignInForm{
		Email:    sanitize.Email(rawEmail),
		Password: sanitize.Password(rawPassword),
	}

	form.Validate()

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.SignIn, data)
		return
	}

	var user *users.User
	var match bool

	user, err := app.services.users.GetByEmail(r.Context(), form.Email)
	if err != nil {
		// Prevent timing attacks - use pre-computed bcrypt hash
		dummyUser := &users.User{PasswordHash: users.PreComputedHash}
		_, _ = dummyUser.MatchesPassword(form.Password) // Intentionally ignored to prevent timing attacks
		match = false

		app.logger.WarnContext(r.Context(), "error getting user by email", slog.String("msg", err.Error()))
	} else {
		match, err = user.MatchesPassword(form.Password)
		if err != nil {
			app.logger.WarnContext(r.Context(), "error matching passwords", slog.String("msg", err.Error()))
			match = false
		}
	}

	if !match {
		app.logger.WarnContext(r.Context(), "passwords doesn't match")
		data := app.newTemplateData(r)
		form := users.SignInForm{Email: form.Email} // Only email
		form.AddError("email", "Invalid email or password.")
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.SignIn, data)
		return
	}

	app.sessionManager.Put(r.Context(), string(users.Key), int(user.ID))
	http.Redirect(w, r, "/goals", http.StatusSeeOther)
}

func (app *app) postSignOut(w http.ResponseWriter, r *http.Request) {
	// Remove server session & client side cookie
	if err := app.sessionManager.Destroy(r.Context()); err != nil {
		app.logger.WarnContext(r.Context(), "error destroying session", slog.String("msg", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: GoalkeeprCookie, Value: "", Path: "/", MaxAge: -1})

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
