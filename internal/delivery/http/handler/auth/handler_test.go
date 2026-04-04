package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ─── Mock ─────────────────────────────────────────────────────────────────────

type mockAuthUseCase struct{ mock.Mock }

func (m *mockAuthUseCase) Register(ctx context.Context, u *entity.User) (*entity.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}
func (m *mockAuthUseCase) Login(ctx context.Context, email, password string) (*domainUseCase.TokenPair, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUseCase.TokenPair), args.Error(1)
}
func (m *mockAuthUseCase) Logout(ctx context.Context, refreshToken string) error {
	return m.Called(ctx, refreshToken).Error(0)
}
func (m *mockAuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (*domainUseCase.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUseCase.TokenPair), args.Error(1)
}
func (m *mockAuthUseCase) SendVerificationCode(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockAuthUseCase) VerifyEmail(ctx context.Context, email, code string) (*domainUseCase.TokenPair, error) {
	args := m.Called(ctx, email, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUseCase.TokenPair), args.Error(1)
}
func (m *mockAuthUseCase) ForgotPassword(ctx context.Context, email string) error {
	return m.Called(ctx, email).Error(0)
}
func (m *mockAuthUseCase) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	return m.Called(ctx, email, code, newPassword).Error(0)
}
func (m *mockAuthUseCase) LoginWithGoogle(ctx context.Context, idToken string) (*domainUseCase.TokenPair, error) {
	args := m.Called(ctx, idToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUseCase.TokenPair), args.Error(1)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func init() {
	gin.SetMode(gin.TestMode)
}

func serve(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func newTestRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	r.POST("/auth/logout", h.Logout)
	r.POST("/auth/refresh", h.RefreshToken)
	return r
}

var tokenPair = &domainUseCase.TokenPair{
	AccessToken:  "access-token",
	RefreshToken: "refresh-token",
}

// ─── Register ─────────────────────────────────────────────────────────────────

func TestRegister_Success(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	user := &entity.User{ID: uuid.New(), Email: "a@b.com"}
	uc.On("Register", mock.Anything, mock.AnythingOfType("*entity.User")).Return(user, nil)

	w := serve(r, http.MethodPost, "/auth/register",
		`{"email":"a@b.com","phone":"+77001234567","password":"Password1","name":"Test","surname":"User"}`)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "message")
	assert.Contains(t, w.Body.String(), "user_id")
	uc.AssertExpectations(t)
}

func TestRegister_InvalidBody(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	// missing required email
	w := serve(r, http.MethodPost, "/auth/register", `{"phone":"+77001234567","password":"Password1"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Register")
}

func TestRegister_EmailExists(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("Register", mock.Anything, mock.AnythingOfType("*entity.User")).
		Return(nil, entity.ErrEmailTaken)

	w := serve(r, http.MethodPost, "/auth/register",
		`{"email":"a@b.com","phone":"+77001234567","password":"Password1","name":"Test","surname":"User"}`)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "email already taken")
}

func TestRegister_InternalError(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("Register", mock.Anything, mock.AnythingOfType("*entity.User")).
		Return(nil, errors.New("db error"))

	w := serve(r, http.MethodPost, "/auth/register",
		`{"email":"a@b.com","phone":"+77001234567","password":"Password1","name":"Test","surname":"User"}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("Login", mock.Anything, "a@b.com", "Password1").Return(tokenPair, nil)

	w := serve(r, http.MethodPost, "/auth/login", `{"email":"a@b.com","password":"Password1"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
	uc.AssertExpectations(t)
}

func TestLogin_InvalidBody(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	// missing password
	w := serve(r, http.MethodPost, "/auth/login", `{"email":"a@b.com"}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Login")
}

func TestLogin_WrongCredentials(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("Login", mock.Anything, "a@b.com", "WrongPass1").
		Return(nil, errors.New("invalid credentials"))

	w := serve(r, http.MethodPost, "/auth/login", `{"email":"a@b.com","password":"WrongPass1"}`)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid credentials")
}

// ─── Logout ───────────────────────────────────────────────────────────────────

func TestLogout_Success(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("Logout", mock.Anything, "some-token").Return(nil)

	w := serve(r, http.MethodPost, "/auth/logout", `{"refresh_token":"some-token"}`)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertExpectations(t)
}

func TestLogout_InvalidBody(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	// missing refresh_token
	w := serve(r, http.MethodPost, "/auth/logout", `{}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "Logout")
}

func TestLogout_Error(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("Logout", mock.Anything, "some-token").Return(errors.New("token store unavailable"))

	w := serve(r, http.MethodPost, "/auth/logout", `{"refresh_token":"some-token"}`)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestRefreshToken_Success(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("RefreshToken", mock.Anything, "old-token").Return(tokenPair, nil)

	w := serve(r, http.MethodPost, "/auth/refresh", `{"refresh_token":"old-token"}`)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "access_token")
	uc.AssertExpectations(t)
}

func TestRefreshToken_InvalidBody(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	w := serve(r, http.MethodPost, "/auth/refresh", `{}`)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "RefreshToken")
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	uc := new(mockAuthUseCase)
	h := NewHandler(uc)
	r := newTestRouter(h)

	uc.On("RefreshToken", mock.Anything, "bad-token").
		Return(nil, errors.New("invalid refresh token"))

	w := serve(r, http.MethodPost, "/auth/refresh", `{"refresh_token":"bad-token"}`)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid refresh token")
}
