package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReadingNoteRepository interface {
	Create(ctx context.Context, note *entity.ReadingNote) (*entity.ReadingNote, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingNote, error)
	ListByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) ([]*entity.ReadingNote, error)
	Update(ctx context.Context, note *entity.ReadingNote) (*entity.ReadingNote, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
