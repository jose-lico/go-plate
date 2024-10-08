package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jose-lico/go-plate/examples/internal/models"
)

func TestUserService(t *testing.T) {
	store := &MockUserStore{}
	cache := &MockCacheStore{}

	service := NewService(store, cache)

	createUserTests := []struct {
		name           string
		payload        RegisterUserPayload
		expectedStatus int
	}{
		{"No Email", RegisterUserPayload{Password: "MyPassword", Name: "José"}, http.StatusBadRequest},
		{"Invalid email", RegisterUserPayload{Email: "abcddas", Password: "MyPassword", Name: "José"}, http.StatusBadRequest},
		{"Email in use", RegisterUserPayload{Email: "example@email.com", Password: "MyPassword", Name: "José"}, http.StatusConflict},
		{"No Password", RegisterUserPayload{Password: "MyPass", Name: "José"}, http.StatusBadRequest},
		{"Short Password", RegisterUserPayload{Email: "example2@email.com", Password: "MyPass", Name: "José"}, http.StatusBadRequest},
		{"Long Password", RegisterUserPayload{Email: "example2@email.com", Password: "MyPassworddddddddddddd", Name: "José"}, http.StatusBadRequest},
		{"No Name", RegisterUserPayload{Email: "example2@email.com", Password: "MyPassword"}, http.StatusBadRequest},
		{"Name too short", RegisterUserPayload{Email: "example2@email.com", Password: "MyPassword", Name: "J"}, http.StatusBadRequest},
		{"Name too long", RegisterUserPayload{Email: "example2@email.com", Password: "MyPassword", Name: "JYfUKncJrcXXYkGOqJHUPyTaifKmDbIQE"}, http.StatusBadRequest},
		{"Create User", RegisterUserPayload{Email: "example2@email.com", Password: "MyPassword", Name: "José"}, http.StatusCreated},
	}

	for _, tc := range createUserTests {
		t.Run(tc.name, func(t *testing.T) {
			marshalled, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(marshalled))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Post("/users", service.createUser)

			router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}
		})
	}
}

type MockUserStore struct{}

func (s *MockUserStore) CreateUser(user *models.User) (*models.User, error) {
	u := &models.User{}
	u.ID = 1
	return u, nil
}

func (s *MockUserStore) GetUserByEmail(email string) (*models.User, error) {
	if email == "example@email.com" {
		u := &models.User{}
		u.ID = 1
		return u, nil
	}

	return nil, errors.New("user not found")
}

func (s *MockUserStore) GetUserByID(id int) (*models.User, error) {
	return nil, nil
}

type MockCacheStore struct{}

func (s *MockCacheStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return nil
}
func (s *MockCacheStore) Get(ctx context.Context, key string) (string, error)           { return "", nil }
func (s *MockCacheStore) SAdd(ctx context.Context, key string, value interface{}) error { return nil }
func (s *MockCacheStore) SRem(ctx context.Context, key string, value interface{}) error { return nil }
func (s *MockCacheStore) Del(ctx context.Context, key string) (int64, error)            { return 1, nil }
func (s *MockCacheStore) GetNativeInstance() interface{}                                { return nil }
