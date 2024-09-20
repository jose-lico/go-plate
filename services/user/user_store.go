package user

import (
	"errors"
	"fmt"

	"github.com/jose-lico/go-plate/models"

	"gorm.io/gorm"
)

type UserStore interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
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

func (s *Store) GetUserByID(id int) (*models.User, error) {
	var user models.User

	result := s.db.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, errors.New("failed to find user by id: " + result.Error.Error())
	}

	return &user, nil
}
