package user

import (
	"fmt"

	"go-plate/models"

	"gorm.io/gorm"
)

type UserStore interface {
	CreateUser(user *models.User) (*models.User, error)
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
