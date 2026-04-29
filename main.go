package main

import (
	"log"
	"os"

	"sorint-fleet/internal/config"
	"sorint-fleet/internal/router"
)

func main() {
	cfg := config.LoadConfig()
	config.InitDB(cfg)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := router.Setup()

	log.Printf("Sorint Fleet API avviata su :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("errore avvio server: %v", err)
	}
}
