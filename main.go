package main

import (
	"context"
	"log"
	"os"

	"sorint-fleet/internal/bootstrap"
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/cron"
	"sorint-fleet/internal/router"
	"sorint-fleet/internal/session"
)

func main() {
	cfg := config.LoadConfig()
	config.InitDB(cfg)

	sessionStore := session.New()
	if err := sessionStore.Ping(context.Background()); err != nil {
		log.Fatalf("Redis not available: %v", err)
	}

	bootstrap.Admin()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	c := cron.New()
	c.Start()
	defer c.Stop()

	r := router.Setup(sessionStore)

	r.Run(":" + port)
}
