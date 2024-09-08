package main

import (
	"log"

	"go-plate/internal/api"
	"go-plate/internal/config"
)

func main() {
	cfg := config.NewAPIConfig()

	cfg.Host = "localhost"
	cfg.Port = 8000

	server := api.NewAPIServer(cfg)

	err := server.Run()

	if err != nil {
		log.Panic(err)
	}
}
