package main

import (
	"log"
	"os"

	"go-plate/internal/api"
	"go-plate/internal/config"
	"go-plate/internal/database"

	"github.com/joho/godotenv"
)

func main() {
	env := os.Getenv("ENV")

	if env == "LOCAL" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("[FATAL] Error loading .env: %v", err)
		}
	}

	rcfg, err := config.NewRedisConfig()

	if err != nil {
		log.Fatalf("[FATAL] Error loading Redis Config: %v", err)
	}

	redis, err := database.NewRedis(rcfg)

	if err != nil {
		log.Fatalf("[FATAL] Error connecting to Redis: %v", err)
	}

	cfg, err := config.NewAPIConfig()

	if err != nil {
		log.Fatalf("[FATAL] Error loading API Config: %v", err)
	}

	api := api.NewAPIServer(cfg, redis)

	err = api.Run()

	if err != nil {
		log.Fatalf("[FATAL] Error launching API server: %v", err)
	}
}
