package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jose-lico/go-plate/api"
	"github.com/jose-lico/go-plate/config"
)

func main() {
	cfg := config.NewAPIConfig()
	api := api.NewAPIServer(cfg)
	api.UseDefaultMiddleware()

	api.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello World!!!")
	})

	err := api.Run()
	if err != nil {
		log.Fatalf("[FATAL] Error launching API server: %v", err)
	}
}
