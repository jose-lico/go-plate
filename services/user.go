package user

import (
	"fmt"
	"net/http"

	"go-plate/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
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

	userRouter.Get("/", s.getUser)
}

func (s *Service) getUser(w http.ResponseWriter, r *http.Request) {
	version := r.Context().Value(middleware.Version).(string)
	if version == "v1" {
		fmt.Fprintln(w, "This is a user in v1")
	} else if version == "v2" {
		fmt.Fprintln(w, "This is a user in v2")
	}
}
