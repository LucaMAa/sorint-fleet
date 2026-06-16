package main

import (
	"context"
	"log"
	"os"

	"sorint-fleet/internal/bootstrap"
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/cron"
	"sorint-fleet/internal/router"
	"sorint-fleet/internal/search"
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

	if _, err := search.NewClientFromEnv(); err != nil {
		log.Printf("Elasticsearch not configured: %v", err)
	} else {
		if err := search.InitIndices(); err != nil {
			log.Printf("failed init ES indices: %v", err)
		}
		go func() {
			if err := search.IndexAllVehicles(); err != nil {
				log.Printf("failed indexing existing vehicles: %v", err)
			}
		}()
	}

	r := router.Setup(sessionStore)

	r.Run(":" + port)
}
