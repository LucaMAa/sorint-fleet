package main

import (
	"log"
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/seed"
)

func main() {
	cfg := config.LoadConfig()
	config.InitDB(cfg)

	if err := seed.Run(config.DB); err != nil {
		log.Fatal(err)
	}

	log.Println("seed completed")
}
