package event

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type eventUseCase struct {
	repo repository.EventRepository
}

func NewEventUseCase(repo repository.EventRepository) domainUseCase.EventUseCase {
	return &eventUseCase{repo: repo}
}

func (u *eventUseCase) List(ctx context.Context, limit, offset int) ([]*entity.Event, int, error) {
	return u.repo.List(ctx, limit, offset)
}

func (u *eventUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.Event, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *eventUseCase) Create(ctx context.Context, e *entity.Event) (*entity.Event, error) {
	e.ID = uuid.New()
	e.CreatedAt = time.Now()
	return u.repo.Create(ctx, e)
}

func (u *eventUseCase) Update(ctx context.Context, e *entity.Event) (*entity.Event, error) {
	return u.repo.Update(ctx, e)
}

func (u *eventUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.repo.Delete(ctx, id)
}
