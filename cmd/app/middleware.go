package main

import (
	"context"
	"net/http"
	"time"

	"github.com/bit8bytes/goalkeepr/ui/layout"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"golang.org/x/time/rate"
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

func (app *app) withRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		limiter := getRateLimiter(r.RemoteAddr)

		if !limiter.Allow() {
			app.render(w, r, http.StatusTooManyRequests, layout.Center, page.RateLimitExceeded, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getRateLimiter(ip string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	limiter, exists := limiters[ip]
	if !exists {
		limiter =
			rate.NewLimiter(rate.Every(time.Minute/5), 5) // 5 per minute
		limiters[ip] = limiter
	}

	return limiter
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
