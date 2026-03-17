package stats

import (
	"context"

	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type statsUseCase struct {
	repo repository.StatsRepository
}

func NewStatsUseCase(repo repository.StatsRepository) domainUseCase.StatsUseCase {
	return &statsUseCase{repo: repo}
}

func (u *statsUseCase) GetStats(ctx context.Context) (*entity.Stats, error) {
	return u.repo.GetStats(ctx)
}
