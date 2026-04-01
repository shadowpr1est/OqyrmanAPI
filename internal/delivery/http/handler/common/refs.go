package common

import "github.com/google/uuid"
import "github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"

// AuthorRef — lightweight author reference for nested API responses.
type AuthorRef struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Bio       string  `json:"bio"`
	BirthDate *string `json:"birth_date"` // "2006-01-02"
	DeathDate *string `json:"death_date"` // "2006-01-02"
	PhotoURL  string  `json:"photo_url"`
}

type BookFileRef struct {
	ID      uuid.UUID             `json:"id"`
	BookID  uuid.UUID             `json:"book_id"`
	Format  entity.BookFileFormat `json:"format"`
	FileURL string                `json:"file_url"`
}

// GenreRef — lightweight genre reference for nested API responses.
type GenreRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// LibraryRef — lightweight library reference for nested API responses.
type LibraryRef struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Address string  `json:"address" binding:"required"`
	Lat     float64 `json:"lat"     binding:"required"`
	Lng     float64 `json:"lng"     binding:"required"`
	Phone   string  `json:"phone"`
}

// UserRef — lightweight user reference for nested API responses.
type UserRef struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
	Role      string `json:"role"`
	QRCode    string `json:"qr_code"`
	CreatedAt string `json:"created_at"`
}

// BookRef — lightweight book reference for nested API responses.
// Fields are omitempty so each consumer populates only what it JOINs.
type BookRef struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	ISBN        string    `json:"isbn"`
	CoverURL    string    `json:"cover_url"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	Year        int       `json:"year"`
	TotalPages  *int      `json:"total_pages"`
	AvgRating   float64   `json:"avg_rating"`
	Author      AuthorRef `json:"author"`
	Genre       GenreRef  `json:"genre"`
}
