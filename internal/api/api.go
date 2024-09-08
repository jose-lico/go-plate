package api

import (
	"fmt"
	"log"
	"net/http"

	"go-plate/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/rs/cors"
)

type APIServer struct {
	cfg *config.APIConfig
}

func NewAPIServer(cfg *config.APIConfig) *APIServer {
	return &APIServer{cfg: cfg}
}

func (s *APIServer) Run() error {
	router := chi.NewRouter()

	cors := cors.New(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   s.cfg.AllowedMethods,
		AllowedHeaders:   s.cfg.AllowedHeaders,
		ExposedHeaders:   s.cfg.AllowedHeaders,
		AllowCredentials: s.cfg.AllowCredentials,
		MaxAge:           s.cfg.MaxAge,
	})

	router.Use(cors.Handler)

	router.Mount("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	}))

	addr := ":" + s.cfg.Port
	log.Printf("[TRACE] Starting api server on %s", addr)
	return http.ListenAndServe(addr, router)
}
