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
	AdminUpdate(ctx context.Context, id uuid.UUID, role *entity.Role, libraryID *uuid.UUID, name, surname, email, phone *string) (*entity.UserView, error)
	UpdateAvatarURL(ctx context.Context, id uuid.UUID, avatarURL string) error
	SetEmailVerified(ctx context.Context, id uuid.UUID) error
	GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error)
	SetGoogleID(ctx context.Context, id uuid.UUID, googleID string) error

	// View method — used by admin GET /users; returns joined library data.
	ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error)
	// UpdatePassword replaces the password hash for the given user.
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	// GetByPhone returns a user by phone number.
	GetByPhone(ctx context.Context, phone string) (*entity.User, error)
	// HardDelete permanently removes a user row (used to release unique constraints on re-registration).
	HardDelete(ctx context.Context, id uuid.UUID) error
}
