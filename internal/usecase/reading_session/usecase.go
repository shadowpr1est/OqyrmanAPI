package reading_session

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type readingSessionUseCase struct {
	sessionRepo repository.ReadingSessionRepository
	bookRepo    repository.BookRepository
}

func NewReadingSessionUseCase(
	sessionRepo repository.ReadingSessionRepository,
	bookRepo repository.BookRepository,
) domainUseCase.ReadingSessionUseCase {
	return &readingSessionUseCase{sessionRepo: sessionRepo, bookRepo: bookRepo}
}

func (u *readingSessionUseCase) Upsert(ctx context.Context, session *entity.ReadingSession, totalPages *int) (*entity.ReadingSession, error) {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	session.UpdatedAt = time.Now()

	// Protect against false "finished" when total is unknown (totalPages not provided or 0).
	if session.Status == entity.StatusFinished {
		if totalPages == nil || *totalPages == 0 {
			session.Status = entity.StatusReading
		} else if session.FinishedAt == nil {
			now := time.Now()
			session.FinishedAt = &now
		}
	}

	result, err := u.sessionRepo.Upsert(ctx, session)
	if err != nil {
		return nil, err
	}

	// If the client reported total_pages and the book doesn't have it yet, update books.total_pages.
	if totalPages != nil && *totalPages > 0 {
		if err := u.bookRepo.UpdateTotalPages(ctx, session.BookID, *totalPages); err != nil {
			slog.WarnContext(ctx, "failed to update total_pages from reader", "book_id", session.BookID, "err", err)
		}
	}

	return result, nil
}

func (u *readingSessionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingSession, error) {
	return u.sessionRepo.GetByID(ctx, id)
}

func (u *readingSessionUseCase) GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSession, error) {
	return u.sessionRepo.GetByUserAndBook(ctx, userID, bookID)
}

func (u *readingSessionUseCase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSession, error) {
	return u.sessionRepo.ListByUser(ctx, userID)
}

func (u *readingSessionUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.sessionRepo.Delete(ctx, id)
}

func (u *readingSessionUseCase) GetByUserAndBookView(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSessionView, error) {
	return u.sessionRepo.GetByUserAndBookView(ctx, userID, bookID)
}

func (u *readingSessionUseCase) ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSessionView, error) {
	return u.sessionRepo.ListByUserView(ctx, userID)
}
