package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
	_ "github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload" // keep for fileupload.File
)

type UserUseCase interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) (*entity.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AdminUpdateUser(ctx context.Context, id uuid.UUID, role *entity.Role, libraryID *uuid.UUID, name, surname, email, phone *string) (*entity.UserView, error)
	AdminDelete(ctx context.Context, id uuid.UUID) error
	UploadAvatar(ctx context.Context, id uuid.UUID, avatar *fileupload.File) (*entity.User, error)
	CreateStaff(ctx context.Context, email, password, name, surname, phone string, libraryID uuid.UUID) (*entity.UserView, error)
	// View method — returns enriched nested data for admin GET /users.
	ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error)
	// Session management
	ListSessions(ctx context.Context, userID uuid.UUID) ([]*entity.Token, error)
	RevokeSession(ctx context.Context, sessionID, userID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
	// ChangePassword verifies oldPassword and sets newPassword.
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}
