package post

import (
	"time"

	"github.com/jose-lico/go-plate/examples/internal/models"
)

type PostPayload struct {
	Title   string `json:"title" validate:"required,max=255"`
	Summary string `json:"summary"` // Only v2 has a summary
	Content string `json:"content" validate:"required,max=1000"`
}

type PostResponsePayload struct {
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type EditPostPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
}

func ModelToResponsePayload(p *models.Post) PostResponsePayload {
	return PostResponsePayload{
		Title:     p.Title,
		Summary:   p.Summary,
		Content:   p.Content,
		CreatedAt: p.CreatedAt,
	}
}
