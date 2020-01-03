package main

import (
	"github.com/brave/go-update/server"
	"github.com/getsentry/sentry-go"
	"log"
	"os"
)

func main() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	})
	if err != nil {
		log.Printf("failed to init sentry-go %v\n", err)
	}
	server.StartServer()
}
