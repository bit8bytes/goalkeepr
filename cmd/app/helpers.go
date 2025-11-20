package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"github.com/bit8bytes/toolbox/validator"
	"golang.org/x/time/rate"
)

func (app *app) render(w http.ResponseWriter, r *http.Request, status int, page page.Page, data any) {
	ts, ok := app.templateCache[page.Name()]
	if !ok {
		err := fmt.Errorf("template not found in cache; layout %s, page: %s",
			page.Layout().Name(),
			page.Name())
		app.renderError(w, r, err, "Error loading this page.")
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, page.Layout().Name(), data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)

	if _, err := buf.WriteTo(w); err != nil {
		app.logger.ErrorContext(r.Context(), "failed to write response", slog.String("msg", err.Error()))
	}
}

func (app *app) renderError(w http.ResponseWriter, r *http.Request, err error, userMessage string) {
	app.logger.ErrorContext(r.Context(), "error occured", slog.String("msg", err.Error()))

	data := app.newTemplateData(r)
	data.Data = map[string]any{
		"TraceID": r.Context().Value(TraceIdKey).(string),
		"Message": userMessage,
	}

	app.render(w, r, http.StatusInternalServerError, page.Error, data)
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

func validateBranding(f *branding.Form) {
	f.Check(validator.MaxChars(f.Title, 512), "branding_title", "Title cannot exceed 512 characters")
	f.Check(validator.MaxChars(f.Description, 2048), "branding_description", "Description cannot exceed 2048 characters")
}

type trace struct{}

func newTrace() *trace {
	return &trace{}
}

func (t *trace) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), TraceIdKey, generateTraceID())))
	})
}

type TraceHandler struct {
	slog.Handler
}

func newTraceHandler(h slog.Handler) *TraceHandler {
	return &TraceHandler{Handler: h}
}

func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	var attrs []any

	if ctx == nil {
		panic("context is nil")
	}

	if traceID, ok := ctx.Value(TraceIdKey).(string); ok && traceID != "" {
		attrs = append(attrs, slog.String("trace_id", traceID))
	}
	if connID, ok := ctx.Value(ConnIdKey).(string); ok && connID != "" {
		attrs = append(attrs, slog.String("conn_id", connID))
	}

	if len(attrs) > 0 {
		r.AddAttrs(slog.Group("request", attrs...))
	}
	return h.Handler.Handle(ctx, r)
}

func generateTraceID() string {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		panic(err) // crypto/rand should never fail
	}
	return hex.EncodeToString(bytes)
}

func (app *app) flash(ctx context.Context) *flash {
	if flashMsg := app.sessionManager.PopString(ctx,
		"flash"); flashMsg != "" {
		return &flash{Content: flashMsg}
	}
	return nil
}

func (app *app) putFlash(ctx context.Context, msg string) {
	app.sessionManager.Put(ctx, "flash", msg)
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
		app.logger.InfoContext(r.Context(), "request",
			slog.String("addr", r.RemoteAddr),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
		)

		next.ServeHTTP(w, r)
	})
}

type limiters struct {
	mu sync.Mutex
	m  map[string]*rate.Limiter
}

func newLimiters() *limiters {
	return &limiters{
		m: make(map[string]*rate.Limiter),
	}
}

func (l *limiters) get(ip string) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	limiter, exists := l.m[ip]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(time.Minute/5), 5)
		l.m[ip] = limiter
	}

	return limiter
}
