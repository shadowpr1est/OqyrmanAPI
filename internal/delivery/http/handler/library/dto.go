package library

type createLibraryRequest struct {
	Name     string  `json:"name"    binding:"required"`
	Address  string  `json:"address" binding:"required"`
	Lat      float64 `json:"lat"     binding:"required"`
	Lng      float64 `json:"lng"     binding:"required"`
	Phone    string  `json:"phone"`
	PhotoURL string  `json:"photo_url"`
}

type updateLibraryRequest struct {
	Name     *string  `json:"name"`
	Address  *string  `json:"address"`
	Lat      *float64 `json:"lat"`
	Lng      *float64 `json:"lng"`
	Phone    *string  `json:"phone"`
	PhotoURL *string  `json:"photo_url"`
}

type libraryResponse struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Phone    string  `json:"phone"`
	PhotoURL string  `json:"photo_url"`
}

type listLibraryResponse struct {
	Items  []*libraryResponse `json:"items"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}
