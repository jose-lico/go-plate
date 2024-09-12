package post

import (
	"fmt"
	"log"
	"net/http"

	"go-plate/internal/database"
	"go-plate/internal/middleware"
	"go-plate/internal/utils"
	"go-plate/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	store PostStore
	redis database.RedisStore
}

func NewService(store PostStore, redis database.RedisStore) *Service {
	return &Service{store: store, redis: redis}
}

func (s *Service) RegisterRoutes(v1 chi.Router, v2 chi.Router) {
	postRouter := chi.NewRouter()
	v1.Mount("/posts", postRouter)
	v2.Mount("/posts", postRouter)

	postRouter.Group(func(r chi.Router) {
		r.Use(middleware.SessionMiddleware(s.redis))
		r.Post("/", s.createPost)
	})
}

func (s *Service) createPost(w http.ResponseWriter, r *http.Request) {
	isAuthenticated := r.Context().Value(middleware.IsAuthenticated).(bool)

	if isAuthenticated {
		var post PostPayload
		if err := utils.ParseJSON(r, &post); err != nil {
			utils.WriteError(w, http.StatusBadRequest, err)
			return
		}

		if err := utils.Validate.Struct(post); err != nil {
			errors := err.(validator.ValidationErrors)
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
			return
		}

		version := r.Context().Value(middleware.Version).(string)

		// Pretend v1 does not support summaries
		if version == "v1" {
			post.Summary = ""
		}

		session := r.Context().Value(middleware.SessionInfo).(middleware.Session)

		_, err := s.store.CreatePost(&models.Post{
			Title:   post.Title,
			Summary: post.Summary,
			Content: post.Content,
			UserID:  uint(session.UserID),
		})

		if err != nil {
			log.Printf("[ERROR] Error creating post: %v", err)
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}
