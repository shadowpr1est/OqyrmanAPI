package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type UserUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context, limit, offset int) ([]*entity.User, int, error) // NEW
	UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role) error        // NEW
	AdminDelete(ctx context.Context, id uuid.UUID) error                         // NEW
}
