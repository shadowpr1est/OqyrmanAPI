package book_machine

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type bookMachineUseCase struct {
	machineRepo repository.BookMachineRepository
}

func NewBookMachineUseCase(machineRepo repository.BookMachineRepository) domainUseCase.BookMachineUseCase {
	return &bookMachineUseCase{machineRepo: machineRepo}
}

func (u *bookMachineUseCase) Create(ctx context.Context, machine *entity.BookMachine) (*entity.BookMachine, error) {
	machine.ID = uuid.New()
	return u.machineRepo.Create(ctx, machine)
}

func (u *bookMachineUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.BookMachine, error) {
	return u.machineRepo.GetByID(ctx, id)
}

func (u *bookMachineUseCase) List(ctx context.Context, limit, offset int) ([]*entity.BookMachine, int, error) {
	return u.machineRepo.List(ctx, limit, offset)
}

func (u *bookMachineUseCase) ListNearby(ctx context.Context, lat, lng, radiusKm float64) ([]*entity.BookMachine, error) {
	return u.machineRepo.ListNearby(ctx, lat, lng, radiusKm)
}

func (u *bookMachineUseCase) Update(ctx context.Context, machine *entity.BookMachine) (*entity.BookMachine, error) {
	return u.machineRepo.Update(ctx, machine)
}

func (u *bookMachineUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.machineRepo.Delete(ctx, id)
}
