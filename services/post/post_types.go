package post

type PostPayload struct {
	Title   string `json:"title" validate:"required,max=255"`
	Summary string `json:"summary"` // Only v2 has a summary
	Content string `json:"content" validate:"required,max=1000"`
}
