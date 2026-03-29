package common

// AuthorRef — lightweight author reference for nested API responses.
type AuthorRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GenreRef — lightweight genre reference for nested API responses.
type GenreRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LibraryRef — lightweight library reference for nested API responses.
type LibraryRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// UserRef — lightweight user reference for nested API responses.
type UserRef struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// BookRef — lightweight book reference for nested API responses.
// Fields are omitempty so each consumer populates only what it JOINs.
type BookRef struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	CoverURL   string    `json:"cover_url,omitempty"`
	Year       int       `json:"year,omitempty"`
	AvgRating  float64   `json:"avg_rating,omitempty"`
	TotalPages *int      `json:"total_pages,omitempty"`
	Author     AuthorRef `json:"author"`
	Genre      GenreRef  `json:"genre"`
}