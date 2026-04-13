package reading_session

type upsertReadingSessionRequest struct {
	BookID      string  `json:"book_id"       binding:"required"`
	Progress    int     `json:"progress"      binding:"min=0,max=100"`
	CfiPosition *string `json:"cfi_position"`
	Status      string  `json:"status"        binding:"required,oneof=reading finished dropped"`
}

type readingSessionResponse struct {
	ID          string  `json:"id"`
	BookID      string  `json:"book_id"`
	Progress    int     `json:"progress"`
	CfiPosition *string `json:"cfi_position,omitempty"`
	Status      string  `json:"status"`
	UpdatedAt   string  `json:"updated_at"`
	FinishedAt  *string `json:"finished_at,omitempty"`
}

type sessionBookRef struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	CoverURL   string `json:"cover_url,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
}

type readingSessionViewResponse struct {
	ID          string         `json:"id"`
	Progress    int            `json:"progress"`
	CfiPosition *string        `json:"cfi_position,omitempty"`
	Status      string         `json:"status"`
	UpdatedAt   string         `json:"updated_at"`
	FinishedAt  *string        `json:"finished_at,omitempty"`
	Book        sessionBookRef `json:"book"`
}