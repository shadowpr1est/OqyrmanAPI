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
func (m *mockUserRepo) AdminUpdate(ctx context.Context, id uuid.UUID, role *entity.Role, libraryID *uuid.UUID, name, surname, email, phone *string) (*entity.UserView, error) {
	args := m.Called(ctx, id, role, libraryID, name, surname, email, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.UserView), args.Error(1)
}
func (m *mockUserRepo) UpdateAvatarURL(ctx context.Context, id uuid.UUID, url string) error {
	return m.Called(ctx, id, url).Error(0)
}
func (m *mockUserRepo) ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error) {
	return nil, 0, nil
}

// ─── AdminUpdateUser ──────────────────────────────────────────────────────────

func TestAdminUpdateUser_StaffWithLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	libID := uuid.New()
	role := entity.RoleStaff
	repo.On("AdminUpdate", mock.Anything, id, &role, &libID,
		(*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil)).
		Return(&entity.UserView{}, nil)

	result, err := uc.AdminUpdateUser(context.Background(), id, &role, &libID, nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	repo.AssertExpectations(t)
}

func TestAdminUpdateUser_StaffWithoutLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	role := entity.RoleStaff
	_, err := uc.AdminUpdateUser(context.Background(), uuid.New(), &role, nil, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "library_id is required for Staff")
	repo.AssertNotCalled(t, "AdminUpdate")
}

func TestAdminUpdateUser_AdminWithoutLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	role := entity.RoleAdmin
	repo.On("AdminUpdate", mock.Anything, id, &role, (*uuid.UUID)(nil),
		(*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil)).
		Return(&entity.UserView{}, nil)

	result, err := uc.AdminUpdateUser(context.Background(), id, &role, nil, nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	repo.AssertExpectations(t)
}

func TestAdminUpdateUser_AdminWithLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	libID := uuid.New()
	role := entity.RoleAdmin
	_, err := uc.AdminUpdateUser(context.Background(), uuid.New(), &role, &libID, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "library_id must be null for non-Staff")
	repo.AssertNotCalled(t, "AdminUpdate")
}

func TestAdminUpdateUser_UserWithLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	libID := uuid.New()
	role := entity.RoleUser
	_, err := uc.AdminUpdateUser(context.Background(), uuid.New(), &role, &libID, nil, nil, nil, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "library_id must be null for non-Staff")
	repo.AssertNotCalled(t, "AdminUpdate")
}

func TestAdminUpdateUser_UserWithoutLibraryID(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	role := entity.RoleUser
	repo.On("AdminUpdate", mock.Anything, id, &role, (*uuid.UUID)(nil),
		(*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil)).
		Return(&entity.UserView{}, nil)

	result, err := uc.AdminUpdateUser(context.Background(), id, &role, nil, nil, nil, nil, nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	repo.AssertExpectations(t)
}

func TestAdminUpdateUser_RepoError(t *testing.T) {
	repo := new(mockUserRepo)
	uc := user.NewUserUseCase(repo, nil)

	id := uuid.New()
	role := entity.RoleUser
	repo.On("AdminUpdate", mock.Anything, id, &role, (*uuid.UUID)(nil),
		(*string)(nil), (*string)(nil), (*string)(nil), (*string)(nil)).
		Return(nil, errors.New("db error"))

	_, err := uc.AdminUpdateUser(context.Background(), id, &role, nil, nil, nil, nil, nil)

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
