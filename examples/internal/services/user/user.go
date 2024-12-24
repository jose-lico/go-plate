package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jose-lico/go-plate/auth"
	"github.com/jose-lico/go-plate/database"
	"github.com/jose-lico/go-plate/examples/internal/models"
	"github.com/jose-lico/go-plate/middleware"
	"github.com/jose-lico/go-plate/ratelimiting"
	"github.com/jose-lico/go-plate/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Service struct {
	logger *zap.Logger
	store  UserStore
	redis  database.RedisStore
}

func NewService(logger *zap.Logger, store UserStore, redis database.RedisStore) *Service {
	return &Service{logger: logger, store: store, redis: redis}
}

func (s *Service) RegisterRoutes(v1 chi.Router) chi.Router {
	userRouter := chi.NewRouter()
	v1.Mount("/users", userRouter)

	userRouter.Group(func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(ratelimiting.NewInMemoryTokenBucket(0.05, 3, 10*time.Minute)))

		r.Post("/register", s.createUser)
		r.Post("/login", s.loginUser)
	})

	userRouter.Group(func(r chi.Router) {
		r.Use(middleware.SessionMiddleware(s.redis))

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

// @Summary Create a new user
// @Description Creates a new user by parsing the provided user data, validating it, and storing it in the database. Returns a cookie on successful creation.
// @Tags Users
// @Accept  json
// @Produce  json
// @Param user body RegisterUserPayload true "User data for registration"
// @Success 201 "User created successfully"
// @Header 201 {string} Set-Cookie "session=value; Path=/; HttpOnly"
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload"
// @Failure 409 {object} utils.ErrorResponse "User with provided email already exists"
// @Failure 500 {object} utils.ErrorResponse "Interal server error"
// @Router /v1/users/register [post]
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
		s.logger.Error("Error hashing password for createUser", zap.Error(err))
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
		return
	}

	u, err := s.store.CreateUser(&models.User{
		Email:    strings.ToLower(user.Email),
		Password: hashedPassword,
		Name:     user.Name,
	})

	if err != nil {
		s.logger.Error("Error creating user", zap.Error(err))
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
		return
	}

	s.generateSession(w, r, u, http.StatusCreated)
}

// @Summary Login user
// @Description Authenticates a user and creates a session
// @Tags Users
// @Accept json
// @Produce json
// @Param user body LoginUserPayload true "Login credentials"
// @Success 200 "Login successful"
// @Failure 400 {object} utils.ErrorResponse "Invalid request payload"
// @Failure 401 {object} utils.ErrorResponse "Invalid credentials"
// @Failure 500 {object} utils.ErrorResponse "Internal server error"
// @Router /v1/users/login [post]
func (s *Service) loginUser(w http.ResponseWriter, r *http.Request) {
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

	// These error messages could allow for user enumeration and should be more generic
	u, err := s.store.GetUserByEmail(user.Email)
	if err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			utils.WriteError(w, http.StatusUnauthorized, err)
			return
		default:
			s.logger.Error("Error getting user from store", zap.Error(err), zap.Any("User", user))
			utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
			return
		}
	}

	if !auth.ComparePasswords(u.Password, []byte(user.Password)) {
		utils.WriteError(w, http.StatusUnauthorized, errors.New("wrong password"))
		return
	}

	s.generateSession(w, r, u, http.StatusOK)
}

func (s *Service) generateSession(w http.ResponseWriter, r *http.Request, u *models.User, status int) {
	sessionToken, err := auth.GenerateToken()
	if err != nil {
		s.logger.Error("Error generating session token", zap.Error(err))
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
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
		s.logger.Error("Failed to marshall session", zap.Error(err))
		utils.WriteError(w, http.StatusInternalServerError, utils.ErrGenericInternalError)
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
