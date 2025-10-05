package main

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/bit8bytes/goalkeepr/pkg/assert"
)

func TestPublicPages(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
	}{
		{
			name:     "landing page",
			urlPath:  "/",
			wantCode: http.StatusOK,
		},
		{
			name:     "signup page",
			urlPath:  "/signup",
			wantCode: http.StatusOK,
		},
		{
			name:     "signin page",
			urlPath:  "/signin",
			wantCode: http.StatusOK,
		},
		{
			name:     "not found page",
			urlPath:  "/s/1",
			wantCode: http.StatusNotFound,
		},
		{
			name:     "not found page",
			urlPath:  "/a/b/c",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
		})
	}
}

func TestAuthRequiredPages(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
	}{
		{
			name:     "goals page redirects to signin",
			urlPath:  "/goals",
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "goals add page redirects to signin",
			urlPath:  "/goals/add/",
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "goals share page redirects to signin",
			urlPath:  "/goals/share/",
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "goals detail page redirects to signin",
			urlPath:  "/goals/1",
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "settings page redirects to signin",
			urlPath:  "/settings",
			wantCode: http.StatusSeeOther,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
		})
	}
}

func TestAuthenticatedUserAccess(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	ts.signup(t, "john@doe.com", "12345678", "12345678")

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
	}{
		{
			name:     "landing page accessible",
			urlPath:  "/",
			wantCode: http.StatusOK,
		},
		{
			name:     "signup page redirects",
			urlPath:  "/signup",
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "signin page redirects",
			urlPath:  "/signin",
			wantCode: http.StatusSeeOther,
		},
		{
			name:     "goals page accessible",
			urlPath:  "/goals",
			wantCode: http.StatusOK,
		},
		{
			name:     "settings page accessible",
			urlPath:  "/settings",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, _ := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
		})
	}
}

func TestSignupHandler(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("GET displays signup form", func(t *testing.T) {
		code, _, body := ts.get(t, "/signup")
		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, `<form action="/signup" method="post" novalidate>`)
	})

	t.Run("POST with valid data creates user and redirects", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "test@example.com")
		form.Add("password", "validpassword")
		form.Add("repeat_password", "validpassword")

		code, headers, _ := ts.postForm(t, "/signup", form)
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/goals")
	})

	t.Run("POST with mismatched passwords shows error", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "test2@example.com")
		form.Add("password", "password123")
		form.Add("repeat_password", "different123")

		code, _, body := ts.postForm(t, "/signup", form)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
		assert.StringContains(t, body, "Passwords do not match")
	})

	t.Run("POST with invalid email shows error", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "invalid-email")
		form.Add("password", "validpassword")
		form.Add("repeat_password", "validpassword")

		code, _, body := ts.postForm(t, "/signup", form)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
		assert.StringContains(t, body, "This field must be a valid email address")
	})

	t.Run("POST with short password shows error", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "test3@example.com")
		form.Add("password", "short")
		form.Add("repeat_password", "short")

		code, _, body := ts.postForm(t, "/signup", form)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
		assert.StringContains(t, body, "This field must be at least 8 characters long")
	})
}

func TestSigninHandler(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("GET displays signin form", func(t *testing.T) {
		code, _, body := ts.get(t, "/signin")
		assert.Equal(t, code, http.StatusOK)
		assert.StringContains(t, body, `<form action="/signin" method="post" novalidate>`)
	})

	ts.signup(t, "user@example.com", "testpassword", "testpassword")

	t.Run("POST with valid credentials signs in and redirects", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "user@example.com")
		form.Add("password", "testpassword")

		code, headers, _ := ts.postForm(t, "/signin", form)
		assert.Equal(t, code, http.StatusSeeOther)
		assert.Equal(t, headers.Get("Location"), "/goals")
	})

	t.Run("POST with invalid email shows error", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "nonexistent@example.com")
		form.Add("password", "testpassword")

		code, _, body := ts.postForm(t, "/signin", form)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
		assert.StringContains(t, body, "Invalid email or password")
	})

	t.Run("POST with invalid password shows error", func(t *testing.T) {
		form := url.Values{}
		form.Add("email", "user@example.com")
		form.Add("password", "wrongpassword")

		code, _, body := ts.postForm(t, "/signin", form)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
		assert.StringContains(t, body, "Invalid email or password")
	})

	t.Run("POST with empty fields shows errors", func(t *testing.T) {
		form := url.Values{}

		code, _, body := ts.postForm(t, "/signin", form)
		assert.Equal(t, code, http.StatusUnprocessableEntity)
		assert.StringContains(t, body, "This field cannot be blank")
	})
}
