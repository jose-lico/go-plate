package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"go-plate/internal/auth"
	"go-plate/internal/database"
	"go-plate/internal/middleware"
	"go-plate/internal/ratelimiting"
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

func (s *Service) RegisterRoutes(v1 chi.Router) chi.Router {
	userRouter := chi.NewRouter()
	v1.Mount("/users", userRouter)

	userRouter.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(ratelimiting.NewInMemoryTokenBucket(0.05, 3, 10*time.Minute)))

		r.Post("/register", s.createUser)
		r.Post("/login", s.getUser)
	})

	userRouter.Group(func(r chi.Router) {
		r.Use(middleware.SessionMiddleware(s.redis))

		/*
			Validating the user in the DB defies the purpose of using a cache like
			redis for session management, ideally the session in the redis cache would be
			invalidated correctly if the user was for example, banned.
		*/
		r.Use(ValidateUserMiddleware(s.store, s.redis))

		r.Get("/secret", func(w http.ResponseWriter, r *http.Request) {

			isValidated := r.Context().Value(IsValidated).(bool)

			if isValidated {
				user := r.Context().Value(User).(*models.User)
				w.Write([]byte(fmt.Sprintf("This is a secret from user %d.", user.ID)))
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("You will never get this lalalala."))
			}
		})
	})

	return userRouter
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

	u, err := s.store.CreateUser(&models.User{
		Email:    strings.ToLower(user.Email),
		Password: hashedPassword,
	})

	if err != nil {
		log.Printf("[ERROR] Error creating user: %v", err)
		utils.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	s.generateSession(w, r, u, http.StatusCreated)
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

	s.generateSession(w, r, u, http.StatusOK)
}

func (s *Service) generateSession(w http.ResponseWriter, r *http.Request, u *models.User, status int) {
	sessionToken, err := auth.GenerateToken()
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("could not generate session token"))
	}

	duration := 24 * time.Hour
	expires := time.Now().Add(duration)

	session := &middleware.Session{
		UserID:       int(u.ID),
		Expiration:   expires,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		UserAgent:    r.Header.Get("User-Agent"),
	}

	marshalled, err := json.Marshal(session)
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("failed to marshall session struct"))
		return
	}

	s.redis.Set(r.Context(), "session:"+sessionToken, marshalled, duration)

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Domain:   "",
	})

	w.WriteHeader(status)
}
