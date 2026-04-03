package review

type createReviewRequest struct {
	BookID string `json:"book_id" binding:"required"`
	Rating int    `json:"rating"  binding:"required,min=1,max=5"`
	Body   string `json:"body"    binding:"required"`
}

type updateReviewRequest struct {
	Rating *int    `json:"rating" binding:"omitempty,min=1,max=5"`
	Body   *string `json:"body"`
}

type reviewResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	BookID    string `json:"book_id"`
	Rating    int    `json:"rating"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
}

type listReviewResponse struct {
	Items  []*reviewResponse `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

type reviewViewResponse struct {
	ID            string `json:"id"`
	BookID        string `json:"book_id"`
	BookTitle     string `json:"book_title"`
	UserID        string `json:"user_id"`
	UserName      string `json:"user_name"`
	UserSurname   string `json:"user_surname"`
	UserAvatarURL string `json:"user_avatar_url,omitempty"`
	Rating        int    `json:"rating"`
	Body          string `json:"body"`
	CreatedAt     string `json:"created_at"`
}

type listReviewViewResponse struct {
	Items  []*reviewViewResponse `json:"items"`
	Total  int                   `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}
