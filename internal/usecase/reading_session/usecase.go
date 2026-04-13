package reading_session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type readingSessionUseCase struct {
	sessionRepo repository.ReadingSessionRepository
}

func NewReadingSessionUseCase(
	sessionRepo repository.ReadingSessionRepository,
) domainUseCase.ReadingSessionUseCase {
	return &readingSessionUseCase{sessionRepo: sessionRepo}
}

func (u *readingSessionUseCase) Upsert(ctx context.Context, session *entity.ReadingSession) (*entity.ReadingSession, error) {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	session.UpdatedAt = time.Now()

	// Clamp progress to 0–100
	if session.Progress < 0 {
		session.Progress = 0
	}
	if session.Progress > 100 {
		session.Progress = 100
	}

	// Protect against false "finished" when progress is not 100%.
	if session.Status == entity.StatusFinished {
		if session.Progress < 100 {
			session.Status = entity.StatusReading
		} else if session.FinishedAt == nil {
			now := time.Now()
			session.FinishedAt = &now
		}
	}

	return u.sessionRepo.Upsert(ctx, session)
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
