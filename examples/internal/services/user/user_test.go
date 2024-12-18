package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jose-lico/go-plate/examples/internal/models"

	"github.com/go-chi/chi/v5"
)

func TestUserService_CreateUser(t *testing.T) {
	type createUserTestCase struct {
		Name           string              `json:"name"`
		Payload        RegisterUserPayload `json:"payload"`
		ExpectedStatus int                 `json:"expectedStatus"`
		ExpectedBody   string              `json:"expectedBody"`
	}

	testFile, err := os.ReadFile("testdata/create_user.json")
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
	}

	var testData []createUserTestCase
	if err := json.Unmarshal(testFile, &testData); err != nil {
		t.Fatalf("Failed to parse test data: %v", err)
	}

	store := &MockUserStore{}
	cache := &MockCacheStore{}
	service := NewService(store, cache)

	for _, tc := range testData {
		t.Run(tc.Name, func(t *testing.T) {
			marshalled, err := json.Marshal(tc.Payload)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest(http.MethodPost, "/users/register", bytes.NewBuffer(marshalled))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := http.ServeMux{}
			router.HandleFunc("POST /users/register", service.createUser)

			router.ServeHTTP(rr, req)

			if rr.Code != tc.ExpectedStatus {
				t.Errorf("expected status code %d, got %d", tc.ExpectedStatus, rr.Code)
			}

			if tc.ExpectedBody != "" && rr.Body.String() != tc.ExpectedBody {
				t.Errorf("expected response body to contain %q, got %q", tc.ExpectedBody, rr.Body.String())
			}
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	store := &MockUserStore{}
	cache := &MockCacheStore{}
	service := NewService(store, cache)

	loginUserTests := []struct {
		name           string
		payload        LoginUserPayload
		expectedStatus int
		expectedBody   string // Optional: to validate response content
	}{
		{"No Email", LoginUserPayload{Password: "MyPassword"}, http.StatusBadRequest, "email is required"},
		{"No Password", LoginUserPayload{Email: "example@email.com"}, http.StatusBadRequest, "password is required"},
		{"Invalid Email", LoginUserPayload{Email: "notanemail", Password: "MyPassword"}, http.StatusBadRequest, "invalid credentials"},
		{"User Not Found", LoginUserPayload{Email: "notfound@email.com", Password: "MyPassword"}, http.StatusUnauthorized, "invalid credentials"},
		{"Invalid Password", LoginUserPayload{Email: "example@email.com", Password: "WrongPassword"}, http.StatusUnauthorized, "invalid credentials"},
		{"Successful Login", LoginUserPayload{Email: "example@email.com", Password: "MyPassword"}, http.StatusOK, ""},
	}

	for _, tc := range loginUserTests {
		t.Run(tc.name, func(t *testing.T) {
			marshalled, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(marshalled))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router := chi.NewRouter()

			router.Post("/users/login", service.loginUser)
			router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				t.Errorf("expected status code %d, got %d", tc.expectedStatus, rr.Code)
			}

			// if tc.expectedBody != "" && rr.Body.String() != tc.expectedBody {
			// 	t.Errorf("expected response body to contain %q, got %q", tc.expectedBody, rr.Body.String())
			// }
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
		u.Password = "$2a$10$ZTN4HGWy6QeenPN2X1Kxfe32u/6kmlI37ndvh0raGFjBeYTroHt/m"
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
