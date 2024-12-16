package api

import (
	"net/http"

	"github.com/jose-lico/go-plate/config"
	"github.com/jose-lico/go-plate/middleware"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
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

func (s *APIServer) UseCORS() {
	cors := cors.New(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   s.cfg.AllowedMethods,
		AllowedHeaders:   s.cfg.AllowedHeaders,
		ExposedHeaders:   s.cfg.AllowedHeaders,
		AllowCredentials: s.cfg.AllowCredentials,
		MaxAge:           s.cfg.MaxAge,
	})

	s.Router.Use(cors.Handler)
}

func (s *APIServer) UseDefaultMiddleware(env string, logger *zap.Logger) {
	cors := cors.New(cors.Options{
		AllowedOrigins:   s.cfg.AllowedOrigins,
		AllowedMethods:   s.cfg.AllowedMethods,
		AllowedHeaders:   s.cfg.AllowedHeaders,
		ExposedHeaders:   s.cfg.AllowedHeaders,
		AllowCredentials: s.cfg.AllowCredentials,
		MaxAge:           s.cfg.MaxAge,
	})

	s.Router.Use(cors.Handler)

	s.Router.Use(chiMiddleware.RequestID)
	s.Router.Use(chiMiddleware.RealIP)
	if env == "LOCAL" {
		s.Router.Use(middleware.ZapLoggerMiddlwareDev(logger))
	} else {
		s.Router.Use(middleware.ZapLoggerMiddlwareProd(logger))
	}
	s.Router.Use(chiMiddleware.Recoverer)
}
