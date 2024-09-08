package api

import (
	"fmt"
	"go-plate/internal/config"
	"net/http"

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

	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port), router)
}
