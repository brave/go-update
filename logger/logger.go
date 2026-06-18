// Package logger provides logging utilities using the go-chi/httplog package
package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/httplog/v3"
)

type contextKeyType struct{}

var key = contextKeyType{}

// New creates a new logger with generic configuration
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// FromContext retrieves the logger from the context
// If no logger is found, it returns the default logger
func FromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(key).(*slog.Logger); ok && logger != nil {
		return logger
	}
	return slog.Default()
}

// Setup creates a new logger and adds it to the context
// It also sets the default logger for code that doesn't use context
func Setup(ctx context.Context) (context.Context, *slog.Logger) {
	logger := New()
	slog.SetDefault(logger)
	return context.WithValue(ctx, key, logger), logger
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
	log := New()
	return httplog.RequestLogger(log, &httplog.Options{
		Level:  slog.LevelInfo,
		Schema: httplog.SchemaECS.Concise(true),
		LogRequestHeaders: []string{
			"Accept",
			"Accept-Encoding",
			"Content-Type",
		},
		LogResponseHeaders: []string{
			"Content-Encoding",
			"Content-Type",
		},
		RecoverPanics: true,
	})
}
