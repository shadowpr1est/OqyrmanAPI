package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type StatsRepository interface {
	GetStats(ctx context.Context) (*entity.Stats, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*entity.UserStats, error)
}
