package repository

import (
	"context"

	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type StatsRepository interface {
	GetStats(ctx context.Context) (*entity.Stats, error)
}
