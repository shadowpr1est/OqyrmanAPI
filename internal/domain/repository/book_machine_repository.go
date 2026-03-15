package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type BookMachineRepository interface {
	Create(ctx context.Context, machine *entity.BookMachine) (*entity.BookMachine, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.BookMachine, error)
	List(ctx context.Context, limit, offset int) ([]*entity.BookMachine, int, error)
	ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.BookMachine, error)
	Update(ctx context.Context, machine *entity.BookMachine) (*entity.BookMachine, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
