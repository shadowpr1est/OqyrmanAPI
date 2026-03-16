package user

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.UserUseCase
}

func NewHandler(uc domainUseCase.UserUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Профиль пользователя
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} userResponse
// @Failure     401 {object} map[string]string
// @Router      /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	user, err := h.uc.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// @Summary     Обновить профиль
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body updateUserRequest true "Данные для обновления"
// @Success     200 {object} userResponse
// @Failure     400 {object} map[string]string
// @Router      /users/me [put]
func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &entity.User{
		ID:        userID,
		Email:     req.Email,
		Phone:     req.Phone,
		FullName:  req.FullName,
		AvatarURL: req.AvatarURL,
	}

	result, err := h.uc.Update(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(result))
}

// @Summary     Удалить аккаунт
// @Tags        users
// @Security    BearerAuth
// @Success     204
// @Failure     401 {object} map[string]string
// @Router      /users/me [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	if err := h.uc.Delete(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func toUserResponse(u *entity.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		Phone:     u.Phone,
		FullName:  u.FullName,
		AvatarURL: u.AvatarURL,
		Role:      string(u.Role),
		QRCode:    u.QRCode,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
