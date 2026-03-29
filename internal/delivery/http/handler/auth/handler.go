package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
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
// @Success     201 {object} tokenResponse
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
	if _, err := h.uc.Register(c.Request.Context(), user); err != nil {
		if errors.Is(err, entity.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		if errors.Is(err, entity.ErrValidation) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		slog.ErrorContext(c.Request.Context(), "register error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	// Сразу логиним после регистрации
	pair, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusCreated, tokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	})
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
// @Accept      json
// @Param       input body logoutRequest true "Refresh токен"
// @Success     204
// @Failure     400 {object} map[string]string
// @Router      /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.uc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
