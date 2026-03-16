package book_machine_book

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type bookMachineBookUseCase struct {
	machineBookRepo repository.BookMachineBookRepository
}

func NewBookMachineBookUseCase(machineBookRepo repository.BookMachineBookRepository) domainUseCase.BookMachineBookUseCase {
	return &bookMachineBookUseCase{machineBookRepo: machineBookRepo}
}

func (u *bookMachineBookUseCase) Create(ctx context.Context, mb *entity.BookMachineBook) (*entity.BookMachineBook, error) {
	mb.ID = uuid.New()
	return u.machineBookRepo.Create(ctx, mb)
}

func (u *bookMachineBookUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.BookMachineBook, error) {
	return u.machineBookRepo.GetByID(ctx, id)
}

func (u *bookMachineBookUseCase) ListByMachine(ctx context.Context, machineID uuid.UUID) ([]*entity.BookMachineBook, error) {
	return u.machineBookRepo.ListByMachine(ctx, machineID)
}

func (u *bookMachineBookUseCase) ListByBook(ctx context.Context, bookID uuid.UUID) ([]*entity.BookMachineBook, error) {
	return u.machineBookRepo.ListByBook(ctx, bookID)
}

func (u *bookMachineBookUseCase) Update(ctx context.Context, mb *entity.BookMachineBook) (*entity.BookMachineBook, error) {
	return u.machineBookRepo.Update(ctx, mb)
}

func (u *bookMachineBookUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.machineBookRepo.Delete(ctx, id)
}
