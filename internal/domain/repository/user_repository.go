package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context, limit, offset int) ([]*entity.User, int, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role, libraryID *uuid.UUID) error
	UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error

	// View method — used by admin GET /users; returns joined library data.
	ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error)
}
