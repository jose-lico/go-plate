package post

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jose-lico/go-plate/database"
	"github.com/jose-lico/go-plate/examples/internal/models"
	"github.com/jose-lico/go-plate/middleware"
	"github.com/jose-lico/go-plate/ratelimiting"
	"github.com/jose-lico/go-plate/utils"
	"go.uber.org/zap"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	logger *zap.Logger
	store  PostStore
	redis  database.RedisStore
}

func NewService(logger *zap.Logger, store PostStore, redis database.RedisStore) *Service {
	return &Service{logger: logger, store: store, redis: redis}
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

// @Summary Create a new post
// @Description Creates a new post. V1 does not support post summaries.
// @Tags Posts
// @Accept json
// @Produce json
// @Security ApiCookieAuth
// @Param post body PostPayload true "Post data for creation"
// @Success 201 "Post created successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload"
// @Failure 401 "Unauthorized"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /v1/posts [post]
// @Router /v2/posts [post]
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
			s.logger.Error("Error creating post", zap.Error(err), zap.Any("Post", post))
			utils.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// @Summary Update a post
// @Description Updates an existing post. Only the post owner can perform this action.
// @Tags Posts
// @Accept json
// @Produce json
// @Security ApiCookieAuth
// @Param id path int true "Post ID"
// @Param post body EditPostPayload true "Post update payload"
// @Success 204 "Post updated successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload"
// @Failure 401 "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - user is not the post owner"
// @Failure 404 {object} utils.ErrorResponse "Post not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /v2/posts/{id} [patch]
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
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s is not a valid id: %w", postID, err))
			return
		}

		post, err := s.store.GetPostByID(postIDAsInt)

		if err == ErrPostNotFound {
			utils.WriteError(w, http.StatusNotFound, err)
			return
		} else if err != nil {
			s.logger.Error("Error getting post", zap.Error(err), zap.Any("Post", post))
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
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
			s.logger.Error("Error updating post", zap.Error(err), zap.Any("Post", post))
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// @Summary Delete a post
// @Description Deletes an existing post. Only the post owner can perform this action.
// @Tags Posts
// @Security ApiCookieAuth
// @Param id path int true "Post ID"
// @Success 204 "Post deleted successfully"
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload"
// @Failure 401 "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden - user is not the post owner"
// @Failure 404 {object} utils.ErrorResponse "Post not found"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /v2/posts/{id} [delete]
func (s *Service) deletePost(w http.ResponseWriter, r *http.Request) {
	isAuthenticated := r.Context().Value(middleware.IsAuthenticated).(bool)

	if isAuthenticated {
		session := r.Context().Value(middleware.SessionInfo).(middleware.Session)

		userID := session.UserID

		postID := r.PathValue("id")
		postIDAsInt, err := strconv.Atoi(postID)

		if err != nil {
			utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s is not a valid id: %w", postID, err))
			return
		}

		err = s.store.DeletePost(postIDAsInt, userID)

		if err == ErrPostNotFound {
			utils.WriteError(w, http.StatusNotFound, err)
		} else if err == ErrUserNotAuthorized {
			utils.WriteError(w, http.StatusForbidden, err)
		} else if err != nil {
			s.logger.Error("Error deleting post", zap.Error(err), zap.Int("Post", postIDAsInt))
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
		}

		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
}

// @Summary Get user's posts
// @Description Retrieves posts for a specific user. Unauthenticated users can only see the latest post.
// @Tags Posts
// @Produce json
// @Security ApiCookieAuth
// @Param id path int true "User ID"
// @Success 200 {object} map[string][]string "List of posts"
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /v1/users/{id}/posts [get]
// @Router /v2/users/{id}/posts [get]
func (s *Service) getPosts(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idAsInt, err := strconv.Atoi(id)

	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%s is not a valid id: %w", id, err))
		return
	}

	// Could check here if user even exists to avoid returning empty list with a 200

	var posts []models.Post

	limit := -1

	isAuthenticated := r.Context().Value(middleware.IsAuthenticated).(bool)

	// Unauthenticated users can only see latest article
	if !isAuthenticated {
		limit = 1
	}

	posts, err = s.store.GetPostsByUserID(idAsInt, limit)

	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	responseData := make([]PostResponsePayload, 0)
	version := r.Context().Value(middleware.Version).(string)

	for _, post := range posts {
		payload := ModelToResponsePayload(&post)
		// Pretend v1 does not support summaries
		if version == "v1" {
			payload.Summary = ""
		}
		responseData = append(responseData, payload)
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{"posts": responseData})
}
