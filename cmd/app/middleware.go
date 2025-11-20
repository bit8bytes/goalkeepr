package main

import (
	"context"
	"net"
	"net/http"

	"github.com/bit8bytes/goalkeepr/ui/page"
)

// Returns the user id from context. If not found it will return 0
func getUserID(r *http.Request) int {
	if id, ok := r.Context().Value(UserIdKey).(int); ok {
		return id
	}
	return 0
}

func (app *app) withAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := app.sessionManager.GetInt(r.Context(), UserIDSessionKey)
		if userID == 0 {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), UserIdKey, userID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func (app *app) withRate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // fallback
		}

		limiter := app.limiters.get(ip)

		if !limiter.Allow() {
			app.render(w, r, http.StatusTooManyRequests, page.RateLimitExceeded, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *app) cache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.env == "dev" {
			w.Header().Set("Cache-Control", "no-cache")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=3600, must-revalidate")
		}
		next.ServeHTTP(w, r)
	})
}
