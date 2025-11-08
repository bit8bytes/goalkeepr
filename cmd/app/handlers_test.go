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

func TestSignInRateLimit(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	for i := 1; i <= 5; i++ {
		code, _, _ := ts.postForm(t, "/signin", url.Values{})
		assert.Equal(t, code, http.StatusUnprocessableEntity)
	}

	code, _, body := ts.postForm(t, "/signin", url.Values{})
	assert.Equal(t, code, http.StatusTooManyRequests)
	assert.StringContains(t, body, "Slow Down There!")
}

func TestSignupRateLimit(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	for i := 1; i <= 5; i++ {
		code, _, _ := ts.postForm(t, "/signup", url.Values{})
		assert.Equal(t, code, http.StatusUnprocessableEntity)
	}

	code, _, body := ts.postForm(t, "/signup", url.Values{})
	assert.Equal(t, code, http.StatusTooManyRequests)
	assert.StringContains(t, body, "Slow Down There!")
}

func BenchmarkRateLimitSignUp(b *testing.B) {
	app := newTestApplication(b)
	ts := newTestServer(b, app.routes())
	defer ts.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ts.postForm(b, "/signup", url.Values{})
		}
	})
}

func BenchmarkRateLimitSignIn(b *testing.B) {
	app := newTestApplication(b)
	ts := newTestServer(b, app.routes())
	defer ts.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ts.postForm(b, "/signin", url.Values{})
		}
	})
}

func FuzzRateLimitGet(f *testing.F) {
	f.Add("192.168.1.1")
	f.Add("::1")
	f.Add("")

	f.Fuzz(func(t *testing.T, ip string) {
		l := newLimiters()

		limiter := l.get(ip)
		if limiter == nil {
			t.Error("limiter should never be nil")
		}
	})
}

func FuzzSignupHandler(f *testing.F) {
	f.Add("test@example.com", "password123", "password123", "")
	f.Add("", "", "", "")
	f.Add("valid@email.com", "short", "short", "")
	f.Add("invalid-email", "validpassword", "validpassword", "")
	f.Add("test@example.com", "password123", "different", "")
	f.Add("user@domain.co.uk", "LongPassword123!", "LongPassword123!", "")
	f.Add("test@example.com", "", "password123", "")
	f.Add("test@example.com", "password123", "", "")
	f.Add("a@b.c", "p", "p", "")

	f.Fuzz(func(t *testing.T, email, password, repeatPassword, website string) {
		app := newTestApplication(t)
		ts := newTestServer(t, app.routes())
		defer ts.Close()

		form := url.Values{}
		form.Add("email", email)
		form.Add("password", password)
		form.Add("repeat_password", repeatPassword)
		form.Add("website", website)

		code, _, body := ts.postForm(t, "/signup", form)

		if code < 200 || code >= 600 {
			t.Errorf("unexpected status code: %d", code)
		}

		if code == http.StatusOK && len(body) == 0 {
			t.Error("empty response body for 200 status")
		}
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
