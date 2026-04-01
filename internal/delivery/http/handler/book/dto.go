package book

import "github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"

type createBookRequest struct {
	AuthorID    string `form:"author_id"   binding:"required"`
	GenreID     string `form:"genre_id"    binding:"required"`
	Title       string `form:"title"       binding:"required"`
	ISBN        string `form:"isbn"`
	CoverURL    string `form:"cover_url"`
	Description string `form:"description"`
	Language    string `form:"language"`
	Year        int    `form:"year"`
	TotalPages  *int   `form:"total_pages"`
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
	TotalPages  *int    `json:"total_pages"`
	// avg_rating намеренно исключён — рейтинг вычисляется автоматически из отзывов
}

type bookResponse struct {
	ID          string             `json:"id"`
	AuthorID    string             `json:"author_id"`
	GenreID     string             `json:"genre_id"`
	Title       string             `json:"title"`
	ISBN        string             `json:"isbn"`
	CoverURL    string             `json:"cover_url"`
	BookFile    common.BookFileRef `json:"book_file"`
	Description string             `json:"description"`
	Language    string             `json:"language"`
	Year        int                `json:"year"`
	AvgRating   float64            `json:"avg_rating"`
	TotalPages  *int               `json:"total_pages,omitempty"`
}

type listBookResponse struct {
	Items  []*bookResponse `json:"items"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type bookViewResponse struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	ISBN        string             `json:"isbn"`
	CoverURL    string             `json:"cover_url,omitempty"`
	Description string             `json:"description"`
	Language    string             `json:"language"`
	Year        int                `json:"year,omitempty"`
	AvgRating   float64            `json:"avg_rating"`
	TotalPages  *int               `json:"total_pages,omitempty"`
	BookFile    *common.BookFileRef `json:"file,omitempty"`
	Author      common.AuthorRef    `json:"author"`
	Genre       common.GenreRef     `json:"genre"`
}

type listBookViewResponse struct {
	Items  []*bookViewResponse `json:"items"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}
