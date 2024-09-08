package api

import (
	"fmt"
	"log"
	"net/http"

	"go-plate/internal/config"

	"github.com/go-chi/chi/v5"
)

type APIServer struct {
	cfg *config.APIConfig
}

func NewAPIServer(cfg *config.APIConfig) *APIServer {
	return &APIServer{cfg: cfg}
}

func (s *APIServer) Run() error {
	router := chi.NewRouter()

	router.Mount("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	}))

	addr := ":" + s.cfg.Port
	log.Printf("Starting api server on %s", addr)
	return http.ListenAndServe(addr, router)
}
