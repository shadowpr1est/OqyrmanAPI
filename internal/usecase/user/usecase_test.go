package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/usecase/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ─────────────────────────────────────────────────────────────────────

type mockUserRepo struct{ mock.Mock }

func (m *mockUserRepo) Create(ctx context.Context, u *entity.User) (*entity.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, u *entity.User) (*entity.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mockUserRepo) ListAll(ctx context.Context, limit, offset int) ([]*entity.User, int, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*entity.User), args.Int(1), args.Error(2)
}
func (m *mockUserRepo) UpdateRole(ctx context.Context, id uuid.UUID, role entity.Role, libraryID *uuid.UUID) error {
	return m.Called(ctx, id, role, libraryID).Error(0)
}
func (m *mockUserRepo) UpdateAvatarURL(ctx context.Context, id uuid.UUID, url string) error {
	return m.Called(ctx, id, url).Error(0)
}

// ─── UpdateRole ───────────────────────────────────────────────────────────────

func TestUpdateRole_StaffWithLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	libID := uuid.New()
	repo.On("UpdateRole", mock.Anything, id, entity.RoleStaff, &libID).Return(nil)

	err := uc.UpdateRole(context.Background(), id, entity.RoleStaff, &libID)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateRole_StaffWithoutLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	err := uc.UpdateRole(context.Background(), uuid.New(), entity.RoleStaff, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "library_id is required for Staff")
	repo.AssertNotCalled(t, "UpdateRole")
}

func TestUpdateRole_AdminWithoutLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	repo.On("UpdateRole", mock.Anything, id, entity.RoleAdmin, (*uuid.UUID)(nil)).Return(nil)

	err := uc.UpdateRole(context.Background(), id, entity.RoleAdmin, nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateRole_AdminWithLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	libID := uuid.New()
	err := uc.UpdateRole(context.Background(), uuid.New(), entity.RoleAdmin, &libID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "library_id must be null for non-Staff")
	repo.AssertNotCalled(t, "UpdateRole")
}

func TestUpdateRole_UserWithLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	libID := uuid.New()
	err := uc.UpdateRole(context.Background(), uuid.New(), entity.RoleUser, &libID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "library_id must be null for non-Staff")
	repo.AssertNotCalled(t, "UpdateRole")
}

func TestUpdateRole_UserWithoutLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	repo.On("UpdateRole", mock.Anything, id, entity.RoleUser, (*uuid.UUID)(nil)).Return(nil)

	err := uc.UpdateRole(context.Background(), id, entity.RoleUser, nil)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestUpdateRole_RepoError(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	repo.On("UpdateRole", mock.Anything, id, entity.RoleUser, (*uuid.UUID)(nil)).
		Return(errors.New("db error"))

	err := uc.UpdateRole(context.Background(), id, entity.RoleUser, nil)

	assert.Error(t, err)
}

// ─── UploadAvatar — storage not configured ────────────────────────────────────

func TestUploadAvatar_NoStorage(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil) // storage = nil

	_, err := uc.UploadAvatar(context.Background(), uuid.New(), nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not configured")
}
