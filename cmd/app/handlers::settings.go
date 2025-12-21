package main

import (
	"database/sql"
	"log/slog"
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

	branding, err := app.services.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading branding settings.")
		return
	}

	forms := map[string]any{
		"Account":  users.UpdateUserForm{Email: user.ToView().Email},
		"Branding": branding.ToView(),
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

	form := &branding.Form{
		Title:       sanitize.Text(r.PostForm.Get("title")),
		Description: sanitize.Text(r.PostForm.Get("description")),
	}

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, page.Settings, data)
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
	if err := app.sessionManager.Destroy(r.Context()); err != nil {
		// Continue because the user already has been deleted
		app.logger.WarnContext(r.Context(), "error destroying session", slog.String("msg", err.Error()))
	}
	http.SetCookie(w, &http.Cookie{Name: GoalkeeprCookie, Value: "", Path: "/", MaxAge: -1})

	// TODO: Use internal/htmx helper library
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/signup")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/signup", http.StatusSeeOther)
}
