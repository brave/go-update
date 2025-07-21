// Package server implements the web server for extension update requests
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // pprof magic
	"os"
	"strconv"
	"time"

	batware "github.com/brave-intl/bat-go/middleware"
	"github.com/brave/go-update/controller"
	"github.com/brave/go-update/extension"
	"github.com/brave/go-update/logger"
	"github.com/brave/go-update/server/middleware"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chiware "github.com/go-chi/chi/v5/middleware"
)

func setupRouter(ctx context.Context, testRouter bool) (context.Context, *chi.Mux) {
	r := chi.NewRouter()
	r.Use(middleware.OptimizedCompress(5, 512, "application/json", "application/xml"))
	r.Use(chiware.Heartbeat("/"))
	r.Use(chiware.Timeout(60 * time.Second))

	shouldLog, ok := os.LookupEnv("LOG_REQUEST")
	if ok && shouldLog == "true" {
		r.Use(logger.RequestLoggerMiddleware())
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
		// nosemgrep: go.lang.security.audit.net.pprof.pprof-debug-exposure
		err := http.ListenAndServe(":9090", batware.Metrics())
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

	serverCtx, r := setupRouter(serverCtx, false)
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
