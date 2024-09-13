package post

import (
	"fmt"

	"go-plate/models"

	"gorm.io/gorm"
)

type PostStore interface {
	CreatePost(p *models.Post) (*models.Post, error)
	GetPosts(userId int, amount int) ([]models.Post, error)
}

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) PostStore {
	return &Store{db: db}
}

func (s *Store) CreatePost(post *models.Post) (*models.Post, error) {
	result := s.db.Create(post)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to create post: %w", result.Error)
	}

	return post, nil
}

func (s *Store) GetPosts(userId int, limit int) ([]models.Post, error) {
	var posts []models.Post

	err := s.db.Where("user_id = ?", userId).
		Order("created_at DESC").
		Limit(limit).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil
}
