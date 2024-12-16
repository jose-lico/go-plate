package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/jose-lico/go-plate/api"
	"github.com/jose-lico/go-plate/config"
	"github.com/jose-lico/go-plate/database"
	"github.com/jose-lico/go-plate/examples/internal/services/post"
	"github.com/jose-lico/go-plate/examples/internal/services/user"
	"github.com/jose-lico/go-plate/logger"
	"github.com/jose-lico/go-plate/middleware"

	"github.com/go-chi/chi/v5"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

func main() {
	env := os.Getenv("ENV")

	// Load env variables for local dev
	if env == "LOCAL" {
		err := godotenv.Load()
		if err != nil {
			panic(fmt.Errorf("error loading env file: %w", err))
		}
	}

	// Create zap logger
	logger, err := logger.CreateLogger(env)
	if err != nil {
		panic(fmt.Errorf("error creating logger: %w", err))
	}
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	// Setup sql
	sqlCFG := config.NewSQLConfig()
	sql, err := database.NewSQLGormDB(sqlCFG, logger)
	if err != nil {
		logger.Fatal("Error connecting to SQL", zap.Error(err))
	}

	// Setup redis
	redisCFG := config.NewRedisConfig()
	redis, err := database.NewRedis(redisCFG, logger)
	if err != nil {
		logger.Fatal("Error connecting to Redis", zap.Error(err))
	}

	// Setup api server
	cfg := config.NewAPIConfig()
	api := api.NewAPIServer(cfg)
	api.UseDefaultMiddleware(env, logger)

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
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigterm
		logger.Info("Received termination signal, shutting down...")
		stop()
	}()

	go func() {
		logger.Info("Starting API server", zap.String("Address", api.Server.Addr))
		if err := api.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("Failed ListenAndServe()", zap.Error(err))
		}

		logger.Info("Stopped serving new connections")
	}()

	<-ctx.Done()

	shutdownContext, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer shutdownCancel()

	if err := api.Server.Shutdown(shutdownContext); err != nil {
		logger.Error("Server shutdown returned an error", zap.Error(err))
	}

	logger.Info("Server shutdown")
}
