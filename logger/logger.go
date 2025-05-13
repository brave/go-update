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

// New creates a new logger with httplog's default configuration
func New() *slog.Logger {
	logger := httplog.NewLogger("go-update", httplog.Options{
		LogLevel:         slog.LevelInfo,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
		Tags: map[string]string{
			"service": "go-update",
		},
	})

	// Return the slog.Logger that httplog is built upon
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

// Panic logs a message at panic level and then calls panic
func Panic(log *slog.Logger, msg string, args ...any) {
	if log != nil {
		log.Error(msg, args...)
	}
	// Still call panic after logging
	panic(msg)
}

// RequestLoggerMiddleware is a middleware that logs HTTP requests using httplog
func RequestLoggerMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	// Create a new httplog logger from the provided slog logger
	httpLogger := httplog.NewLogger("go-update", httplog.Options{
		LogLevel:         slog.LevelInfo,
		Concise:          true,
		RequestHeaders:   true,
		MessageFieldName: "message",
		Tags: map[string]string{
			"service": "go-update",
		},
	})

	return httplog.RequestLogger(httpLogger)
}

// LogEntrySetField adds a field to the current request's log entry
func LogEntrySetField(ctx context.Context, key string, value slog.Value) {
	httplog.LogEntrySetField(ctx, key, value)
}
