package user

import (
	"fmt"
	"net/http"
	"time"

	"go-plate/internal/database"
	"go-plate/internal/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

type Service struct {
	redis database.RedisStore
}

func NewService(redis database.RedisStore) *Service {
	return &Service{redis: redis}
}

func (s *Service) RegisterRoutes(router chi.Router) {
	version1Router := chi.NewRouter()
	version2Router := chi.NewRouter()

	version1Router.Use(middleware.VersionURLMiddleware("v1"))
	version2Router.Use(middleware.VersionURLMiddleware("v2"))

	router.Mount("/v1", version1Router)
	router.Mount("/v2", version2Router)

	userRouter := chi.NewRouter()
	version1Router.Mount("/user", userRouter)
	version2Router.Mount("/user", userRouter)

	userRouter.Get("/{id}", s.getUser)
	userRouter.Post("/{id}", s.createUser)
}

func (s *Service) createUser(w http.ResponseWriter, r *http.Request) {
	duration := 24 * time.Hour

	id := r.PathValue("id")

	s.redis.Set(r.Context(), "user:"+id, id, duration)

	w.WriteHeader(http.StatusCreated)
}

func (s *Service) getUser(w http.ResponseWriter, r *http.Request) {
	idURL := r.PathValue("id")
	id, err := s.redis.Get(r.Context(), "user:"+idURL)

	if err == redis.Nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "No user with ID %s", idURL)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	version := r.Context().Value(middleware.Version).(string)
	if version == "v1" {
		fmt.Fprintf(w, "This is a user in v1. ID %s", id)
	} else if version == "v2" {
		fmt.Fprintf(w, "This is a user in v2. ID %s", id)
	}
}
