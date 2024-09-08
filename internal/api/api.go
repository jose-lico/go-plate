package api

import (
	"fmt"
	"log"
	"net/http"

	"go-plate/internal/config"
	user "go-plate/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Mount("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!")
	}))

	subRouter := chi.NewRouter()
	router.Mount("/api", subRouter)

	userService := user.NewService()
	userService.RegisterRoutes(subRouter)

	addr := ":" + s.cfg.Port
	log.Printf("[TRACE] Starting api server on %s", addr)
	return http.ListenAndServe(addr, router)
}
