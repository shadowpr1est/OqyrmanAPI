package library_book

type createLibraryBookRequest struct {
	LibraryID       string `json:"library_id"        binding:"required"`
	BookID          string `json:"book_id"           binding:"required"`
	TotalCopies     int    `json:"total_copies"      binding:"required"`
	AvailableCopies int    `json:"available_copies"  binding:"required"`
}

type updateLibraryBookRequest struct {
	TotalCopies     *int `json:"total_copies"`
	AvailableCopies *int `json:"available_copies"`
}

type libraryBookResponse struct {
	ID              string `json:"id"`
	LibraryID       string `json:"library_id"`
	BookID          string `json:"book_id"`
	TotalCopies     int    `json:"total_copies"`
	AvailableCopies int    `json:"available_copies"`
}

type libraryBookSearchItem struct {
	LibraryBookID   string  `json:"library_book_id"`
	BookID          string  `json:"book_id"`
	Title           string  `json:"title"`
	Author          string  `json:"author"`
	Genre           string  `json:"genre"`
	CoverURL        *string `json:"cover_url"`
	Year            *int    `json:"year"`
	TotalCopies     int     `json:"total_copies"`
	AvailableCopies int     `json:"available_copies"`
	IsAvailable     bool    `json:"is_available"`
}

type libraryBookSearchResponse struct {
	Items  []*libraryBookSearchItem `json:"items"`
	Total  int                      `json:"total"`
	Limit  int                      `json:"limit"`
	Offset int                      `json:"offset"`
}
