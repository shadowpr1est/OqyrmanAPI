package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/usecase/auth"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// ─── Mocks ────────────────────────────────────────────────────────────────────

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
func (m *mockUserRepo) SetEmailVerified(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockUserRepo) GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) SetGoogleID(ctx context.Context, id uuid.UUID, googleID string) error {
	return nil
}
func (m *mockUserRepo) ListAllView(ctx context.Context, limit, offset int) ([]*entity.UserView, int, error) {
	return nil, 0, nil
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return m.Called(ctx, id, passwordHash).Error(0)
}
func (m *mockUserRepo) GetByPhone(ctx context.Context, phone string) (*entity.User, error) {
	return nil, entity.ErrNotFound
}
func (m *mockUserRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockVerifRepo struct{ mock.Mock }

func (m *mockVerifRepo) Save(ctx context.Context, code *entity.EmailVerificationCode) error {
	return m.Called(ctx, code).Error(0)
}
func (m *mockVerifRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.EmailVerificationCode, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.EmailVerificationCode), args.Error(1)
}
func (m *mockVerifRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

type mockResetRepo struct{ mock.Mock }

func (m *mockResetRepo) Save(ctx context.Context, code *entity.PasswordResetCode) error {
	return m.Called(ctx, code).Error(0)
}
func (m *mockResetRepo) GetByUserAndCode(ctx context.Context, userID uuid.UUID, code string) (*entity.PasswordResetCode, error) {
	args := m.Called(ctx, userID, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PasswordResetCode), args.Error(1)
}
func (m *mockResetRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

type mockTokenRepo struct{ mock.Mock }

func (m *mockTokenRepo) Save(ctx context.Context, t *entity.Token) error {
	return m.Called(ctx, t).Error(0)
}
func (m *mockTokenRepo) GetByRefreshToken(ctx context.Context, token string) (*entity.Token, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Token), args.Error(1)
}
func (m *mockTokenRepo) DeleteByRefreshToken(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *mockTokenRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Token, error) {
	return nil, nil
}
func (m *mockTokenRepo) DeleteByID(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}
func (m *mockTokenRepo) DeleteAllByUserID(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	return 0, nil
}

// noopLoginAttemptRepo — never locks, records nothing. Sufficient for tests
// that don't exercise lockout behavior directly.
type noopLoginAttemptRepo struct{}

func (noopLoginAttemptRepo) RecordFailedAttempt(ctx context.Context, email string, maxAttempts int, lockoutDuration time.Duration) (bool, error) {
	return false, nil
}
func (noopLoginAttemptRepo) IsLocked(ctx context.Context, email string, lockoutDuration time.Duration) (bool, error) {
	return false, nil
}
func (noopLoginAttemptRepo) Reset(ctx context.Context, email string) error { return nil }

// ─── Helpers ──────────────────────────────────────────────────────────────────

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	assert.NoError(t, err)
	return string(hash)
}

func newJWT(t *testing.T) *jwt.Manager {
	t.Helper()
	m, err := jwt.NewManager("test-secret-key-32-bytes-minimum!", 60)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// ─── Register ─────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	verifRepo := new(mockVerifRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, verifRepo, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	createdUser := &entity.User{ID: uuid.New(), Email: "test@example.com"}
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, entity.ErrNotFound)
	userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(createdUser, nil)
	verifRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationCode")).Return(nil)

	user := &entity.User{
		Email:        "test@example.com",
		Phone:        "+77001234567",
		PasswordHash: "Password1",
	}
	result, err := uc.Register(context.Background(), user)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	userRepo.AssertExpectations(t)
	verifRepo.AssertExpectations(t)
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	existing := &entity.User{ID: uuid.New(), Email: "test@example.com", EmailVerified: true}
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(existing, nil)

	user := &entity.User{
		Email:        "test@example.com",
		Phone:        "+77001234567",
		PasswordHash: "Password1",
	}
	_, err := uc.Register(context.Background(), user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email already taken")
	userRepo.AssertExpectations(t)
}

func TestRegister_PasswordTooShort(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	user := &entity.User{
		Email:        "test@example.com",
		PasswordHash: "Pass1",
	}
	_, err := uc.Register(context.Background(), user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8 characters")
}

func TestRegister_PasswordNoUppercase(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	user := &entity.User{
		Email:        "test@example.com",
		PasswordHash: "password1",
	}
	_, err := uc.Register(context.Background(), user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uppercase")
}

func TestRegister_PasswordNoDigit(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	user := &entity.User{
		Email:        "test@example.com",
		PasswordHash: "PasswordNoDigit",
	}
	_, err := uc.Register(context.Background(), user)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "digit")
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	userID := uuid.New()
	hashedPw := hashPassword(t, "Password1")
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(
		&entity.User{ID: userID, Email: "test@example.com", PasswordHash: hashedPw, Role: entity.RoleUser, EmailVerified: true},
		nil,
	)
	tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.Token")).Return(nil)

	pair, err := uc.Login(context.Background(), "test@example.com", "Password1")

	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	userRepo.On("GetByEmail", mock.Anything, "noone@example.com").Return(nil, errors.New("not found"))

	_, err := uc.Login(context.Background(), "noone@example.com", "Password1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestLogin_WrongPassword(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	hashedPw := hashPassword(t, "CorrectPassword1")
	userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(
		&entity.User{ID: uuid.New(), Email: "test@example.com", PasswordHash: hashedPw},
		nil,
	)

	_, err := uc.Login(context.Background(), "test@example.com", "WrongPassword1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	tokenRepo.On("DeleteByRefreshToken", mock.Anything, "some-refresh-token").Return(nil)

	err := uc.Logout(context.Background(), "some-refresh-token")

	assert.NoError(t, err)
	tokenRepo.AssertExpectations(t)
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestRefreshToken_Success(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	userID := uuid.New()
	existingToken := &entity.Token{
		ID:           uuid.New(),
		UserID:       userID,
		RefreshToken: "old-refresh-token",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}
	user := &entity.User{ID: userID, Email: "test@example.com", Role: entity.RoleUser}

	tokenRepo.On("GetByRefreshToken", mock.Anything, "old-refresh-token").Return(existingToken, nil)
	userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	tokenRepo.On("DeleteByRefreshToken", mock.Anything, "old-refresh-token").Return(nil)
	tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.Token")).Return(nil)

	pair, err := uc.RefreshToken(context.Background(), "old-refresh-token")

	assert.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, "old-refresh-token", pair.RefreshToken)
	tokenRepo.AssertExpectations(t)
}

func TestRefreshToken_Expired(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	expiredToken := &entity.Token{
		ID:           uuid.New(),
		UserID:       uuid.New(),
		RefreshToken: "expired-token",
		ExpiresAt:    time.Now().Add(-1 * time.Hour), // уже истёк
	}
	tokenRepo.On("GetByRefreshToken", mock.Anything, "expired-token").Return(expiredToken, nil)
	tokenRepo.On("DeleteByRefreshToken", mock.Anything, "expired-token").Return(nil)

	_, err := uc.RefreshToken(context.Background(), "expired-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expired")
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	userRepo := new(mockUserRepo)
	tokenRepo := new(mockTokenRepo)
	uc := auth.NewAuthUseCase(userRepo, tokenRepo, nil, nil, noopLoginAttemptRepo{}, nil, newJWT(t), "", "test-otp-secret", 30)

	_ = userRepo // не используется в этом тесте

	tokenRepo.On("GetByRefreshToken", mock.Anything, "invalid-token").Return(nil, errors.New("not found"))

	_, err := uc.RefreshToken(context.Background(), "invalid-token")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid refresh token")
}
