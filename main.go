package main

import (
	"log"
	"os"

	"sorint-fleet/internal/bootstrap"
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/cron"
	"sorint-fleet/internal/router"
)

func main() {
	cfg := config.LoadConfig()
	config.InitDB(cfg)

	bootstrap.Admin()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	c := cron.New()
	c.Start()
	defer c.Stop()

	r := router.Setup()

	log.Printf("Sorint Fleet API avviata su :%s", port)
	r.Run(":" + port)
}
