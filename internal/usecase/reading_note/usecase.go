package reading_note

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type readingNoteUseCase struct {
	noteRepo repository.ReadingNoteRepository
}

func NewReadingNoteUseCase(noteRepo repository.ReadingNoteRepository) domainUseCase.ReadingNoteUseCase {
	return &readingNoteUseCase{noteRepo: noteRepo}
}

func (u *readingNoteUseCase) Create(ctx context.Context, note *entity.ReadingNote) (*entity.ReadingNote, error) {
	note.ID = uuid.New()
	now := time.Now()
	note.CreatedAt = now
	note.UpdatedAt = now
	return u.noteRepo.Create(ctx, note)
}

func (u *readingNoteUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingNote, error) {
	return u.noteRepo.GetByID(ctx, id)
}

func (u *readingNoteUseCase) ListByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) ([]*entity.ReadingNote, error) {
	return u.noteRepo.ListByUserAndBook(ctx, userID, bookID)
}

func (u *readingNoteUseCase) Update(ctx context.Context, note *entity.ReadingNote) (*entity.ReadingNote, error) {
	note.UpdatedAt = time.Now()
	return u.noteRepo.Update(ctx, note)
}

func (u *readingNoteUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.noteRepo.Delete(ctx, id)
}

func (u *readingNoteUseCase) GetByIDView(ctx context.Context, id uuid.UUID) (*entity.ReadingNoteView, error) {
	return u.noteRepo.GetByIDView(ctx, id)
}

func (u *readingNoteUseCase) ListByUserAndBookView(ctx context.Context, userID, bookID uuid.UUID) ([]*entity.ReadingNoteView, error) {
	return u.noteRepo.ListByUserAndBookView(ctx, userID, bookID)
}
