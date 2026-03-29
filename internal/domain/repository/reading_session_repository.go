package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ReadingSessionRepository interface {
	Upsert(ctx context.Context, session *entity.ReadingSession) (*entity.ReadingSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ReadingSession, error)
	GetByUserAndBook(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSession, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSession, error)
	Delete(ctx context.Context, id uuid.UUID) error

	// View methods — used by GET endpoints; return joined book/author data.
	GetByUserAndBookView(ctx context.Context, userID, bookID uuid.UUID) (*entity.ReadingSessionView, error)
	ListByUserView(ctx context.Context, userID uuid.UUID) ([]*entity.ReadingSessionView, error)
}
