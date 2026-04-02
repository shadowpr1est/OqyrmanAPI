package common

import "github.com/google/uuid"
import "github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"

// AuthorRef — lightweight author reference for nested API responses.
type AuthorRef struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Bio       string  `json:"bio,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"` // "2006-01-02"
	DeathDate *string `json:"death_date,omitempty"` // "2006-01-02"
	PhotoURL  string  `json:"photo_url,omitempty"`
}

type BookFileRef struct {
	ID      uuid.UUID             `json:"id"`
	BookID  uuid.UUID             `json:"book_id,omitempty"`
	Format  entity.BookFileFormat `json:"format,omitempty"`
	FileURL string                `json:"file_url,omitempty"`
}

// GenreRef — lightweight genre reference for nested API responses.
type GenreRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

// LibraryRef — lightweight library reference for nested API responses.
type LibraryRef struct {
	ID      string  `json:"id"`
	Name    string  `json:"name,omitempty"`
	Address string  `json:"address,omitempty" binding:"required"`
	Lat     float64 `json:"lat,omitempty"     binding:"required"`
	Lng     float64 `json:"lng,omitempty"     binding:"required"`
	Phone   string  `json:"phone,omitempty"`
}

// UserRef — lightweight user reference for nested API responses.
type UserRef struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Name      string `json:"name"`
	Surname   string `json:"surname"`
	AvatarURL string `json:"avatar_url,omitempty"`
	Role      string `json:"role,omitempty,omitempty"`
	QRCode    string `json:"qr_code,omitempty,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
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

type ReservationBookRef struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	CoverURL string `json:"cover_url,omitempty"`
}
