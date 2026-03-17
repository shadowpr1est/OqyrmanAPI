package usecase

import (
	"context"

	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type StatsUseCase interface {
	GetStats(ctx context.Context) (*entity.Stats, error)
}
