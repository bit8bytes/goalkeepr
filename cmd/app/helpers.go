package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/ui/layout"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"github.com/bit8bytes/toolbox/validator"
)

func (app *app) render(w http.ResponseWriter, r *http.Request, status int, templateLayout, templatePage string, data any) {
	ts, ok := app.templateCache[templatePage]
	if !ok {
		app.logger.Error("Template not found in cache")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, templateLayout, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)

	buf.WriteTo(w)
}

func (app *app) renderError(w http.ResponseWriter, r *http.Request, err error, userMessage string) {
	connTraceID := getConnTraceID(r)
	reqTraceID := getRequestTraceID(r)

	app.logger.Error("error occured",
		slog.Group("trace", "conn", connTraceID, "req", reqTraceID),
		slog.String("msg", err.Error()),
	)

	data := app.newTemplateData(r)
	data.Data = map[string]any{
		"TraceID": reqTraceID,
		"Message": userMessage,
	}

	app.render(w, r, http.StatusInternalServerError, layout.Center, page.Error, data)
}

func (app *app) newTemplateData(r *http.Request) *templateData {
	userID := app.sessionManager.GetInt(r.Context(), UserIDSessionKey)
	return &templateData{
		Metadata: metadata{
			Year: time.Now().Year(),
		},
		IsAuthenticated: userID != 0,
	}
}

type formValidator interface {
	Check(ok bool, key, message string)
	Valid() bool
}

func validateEmail(form formValidator, email string) {
	form.Check(validator.NotBlank(email), "email", "This field cannot be blank")
	form.Check(validator.Matches(email, validator.EmailRX), "email", "This field must be a valid email address")
}

func validatePassword(form formValidator, password string) {
	form.Check(validator.NotBlank(password), "password", "This field cannot be blank")
	form.Check(validator.MinChars(password, 8), "password", "This field must be at least 8 characters long")
}

func validateRepeatPassword(form formValidator, password, repeatPassword string) {
	form.Check(validator.NotBlank(repeatPassword), "repeat_password", "This field cannot be blank")
	form.Check(password == repeatPassword, "repeat_password", "Passwords do not match")
}

func validateAddGoal(f *goals.Form) {
	f.Check(validator.NotBlank(f.Goal), "goal", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Goal, 1024), "goal", "Goal cannot exceed 1024 characters")
	f.Check(validator.NotBlank(f.Due), "due", "This field cannot be blank")
	f.Check(validator.PermittedValue(f.VisibleToPublic, true, false), "visible", "This field can only be set or unset")
}

func validateEditGoal(f *goals.Form) {
	f.Check(validator.NotBlank(f.Goal), "goal", "This field cannot be blank")
	f.Check(validator.MaxChars(f.Goal, 1024), "goal", "Goal cannot exceed 1024 characters")
	f.Check(validator.NotBlank(f.Due), "due", "This field cannot be blank")
	f.Check(validator.PermittedValue(f.Achieved, true, false), "achieved", "This field can only be set or unset")
	f.Check(validator.PermittedValue(f.VisibleToPublic, true, false), "visible", "This field can only be set or unset")
}

func validateBranding(f *brandingForm) {
	f.Check(validator.MaxChars(f.Title, 512), "branding_title", "Title cannot exceed 512 characters")
	f.Check(validator.MaxChars(f.Description, 2048), "branding_description", "Description cannot exceed 2048 characters")
}

type trace struct{}

const ReqTraceIDKey contextKey = "reqTraceID"
const ConnTraceIDKey contextKey = "connTraceID"

func NewTrace() *trace {
	return &trace{}
}

func (t *trace) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := generateTraceID()
		ctx := context.WithValue(r.Context(), ReqTraceIDKey, traceID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

func generateTraceID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getRequestTraceID(r *http.Request) string {
	traceID, _ := r.Context().Value(ReqTraceIDKey).(string)
	return traceID
}

func getConnTraceID(r *http.Request) string {
	traceID, _ := r.Context().Value(ConnTraceIDKey).(string)
	return traceID
}

func commonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		w.Header().Set("Server", "Go")

		next.ServeHTTP(w, r)
	})
}

func (app *app) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connTraceID := getConnTraceID(r)
		reqTraceID := getRequestTraceID(r)

		app.logger.Info("request",
			slog.Group("trace", "conn", connTraceID, "req", reqTraceID),
			slog.String("addr", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		next.ServeHTTP(w, r)
	})
}
