// Package server implements the web server for extension update requests
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/brave-intl/bat-go/middleware"
	"github.com/brave/go-update/controller"
	"github.com/brave/go-update/extension"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi"
	chiware "github.com/go-chi/chi/middleware"
	"github.com/pressly/lg"
	"github.com/sirupsen/logrus"
)

func setupLogger(ctx context.Context) (context.Context, *logrus.Logger) {
	logger := logrus.New()
	// Redirect output from the standard logging package "log"
	lg.RedirectStdlogOutput(logger)
	lg.DefaultLogger = logger
	ctx = lg.WithLoggerContext(ctx, logger)
	return ctx, logger
}

func setupRouter(ctx context.Context, logger *logrus.Logger, testRouter bool) (context.Context, *chi.Mux) {
	r := chi.NewRouter()
	r.Use(chiware.RequestID)
	r.Use(chiware.RealIP)
	r.Use(chiware.DefaultCompress)
	r.Use(chiware.Heartbeat("/"))
	r.Use(chiware.Timeout(60 * time.Second))
	r.Use(middleware.BearerToken)
	log, ok := os.LookupEnv("LOG_REQUEST")
	if ok && log == "true" && logger != nil {
		// Also handles panic recovery
		r.Use(middleware.RequestLogger(logger))
	}
	extensions := extension.OfferedExtensions
	r.Mount("/extensions", controller.ExtensionsRouter(extensions, testRouter))
	return ctx, r
}

// StartServer starts the component updater server on port 8192
func StartServer() {
	serverCtx, logger := setupLogger(context.Background())
	logger.WithFields(logrus.Fields{"prefix": "main"}).Info("Starting server")

	go func() {
		// setup metrics on another non-public port 9090
		err := http.ListenAndServe(":9090", middleware.Metrics())
		if err != nil {
			sentry.CaptureException(err)
			panic(fmt.Sprintf("metrics HTTP server start failed: %s", err.Error()))
		}
	}()

	serverCtx, r := setupRouter(serverCtx, logger, false)
	port := ":8192"
	fmt.Printf("Starting server: http://localhost%s", port)
	srv := http.Server{Addr: port, Handler: chi.ServerBaseContext(serverCtx, r)}
	err := srv.ListenAndServe()
	if err != nil {
		sentry.CaptureException(err)
		log.Panic(err)
	}
}
