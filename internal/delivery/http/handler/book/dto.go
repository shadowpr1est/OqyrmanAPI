package book

type createBookRequest struct {
	AuthorID    string `form:"author_id"   binding:"required"`
	GenreID     string `form:"genre_id"    binding:"required"`
	Title       string `form:"title"       binding:"required"`
	ISBN        string `form:"isbn"`
	CoverURL    string `form:"cover_url"`
	Description string `form:"description"`
	Language    string `form:"language"`
	Year        int    `form:"year"`
	// avg_rating убран — рейтинг вычисляется автоматически из отзывов пользователей.
	// При создании книги всегда 0.0, обновляется через review usecase.
}

type updateBookRequest struct {
	AuthorID    *string `json:"author_id"`
	GenreID     *string `json:"genre_id"`
	Title       *string `json:"title"`
	ISBN        *string `json:"isbn"`
	CoverURL    *string `json:"cover_url"`
	Description *string `json:"description"`
	Language    *string `json:"language"`
	Year        *int    `json:"year"`
	// avg_rating намеренно исключён — рейтинг вычисляется автоматически из отзывов
}

type bookResponse struct {
	ID          string  `json:"id"`
	AuthorID    string  `json:"author_id"`
	GenreID     string  `json:"genre_id"`
	Title       string  `json:"title"`
	ISBN        string  `json:"isbn"`
	CoverURL    string  `json:"cover_url"`
	Description string  `json:"description"`
	Language    string  `json:"language"`
	Year        int     `json:"year"`
	AvgRating   float64 `json:"avg_rating"`
}

type listBookResponse struct {
	Items  []*bookResponse `json:"items"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}
