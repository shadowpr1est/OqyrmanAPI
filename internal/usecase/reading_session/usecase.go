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

func NewReadingSessionUseCase(sessionRepo repository.ReadingSessionRepository) domainUseCase.ReadingSessionUseCase {
	return &readingSessionUseCase{sessionRepo: sessionRepo}
}

func (u *readingSessionUseCase) Upsert(ctx context.Context, session *entity.ReadingSession) (*entity.ReadingSession, error) {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	session.UpdatedAt = time.Now()
	if session.Status == entity.StatusFinished && session.FinishedAt == nil {
		now := time.Now()
		session.FinishedAt = &now
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
