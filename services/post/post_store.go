package post

import (
	"errors"
	"fmt"

	"github.com/jose-lico/go-plate/models"

	"gorm.io/gorm"
)

var (
	ErrPostNotFound      = errors.New("post not found")
	ErrUserNotAuthorized = errors.New("user not authorized to delete/edit post")
)

type PostStore interface {
	CreatePost(p *models.Post) (*models.Post, error)
	GetPostByID(postID int) (*models.Post, error)
	GetPosts(userId int, amount int) ([]models.Post, error)
	UpdatePost(post *models.Post, updates interface{}) error
	DeletePost(postID, userID int) error
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

func (s *Store) GetPostByID(postID int) (*models.Post, error) {
	var post models.Post

	result := s.db.Where("id = ?", postID).First(&post)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrPostNotFound
	}
	if result.Error != nil {
		return nil, fmt.Errorf("error fetching post: %w", result.Error)
	}

	return &post, nil
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

func (s *Store) UpdatePost(post *models.Post, updates interface{}) error {
	result := s.db.Model(&post).Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update post: %w", result.Error)
	}

	return nil
}

func (s *Store) DeletePost(postID, userID int) error {
	post, err := s.GetPostByID(postID)
	if err != nil {
		return err
	}

	if post.UserID != uint(userID) {
		return ErrUserNotAuthorized
	}

	result := s.db.Delete(post)
	if result.Error != nil {
		return fmt.Errorf("error deleting post: %w", result.Error)
	}

	return nil
}
