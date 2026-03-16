package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.AuthUseCase
}

func NewHandler(uc domainUseCase.AuthUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Регистрация
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body registerRequest true "Данные для регистрации"
// @Success     201 {object} userResponse
// @Failure     400 {object} map[string]string
// @Router      /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &entity.User{
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: req.Password,
		FullName:     req.FullName,
	}

	result, err := h.uc.Register(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toUserResponse(result))
}

// @Summary     Вход
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body loginRequest true "Email и пароль"
// @Success     200 {object} tokenResponse
// @Failure     401 {object} map[string]string
// @Router      /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// @Summary     Выход
// @Tags        auth
// @Security    BearerAuth
// @Success     204
// @Failure     401 {object} map[string]string
// @Router      /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	if err := h.uc.Logout(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Обновить токен
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body refreshRequest true "Refresh токен"
// @Success     200 {object} tokenResponse
// @Failure     401 {object} map[string]string
// @Router      /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pair, err := h.uc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
}

// @Summary     Текущий пользователь
// @Tags        auth
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} userResponse
// @Failure     401 {object} map[string]string
// @Router      /auth/me [get]
func (h *Handler) Me(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	user, err := h.uc.Me(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

func toUserResponse(u *entity.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		Phone:     u.Phone,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		Role:      string(u.Role),
	}
}
