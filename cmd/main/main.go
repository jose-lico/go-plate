package main

import (
	"log"

	"go-plate/internal/api"
	"go-plate/internal/config"
)

func main() {
	cfg, err := config.NewAPIConfig()

	if err != nil {
		log.Fatalf("[FATAL] Error loading API Config: %v", err)
	}

	server := api.NewAPIServer(cfg)

	err = server.Run()

	if err != nil {
		log.Fatalf("[FATAL] Error launching server: %v", err)
	}
}
