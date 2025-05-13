// Package server implements the web server for extension update requests
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	_ "net/http/pprof" // pprof magic
	"os"
	"strconv"
	"time"

	"github.com/brave-intl/bat-go/middleware"
	"github.com/brave/go-update/controller"
	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/logger"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
)

func setupRouter(ctx context.Context, log *slog.Logger, testRouter bool) (context.Context, *chi.Mux) {
	r := chi.NewRouter()
	r.Use(chiware.RequestID)
	r.Use(chiware.RealIP)
	r.Use(chiware.Compress(5, "application/*", "text/*"))
	r.Use(chiware.Heartbeat("/"))
	r.Use(chiware.Timeout(60 * time.Second))
	r.Use(middleware.BearerToken)
	shouldLog, ok := os.LookupEnv("LOG_REQUEST")
	if ok && shouldLog == "true" && log != nil {
		// Use our custom slog-based request logger
		r.Use(logger.RequestLoggerMiddleware(log))
	}
	extensions := extension.OfferedExtensions
	r.Mount("/extensions", controller.ExtensionsRouter(extensions, testRouter))
	return ctx, r
}

// StartServer starts the component updater server on port 8192
func StartServer() {
	serverCtx, log := logger.Setup(context.Background())
	log.Info("Starting server", "prefix", "main")

	go func() {
		// setup metrics on another non-public port 9090
		err := http.ListenAndServe(":9090", middleware.Metrics())
		if err != nil {
			sentry.CaptureException(err)
			logger.Panic(log, "Metrics HTTP server failed to start", err)
		}
	}()

	// Add profiling flag to enable profiling routes.
	if on, _ := strconv.ParseBool(os.Getenv("PPROF_ENABLED")); on {
		// pprof attaches routes to default serve mux
		// host:6061/debug/pprof/
		go func() {
			if err := http.ListenAndServe(":6061", http.DefaultServeMux); err != nil {
				log.Error("Server failed to start", "error", err)
			}
		}()
	}

	serverCtx, r := setupRouter(serverCtx, log, false)
	port := ":8192"
	log.Info("Starting HTTP server", "url", fmt.Sprintf("http://localhost%s", port))

	srv := http.Server{
		Addr:        port,
		Handler:     r,
		BaseContext: func(_ net.Listener) context.Context { return serverCtx },
	}
	err := srv.ListenAndServe()
	if err != nil {
		sentry.CaptureException(err)
		logger.Panic(log, "Server failed to start", err)
	}
}
