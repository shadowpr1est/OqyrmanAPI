package genre

type createGenreRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type updateGenreRequest struct {
	Name *string `json:"name"`
	Slug *string `json:"slug"`
}

type genreResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
