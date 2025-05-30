// Package logger provides logging utilities using the go-chi/httplog package
package logger

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v2"
)

// ContextKey type for storing logger in context
type contextKeyType struct{}

// Key is the context key used to store the logger
var Key = contextKeyType{}

// New creates a new logger with generic configuration
func New() *slog.Logger {
	// Use httplog (built on log/slog) but configured for generic logging (not HTTP-specific)
	logger := httplog.NewLogger("go-update", httplog.Options{
		LogLevel:         slog.LevelInfo,
		Concise:          true,
		RequestHeaders:   false, // No HTTP request headers in generic logger
		MessageFieldName: "message",
	})

	return logger.Logger
}

// FromContext retrieves the logger from the context
// If no logger is found, it returns the default logger
func FromContext(ctx context.Context) *slog.Logger {
	if logger := httplog.LogEntry(ctx); logger != nil {
		return logger
	}
	return slog.Default()
}

// WithContext adds a logger to the context and returns the new context
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	// httplog provides its own context management, but we'll keep this for compatibility
	return context.WithValue(ctx, Key, logger)
}

// Setup creates a new logger and adds it to the context
// It also sets the default logger for code that doesn't use context
func Setup(ctx context.Context) (context.Context, *slog.Logger) {
	logger := New()
	slog.SetDefault(logger)
	return WithContext(ctx, logger), logger
}

// Panic logs a message at error level and then panics with the provided error
func Panic(log *slog.Logger, msg string, err error) {
	if log != nil {
		log.Error(msg, "error", err)
	}
	panic(err)
}

// RequestLoggerMiddleware is a middleware that logs HTTP requests using httplog
func RequestLoggerMiddleware() func(http.Handler) http.Handler {
	// Create a httplog logger for request logging
	httpLogger := httplog.NewLogger("go-update", httplog.Options{
		LogLevel:         slog.LevelInfo,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
	})

	return httplog.RequestLogger(httpLogger)
}
