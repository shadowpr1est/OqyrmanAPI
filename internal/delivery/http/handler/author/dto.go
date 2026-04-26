package author

type createAuthorRequest struct {
	Name      string  `json:"name"       binding:"required"`
	Bio       string  `json:"bio"`
	BioKK     string  `json:"bio_kk"`
	BirthDate *string `json:"birth_date"` // "2006-01-02"
	DeathDate *string `json:"death_date"` // "2006-01-02"
	PhotoURL  string  `json:"photo_url"`
}

type updateAuthorRequest struct {
	Name      *string `json:"name"`
	Bio       *string `json:"bio"`
	BioKK     *string `json:"bio_kk"`
	BirthDate *string `json:"birth_date"`
	DeathDate *string `json:"death_date"`
	PhotoURL  *string `json:"photo_url"`
}

type authorResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Bio       string  `json:"bio"`
	BioKK     string  `json:"bio_kk"`
	BirthDate *string `json:"birth_date"`
	DeathDate *string `json:"death_date"`
	PhotoURL  string  `json:"photo_url"`
}

type listAuthorResponse struct {
	Items  []*authorResponse `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
