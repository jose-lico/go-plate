package api

import (
	"net/http"

	"github.com/jose-lico/go-plate/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

type APIServer struct {
	Server *http.Server
	Router *chi.Mux
	cfg    *config.APIConfig
}

func NewAPIServer(cfg *config.APIConfig) *APIServer {
	server := &APIServer{Server: &http.Server{Addr: ":" + cfg.Port}, Router: chi.NewRouter(), cfg: cfg}
	server.Server.Handler = server.Router

	return server
}

func (s *APIServer) UseDefaultMiddleware() {
	cors := cors.New(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   s.cfg.AllowedMethods,
		AllowedHeaders:   s.cfg.AllowedHeaders,
		ExposedHeaders:   s.cfg.AllowedHeaders,
		AllowCredentials: s.cfg.AllowCredentials,
		MaxAge:           s.cfg.MaxAge,
	})

	s.Router.Use(cors.Handler)

	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.Recoverer)
}
