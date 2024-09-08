package api

import (
	"log"
	"net/http"

	"go-plate/internal/config"
	"go-plate/internal/database"
	"go-plate/services/user"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
)

type APIServer struct {
	cfg   *config.APIConfig
	redis database.RedisStore
}

func NewAPIServer(cfg *config.APIConfig, redis database.RedisStore) *APIServer {
	return &APIServer{cfg: cfg, redis: redis}
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

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	subRouter := chi.NewRouter()
	router.Mount("/api", subRouter)

	userService := user.NewService(s.redis)
	userService.RegisterRoutes(subRouter)

	addr := ":" + s.cfg.Port
	log.Printf("[TRACE] Starting API server on %s", addr)
	return http.ListenAndServe(addr, router)
}
