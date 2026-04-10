package notes

type createNoteRequest struct {
	BookID   string `json:"book_id"  binding:"required"`
	Position string `json:"position" binding:"required"`
	Content  string `json:"content"  binding:"required"`
}

type updateNoteRequest struct {
	Position *string `json:"position"`
	Content  *string `json:"content"`
}

type noteResponse struct {
	ID        string `json:"id"`
	BookID    string `json:"book_id"`
	Position  string `json:"position"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type noteViewResponse struct {
	ID        string `json:"id"`
	BookID    string `json:"book_id"`
	BookTitle string `json:"book_title"`
	Position  string `json:"position"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
