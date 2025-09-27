package main

import (
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

// Returns the user id from context. If not found it will return 0
func getUserID(r *http.Request) int {
	if id, ok := r.Context().Value(userIDKey).(int); ok {
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

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
