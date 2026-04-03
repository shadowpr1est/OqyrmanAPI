package notes

type createNoteRequest struct {
	BookID  string `json:"book_id" binding:"required"`
	Page    int    `json:"page"    binding:"required"`
	Content string `json:"content" binding:"required"`
}

type updateNoteRequest struct {
	Page    *int    `json:"page"`
	Content *string `json:"content"`
}

type noteResponse struct {
	ID        string `json:"id"`
	BookID    string `json:"book_id"`
	Page      int    `json:"page"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type noteViewResponse struct {
	ID        string `json:"id"`
	BookID    string `json:"book_id"`
	BookTitle string `json:"book_title"`
	Page      int    `json:"page"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}
