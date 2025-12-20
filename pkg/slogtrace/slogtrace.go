// Package slogtrace provides a structured logging handler that adds trace ID
// context to log records for distributed tracing support.
package slogtrace

import (
	"context"
	"crypto/rand"
	"io"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel/trace"
)

// ContextKey is a type for context keys to avoid collisions.
type ContextKey string

const (
	// CtxTraceIdKey is the context key used to store and retrieve trace IDs.
	CtxTraceIdKey ContextKey = "TRACE_ID_KEY"
)

// TraceHandler is a slog.Handler wrapper that automatically injects trace IDs
// from the context into log records.
type TraceHandler struct {
	slog.Handler
}

// NewHandler creates a new TraceHandler that wraps the given slog.Handler.
// The returned handler will automatically add trace_id attributes to log
// records when a trace ID is present in the context.
func NewHandler(h slog.Handler) *TraceHandler {
	return &TraceHandler{Handler: h}
}

// NewTextHandler creates a TraceHandler wrapping a TextHandler.
// This is a convenience function equivalent to NewHandler(slog.NewTextHandler(w, opts)).
func NewTextHandler(w io.Writer, opts *slog.HandlerOptions) *TraceHandler {
	return NewHandler(slog.NewTextHandler(w, opts))
}

// NewJSONHandler creates a TraceHandler wrapping a JSONHandler.
// This is a convenience function equivalent to NewHandler(slog.NewJSONHandler(w, opts)).
func NewJSONHandler(w io.Writer, opts *slog.HandlerOptions) *TraceHandler {
	return NewHandler(slog.NewJSONHandler(w, opts))
}

// Enabled reports whether the handler handles records at the given level.
// The handler defers to the wrapped handler's implementation.
func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

// Handle processes a log record and adds the trace ID attribute if present in
// the context. If the context is nil, it falls back to using context.Background.
// The trace ID is retrieved from the context using CtxTraceIdKey and added as
// a "trace_id" attribute to the log record.
func (h *TraceHandler) Handle(ctx context.Context, r slog.Record) error {
	if ctx == nil {
		return h.Handler.Handle(context.Background(), r) // fallback
	}

	if traceID, ok := ctx.Value(CtxTraceIdKey).(string); ok && traceID != "" {
		r.AddAttrs(slog.String("trace_id", traceID))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs returns a new TraceHandler with additional attributes.
// The attributes are added to the wrapped handler, maintaining the trace ID functionality.
func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new TraceHandler with a group name.
// The group is added to the wrapped handler, maintaining the trace ID functionality.
func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return &TraceHandler{Handler: h.Handler.WithGroup(name)}
}

// Middleware creates an HTTP middleware that injects trace IDs into request contexts.
//
// A unique trace ID is generated for each request and stored in the context using the
// slogtrace.CtxTraceIdKey. This ID can be used for request correlation in logs and error pages.
//
// Usage:
//
//	http.Handle("/", slogtrace.Middleware(handler))
//
// Parameters:
//   - next: The next handler in the middleware chain
//
// Returns an HTTP handler that wraps the next handler with trace ID injection.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := generateTraceID()
		ctx := context.WithValue(r.Context(), CtxTraceIdKey, traceID.String())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TelemetryTracer is a middleware component that manages request trace IDs.
//
// Deprecated: Use Middleware function directly instead.
// It generates unique trace IDs for each request and adds them to the request context,
// enabling distributed tracing and request correlation across logs and services.
type TelemetryTracer struct{}

// generateTraceID creates a new random 128-bit trace ID.
//
// The trace ID is generated using cryptographically secure random bytes and conforms
// to the OpenTelemetry trace.TraceID format. Panics if random generation fails, as
// this indicates a critical system error.
//
// Returns a new trace.TraceID.
func generateTraceID() trace.TraceID {
	var traceID [16]byte
	if _, err := rand.Read(traceID[:]); err != nil {
		panic(err)
	}
	return trace.TraceID(traceID)
}
