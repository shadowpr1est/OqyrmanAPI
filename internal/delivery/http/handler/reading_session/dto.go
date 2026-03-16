package reading_session

type upsertReadingSessionRequest struct {
	BookID      string `json:"book_id"      binding:"required"`
	CurrentPage int    `json:"current_page" binding:"required"`
	Status      string `json:"status"       binding:"required,oneof=reading finished dropped"`
}

type readingSessionResponse struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	BookID      string `json:"book_id"`
	CurrentPage int    `json:"current_page"`
	Status      string `json:"status"`
	UpdatedAt   string `json:"updated_at"`
}
