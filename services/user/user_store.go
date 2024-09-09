package user

import (
	"errors"
	"fmt"

	"go-plate/models"

	"gorm.io/gorm"
)

type UserStore interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) UserStore {
	return &Store{db: db}
}

func (s *Store) CreateUser(user *models.User) (*models.User, error) {
	result := s.db.Create(user)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	return user, nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	var user models.User

	result := s.db.Where("LOWER(email) = LOWER(?)", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user does not exist: %w", result.Error)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", result.Error)
	}

	return &user, nil
}
