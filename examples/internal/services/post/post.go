package post

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jose-lico/go-plate/database"
	"github.com/jose-lico/go-plate/examples/internal/models"
	"github.com/jose-lico/go-plate/middleware"
	"github.com/jose-lico/go-plate/ratelimiting"
	"github.com/jose-lico/go-plate/utils"

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

func (s *Service) RegisterRoutes(v1 chi.Router, v2 chi.Router, userRouter chi.Router) {
	postRouter := chi.NewRouter()
	v1.Mount("/posts", postRouter)
	v2.Mount("/posts", postRouter)

	postRouter.Group(func(r chi.Router) {
		r.Use(middleware.SessionMiddleware(s.redis))

		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimitMiddleware(ratelimiting.NewRedisTokenBucket("/posts", s.redis, 0.1, 20, 10*time.Minute)))

			// `/posts/user/1` returns same as `/users/1/posts`
			r.Get("/user/{id}", s.getPosts)
		})

		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimitMiddleware(ratelimiting.NewRedisTokenBucket("/posts", s.redis, 0.1, 20, 10*time.Minute)))

			r.Post("/", s.createPost)
			r.Patch("/{id}", s.updatePost)
			r.Delete("/{id}", s.deletePost)
		})
	})

	v2.Mount("/users", userRouter)
	userRouter.Group(func(r chi.Router) {
		r.Use(middleware.SessionMiddleware(s.redis))

		// `/users/1/posts` returns same as `/posts/user/1`
		// I prefer this one
		r.Get("/{id}/posts", s.getPosts)
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
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (s *Service) updatePost(w http.ResponseWriter, r *http.Request) {
	isAuthenticated := r.Context().Value(middleware.IsAuthenticated).(bool)

	if isAuthenticated {
		var update EditPostPayload
		if err := utils.ParseJSON(r, &update); err != nil {
			utils.WriteError(w, http.StatusBadRequest, err)
			return
		}

		if err := utils.Validate.Struct(update); err != nil {
			errors := err.(validator.ValidationErrors)
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
			return
		}

		postID := r.PathValue("id")
		postIDAsInt, err := strconv.Atoi(postID)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		post, err := s.store.GetPostByID(postIDAsInt)

		if err == ErrPostNotFound {
			utils.WriteError(w, http.StatusNotFound, err)
			return
		} else if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		session := r.Context().Value(middleware.SessionInfo).(middleware.Session)
		userID := session.UserID

		if post.UserID != uint(userID) {
			utils.WriteError(w, http.StatusForbidden, ErrUserNotAuthorized)
			return
		}

		err = s.store.UpdatePost(post, update)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (s *Service) deletePost(w http.ResponseWriter, r *http.Request) {
	isAuthenticated := r.Context().Value(middleware.IsAuthenticated).(bool)

	if isAuthenticated {
		session := r.Context().Value(middleware.SessionInfo).(middleware.Session)

		userID := session.UserID

		postID := r.PathValue("id")
		postIDAsInt, err := strconv.Atoi(postID)

		if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = s.store.DeletePost(postIDAsInt, userID)

		if err == ErrPostNotFound {
			utils.WriteError(w, http.StatusNotFound, err)
		} else if err == ErrUserNotAuthorized {
			utils.WriteError(w, http.StatusForbidden, err)
		} else if err != nil {
			utils.WriteError(w, http.StatusInternalServerError, err)
		}

		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

func (s *Service) getPosts(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	id := r.PathValue("id")
	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var posts []models.Post

	limit := -1

	isAuthenticated := r.Context().Value(middleware.IsAuthenticated).(bool)

	// Unauthenticated users can only see latest article
	if !isAuthenticated {
		limit = 1
	}

	posts, err = s.store.GetPosts(idAsInt, limit)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	var responseData []PostResponsePayload
	version := r.Context().Value(middleware.Version).(string)

	for _, post := range posts {
		payload := ModelToResponsePayload(&post)

		// Pretend v1 does not support summaries
		if version == "v1" {
			payload.Summary = ""
		}

		responseData = append(responseData, payload)
	}

	if len(responseData) > 0 {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"posts": responseData,
	})
}
