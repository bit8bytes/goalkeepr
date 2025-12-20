package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/sanitize"
	"github.com/bit8bytes/goalkeepr/internal/users"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"github.com/bit8bytes/toolbox/vcs"
)

func (app *app) getNotFound(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusNotFound, page.NotFound, data)
}

func (app *app) getPrivacy(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, page.Privacy, data)
}

func (app *app) getImprint(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, page.Imprint, data)
}

func (app *app) getLanding(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, r, http.StatusOK, page.Landing, data)
}

type flash struct {
	Content string
}

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

	validateBranding(form)

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

func (app *app) getShare(w http.ResponseWriter, r *http.Request) {
	publicID := r.PathValue("id")
	if publicID == "" {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusNotFound, page.NotFound, data)
		return
	}

	userID, err := app.services.share.GetUserIDByPublicID(r.Context(), publicID)
	if err != nil {
		data := app.newTemplateData(r)
		app.render(w, r, http.StatusNotFound, page.NotFound, data)
		return
	}

	goals, err := app.services.goals.GetAllShared(r.Context(), userID)
	if err != nil {
		app.renderError(w, r, err, "Error loading shared goals.")
		return
	}

	b, err := app.services.branding.GetByUserID(r.Context(), userID)
	if err != nil && err != sql.ErrNoRows {
		app.renderError(w, r, err, "Error loading page branding.")
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

	data := app.newTemplateData(r)
	data.Data = map[string]any{
		"Goals":    goals,
		"Branding": brandingData,
	}

	app.render(w, r, http.StatusOK, page.Share, data)
}

func (app *app) getHealthz(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"status": "available",
		"system_info": map[string]string{
			"env":     app.config.Env.String(),
			"version": vcs.Version(),
		},
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
