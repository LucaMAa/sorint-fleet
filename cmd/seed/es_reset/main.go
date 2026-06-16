package main

import (
	"log"
	"sorint-fleet/internal/config"
	"sorint-fleet/internal/search"
	"sorint-fleet/internal/seed"
)

func main() {
	cfg := config.LoadConfig()
	config.InitDB(cfg)

	if _, err := search.NewClientFromEnv(); err != nil {
		log.Fatal(err)
	}

	if err := seed.ResetElasticsearch(); err != nil {
		log.Fatal(err)
	}

	log.Println("Elasticsearch reset completed")
}
