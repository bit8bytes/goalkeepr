package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/data"
	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/share"
	"github.com/bit8bytes/goalkeepr/internal/users"
)

func newTestApplication(t *testing.T) *app {
	var cfg config

	cfg.db.Driver = "sqlite"
	cfg.db.Path = ":memory:"

	dbP, err := data.New(cfg.db.Driver, cfg.db.Path)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := dbP.AutoMigrate(); err != nil {
		t.Fatal(err)
	}

	// Create an instance of the template cache.
	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = GoalkeeprCookieName

	modules := &modules{
		users:    users.New(dbP.DB),
		goals:    goals.New(dbP.DB),
		branding: branding.New(dbP.DB),
		share:    share.New(dbP.DB),
	}

	return &app{
		config:         cfg,
		logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		templateCache:  templateCache,
		sessionManager: sessionManager,
		modules:        modules,
	}
}

// Define a custom testServer type which embeds a httptest.Server instance.
type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add the cookie jar to the test server client. Any response cookies will
	// now be stored and sent with subsequent requests when using this client.
	ts.Client().Jar = jar

	// Disable redirect-following for the test server client by setting a custom
	// CheckRedirect function. This function will be called whenever a 3xx
	// response is received by the client, and by always returning a
	// http.ErrUseLastResponse error it forces the client to immediately return
	// the received response.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) signup(t *testing.T, email, password, repeatPassword string) {
	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	form.Add("repeat_password", repeatPassword)
	if code, _, _ := ts.postForm(t, "/signup", form); code != http.StatusSeeOther {
		t.Fatalf("login failed: expected redirect, got status %d", code)
	}
}

func (ts *testServer) signin(t *testing.T, email, password string) {
	form := url.Values{}
	form.Add("email", email)
	form.Add("password", password)
	if code, _, _ := ts.postForm(t, "/signin", form); code != http.StatusSeeOther {
		t.Fatalf("login failed: expected redirect, got status %d", code)
	}

}

// Implement a get() method on our custom testServer type. This makes a GET
// request to a given url path using the test server client, and returns the
// response status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}

// Implement a post() method on our custom testServer type for making POST requests.
func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
	if err != nil {
		t.Fatal(err)
	}

	// Read the response body from the test server.
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	// Return the response status, headers and body.
	return rs.StatusCode, rs.Header, string(body)
}
