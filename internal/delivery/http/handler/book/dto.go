package book

type createBookRequest struct {
	AuthorID    string  `json:"author_id"   binding:"required"`
	GenreID     string  `json:"genre_id"    binding:"required"`
	Title       string  `json:"title"       binding:"required"`
	ISBN        string  `json:"isbn"`
	CoverURL    string  `json:"cover_url"`
	Description string  `json:"description"`
	Language    string  `json:"language"`
	Year        int     `json:"year"`
	AvgRating   float64 `json:"avg_rating"`
}

type updateBookRequest struct {
	AuthorID    *string  `json:"author_id"`
	GenreID     *string  `json:"genre_id"`
	Title       *string  `json:"title"`
	ISBN        *string  `json:"isbn"`
	CoverURL    *string  `json:"cover_url"`
	Description *string  `json:"description"`
	Language    *string  `json:"language"`
	Year        *int     `json:"year"`
	AvgRating   *float64 `json:"avg_rating"`
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
