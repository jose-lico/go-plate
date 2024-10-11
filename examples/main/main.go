package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jose-lico/go-plate/api"
	"github.com/jose-lico/go-plate/config"
	"github.com/jose-lico/go-plate/database"
	"github.com/jose-lico/go-plate/examples/internal/services/post"
	"github.com/jose-lico/go-plate/examples/internal/services/user"
	"github.com/jose-lico/go-plate/middleware"
	"github.com/jose-lico/go-plate/utils"

	_ "github.com/jose-lico/go-plate/docs"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Load env variables for local dev
	env := os.Getenv("ENV")

	if env == "LOCAL" {
		err := utils.LoadEnvs()
		if err != nil {
			log.Fatalf("[FATAL] Error loading .env: %v", err)
		}
	}

	// Setup sql
	sqlCFG := config.NewSQLConfig()
	sql, err := database.NewSQLGormDB(sqlCFG)
	if err != nil {
		log.Fatalf("[FATAL] Error connecting to SQL: %v", err)
	}

	// Setup redis
	redisCFG := config.NewRedisConfig()
	redis, err := database.NewRedis(redisCFG)
	if err != nil {
		log.Fatalf("[FATAL] Error connecting to Redis: %v", err)
	}

	// Setup api server
	cfg := config.NewAPIConfig()
	api := api.NewAPIServer(cfg)
	api.UseDefaultMiddleware()

	subRouter := chi.NewRouter()
	api.Router.Mount("/api", subRouter)

	v1Router := chi.NewRouter()
	v2Router := chi.NewRouter()
	subRouter.Mount("/v1", v1Router)
	subRouter.Mount("/v2", v2Router)
	v1Router.Use(middleware.VersionURLMiddleware("v1"))
	v2Router.Use(middleware.VersionURLMiddleware("v2"))

	userStore := user.NewStore(sql)
	userService := user.NewService(userStore, redis)
	userRouter := userService.RegisterRoutes(v1Router)

	postStore := post.NewStore(sql)
	postServer := post.NewService(postStore, redis)
	postServer.RegisterRoutes(v1Router, v2Router, userRouter)

	api.Router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://%s:%s/swagger/doc.json", cfg.Host, cfg.Port)),
	))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("[TRACE] Starting API server on %s", api.Server.Addr)
		if err := api.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Listen and serve error: %v", err)
		}

		log.Println("[TRACE] Stopped serving new connections.")
	}()

	<-ctx.Done()

	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()

	if err := api.Server.Shutdown(shutdownContext); err != nil {
		log.Printf("[ERROR] Server shutdown returned an error: %v\n", err)
	}

	log.Println("[TRACE] Server shutdown")
}
