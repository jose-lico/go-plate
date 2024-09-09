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
	// Load env variables for local dev
	env := os.Getenv("ENV")

	if env == "LOCAL" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("[FATAL] Error loading .env: %v", err)
		}
	}

	// Setup redis
	redisCFG, err := config.NewRedisConfig()

	if err != nil {
		log.Fatalf("[FATAL] Error loading Redis Config: %v", err)
	}

	redis, err := database.NewRedis(redisCFG)

	if err != nil {
		log.Fatalf("[FATAL] Error connecting to Redis: %v", err)
	}

	// Setup sql
	sqlCFG, err := config.NewSQLGormConfig()

	if err != nil {
		log.Fatalf("[FATAL] Error loading SQL Config: %v", err)
	}

	sql, err := database.NewSQLGormDB(sqlCFG)

	if err != nil {
		log.Fatalf("[FATAL] Error connecting to SQL: %v", err)
	}

	// Setup api server
	cfg, err := config.NewAPIConfig()

	if err != nil {
		log.Fatalf("[FATAL] Error loading API Config: %v", err)
	}

	api := api.NewAPIServer(cfg, redis, sql)

	err = api.Run()

	if err != nil {
		log.Fatalf("[FATAL] Error launching API server: %v", err)
	}
}
