package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type userUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUseCase(userRepo repository.UserRepository) domainUseCase.UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

func (u *userUseCase) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUseCase) Update(ctx context.Context, user *entity.User) (*entity.User, error) {
	return u.userRepo.Update(ctx, user)
}

func (u *userUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}

func (u *userUseCase) ListAll(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	return u.userRepo.ListAll(ctx, limit, offset)
}

func (u *userUseCase) UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role) error {
	return u.userRepo.UpdateRole(ctx, id, role)
}

func (u *userUseCase) AdminDelete(ctx context.Context, id uuid.UUID) error {
	return u.userRepo.Delete(ctx, id)
}
