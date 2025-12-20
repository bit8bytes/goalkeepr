package main

import (
	"context"
	"net"
	"net/http"

	"github.com/bit8bytes/goalkeepr/internal/users"
	"github.com/bit8bytes/goalkeepr/ui/page"
)

// Returns the user id from context. If not found it will return 0
func getUserID(r *http.Request) int {
	if id, ok := r.Context().Value(users.Key).(int); ok {
		return id
	}
	return 0
}

func (app *app) withAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := app.sessionManager.GetInt(r.Context(), string(users.Key))
		if userID == 0 {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), users.Key, userID)
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
