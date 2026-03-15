package entity

import "github.com/google/uuid"

type BookFile struct {
	ID      uuid.UUID `db:"id"`
	BookID  uuid.UUID `db:"book_id"`
	Format  string    `db:"format"`
	FileURL string    `db:"file_url"`
	IsAudio bool      `db:"is_audio"`
}
