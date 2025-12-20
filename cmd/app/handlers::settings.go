package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/sanitize"
	"github.com/bit8bytes/goalkeepr/internal/users"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

func (app *app) getSettings(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	user, err := app.services.users.GetByID(r.Context(), userID)
	if err != nil {
		app.renderError(w, r, err, "Error loading user settings.")
		return
	}

	b, err := app.services.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading branding settings.")
		return
	}

	brandingForm := branding.Form{}
	if b != nil {
		if b.Title.Valid {
			brandingForm.Title = b.Title.String
		}
		if b.Description.Valid {
			brandingForm.Description = b.Description.String
		}
	}

	forms := map[string]any{
		"Account":  users.UpdateUserForm{Email: user.Email},
		"Branding": brandingForm,
	}

	data := app.newTemplateData(r)
	data.Form = forms
	data.Flash = app.flash(r.Context())
	app.render(w, r, http.StatusOK, page.Settings, data)
}

func (app *app) postBranding(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		app.renderError(w, r, err, "Error processing form data.")
		return
	}

	rawTitle := r.PostForm.Get("title")
	rawDescription := r.PostForm.Get("description")

	form := &branding.Form{
		Title:       sanitize.Text(rawTitle),
		Description: sanitize.Text(rawDescription),
	}

	if !form.Valid() {
		// Todo: better error handling for form but this requires the users email.
		app.renderError(w, r, fmt.Errorf("title or description to long"), "Title or description to long.")
		return
	}

	if err := app.services.branding.CreateOrUpdate(r.Context(), getUserID(r), form.Title, form.Description); err != nil {
		app.renderError(w, r, err, "Error updating branding settings.")
		return
	}

	app.putFlash(r.Context(), "Branding saved")
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

func (app *app) deleteUser(w http.ResponseWriter, r *http.Request) {
	if err := app.services.users.DeleteByID(r.Context(), getUserID(r)); err != nil {
		app.renderError(w, r, err, "Error deleting your account.")
		return
	}

	// Delete session token & client cookie
	app.sessionManager.Destroy(r.Context())
	http.SetCookie(w, &http.Cookie{Name: GoalkeeprCookie, Value: "", Path: "/", MaxAge: -1})

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/signup")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/signup", http.StatusSeeOther)
}
