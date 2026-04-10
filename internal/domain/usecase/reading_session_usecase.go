package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReadingSessionUseCase interface {
	Upsert(ctx context.Context, session *entity.ReadingSession, totalPages *int) (*entity.ReadingSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingSession, error)
	GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSession, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSession, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// View methods — return enriched nested data for GET endpoints.
	GetByUserAndBookView(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSessionView, error)
	ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSessionView, error)
}
