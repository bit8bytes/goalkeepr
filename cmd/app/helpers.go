package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/users"
	"github.com/bit8bytes/goalkeepr/ui/page"
	"golang.org/x/time/rate"
)

type flash struct {
	Content string
}

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
	data.Data = ErrorPageData{
		TraceID: r.Context().Value(TraceIdKey).(string),
		Message: userMessage,
	}

	app.render(w, r, http.StatusInternalServerError, page.Error, data)
}

func (app *app) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t") // json.Marshal is faster then MarshalIndent
	if err != nil {
		return err
	}

	js = append(js, '\n')

	maps.Copy(w.Header(), headers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)
	if err != nil {
		return err
	}

	return nil
}

func (app *app) newTemplateData(r *http.Request) *templateData {
	userID := app.sessionManager.GetInt64(r.Context(), string(users.Key))
	return &templateData{
		Metadata: metadata{
			Year: time.Now().Year(),
		},
		IsAuthenticated: userID != 0,
	}
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
		// Allow 60 requests per minute (1 request per second with burst of 5)
		limiter = rate.NewLimiter(rate.Every(time.Second), 5)

		l.m[ip] = limiter
	}

	return limiter
}
