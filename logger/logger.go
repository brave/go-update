// Package logger provides logging utilities using Go's standard log/slog package
package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	chiware "github.com/go-chi/chi/v5/middleware"
)

// ContextKey type for storing logger in context
type contextKeyType struct{}

// Key is the context key used to store the logger
var Key = contextKeyType{}

// New creates a new slog.Logger with the default text handler
func New() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// FromContext retrieves the logger from the context
// If no logger is found, it returns the default logger
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(Key).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

// WithContext adds a logger to the context and returns the new context
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, Key, logger)
}

// Setup creates a new logger and adds it to the context
// It also sets the default logger for code that doesn't use context
func Setup(ctx context.Context) (context.Context, *slog.Logger) {
	logger := New()
	slog.SetDefault(logger)
	return WithContext(ctx, logger), logger
}

// Panic logs a message at panic level and then calls panic
func Panic(log *slog.Logger, msg string, args ...any) {
	if log != nil {
		log.Error(msg, args...)
	}
	// Still call panic after logging
	panic(msg)
}

// RequestLoggerMiddleware is a middleware that logs HTTP requests
// It adds a request-specific logger to the context and logs request completion
func RequestLoggerMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a request-specific logger with request ID
			requestID := chiware.GetReqID(r.Context())
			reqLogger := log.With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)

			// Add the logger to the request context
			r = r.WithContext(WithContext(r.Context(), reqLogger))

			// Use chi middleware to track response
			ww := chiware.NewWrapResponseWriter(w, r.ProtoMajor)

			// Call the next handler
			next.ServeHTTP(ww, r)

			// Log the response
			reqLogger.Info("Request completed",
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(start).String(),
			)
		})
	}
}
