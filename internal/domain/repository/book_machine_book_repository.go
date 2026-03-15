package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type BookMachineBookRepository interface {
	Create(ctx context.Context, mb *entity.BookMachineBook) (*entity.BookMachineBook, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.BookMachineBook, error)
	ListByMachine(ctx context.Context, machineID uuid.UUID) ([]*entity.BookMachineBook, error)
	ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookMachineBook, error)
	Update(ctx context.Context, mb *entity.BookMachineBook) (*entity.BookMachineBook, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
