package user

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"go-plate/internal/auth"
	"go-plate/internal/database"
	"go-plate/internal/middleware"
	"go-plate/internal/utils"
	"go-plate/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	store UserStore
	redis database.RedisStore
}

func NewService(store UserStore, redis database.RedisStore) *Service {
	return &Service{store: store, redis: redis}
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

	userRouter.Post("/register", s.createUser)
	userRouter.Post("/login", s.getUser)
}

func (s *Service) createUser(w http.ResponseWriter, r *http.Request) {
	var user RegisterUserPayload
	if err := utils.ParseJSON(r, &user); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	_, err := s.store.GetUserByEmail(user.Email)
	if err == nil {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("user with email %s already exists", user.Email))
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	_, err = s.store.CreateUser(&models.User{
		Email:    strings.ToLower(user.Email),
		Password: hashedPassword,
	})

	if err != nil {
		log.Printf("[ERROR] Error creating user: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Service) getUser(w http.ResponseWriter, r *http.Request) {
	var user LoginUserPayload
	if err := utils.ParseJSON(r, &user); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err)
		return
	}

	if err := utils.Validate.Struct(user); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	u, err := s.store.GetUserByEmail(user.Email)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !auth.ComparePasswords(u.Password, []byte(user.Password)) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}
