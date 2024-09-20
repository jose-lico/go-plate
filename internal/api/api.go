package api

import (
	"log"
	"net/http"

	"github.com/jose-lico/go-plate/internal/config"
	"github.com/jose-lico/go-plate/internal/database"
	"github.com/jose-lico/go-plate/internal/middleware"
	"github.com/jose-lico/go-plate/services/post"
	"github.com/jose-lico/go-plate/services/user"

	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

type APIServer struct {
	cfg   *config.APIConfig
	sql   *gorm.DB
	redis database.RedisStore
}

func NewAPIServer(cfg *config.APIConfig, sql *gorm.DB, redis database.RedisStore) *APIServer {
	return &APIServer{cfg: cfg, sql: sql, redis: redis}
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

	router.Use(chi_middleware.RequestID)
	router.Use(chi_middleware.RealIP)
	router.Use(chi_middleware.Logger)
	router.Use(chi_middleware.Recoverer)

	subRouter := chi.NewRouter()
	router.Mount("/api", subRouter)

	v1Router := chi.NewRouter()
	v2Router := chi.NewRouter()
	subRouter.Mount("/v1", v1Router)
	subRouter.Mount("/v2", v2Router)
	v1Router.Use(middleware.VersionURLMiddleware("v1"))
	v2Router.Use(middleware.VersionURLMiddleware("v2"))

	userStore := user.NewStore(s.sql)
	userService := user.NewService(userStore, s.redis)
	userRouter := userService.RegisterRoutes(v1Router)

	postStore := post.NewStore(s.sql)
	postServer := post.NewService(postStore, s.redis)
	postServer.RegisterRoutes(v1Router, v2Router, userRouter)

	addr := ":" + s.cfg.Port
	log.Printf("[TRACE] Starting API server on %s", addr)
	return http.ListenAndServe(addr, router)
}
