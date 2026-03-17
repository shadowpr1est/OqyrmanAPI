package user

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/fileupload"
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

// @Summary     Список пользователей
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Param       limit  query int false "Лимит"  default(20)
// @Param       offset query int false "Смещение" default(0)
// @Success     200 {object} map[string]interface{}
// @Router      /admin/users [get]
func (h *Handler) ListAll(c *gin.Context) {
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		fmt.Sscan(l, &limit)
	}
	if o := c.Query("offset"); o != "" {
		fmt.Sscan(o, &offset)
	}
	users, total, err := h.uc.ListAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	items := make([]userResponse, len(users))
	for i, u := range users {
		items[i] = toUserResponse(u)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

// @Summary     Изменить роль пользователя
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Param       id    path string          true "ID пользователя"
// @Param       input body updateRoleRequest true "Новая роль"
// @Success     204
// @Router      /admin/users/:id/role [patch]
func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.uc.UpdateRole(c.Request.Context(), id, entity.Role(req.Role)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Удалить пользователя (admin)
// @Tags        users
// @Security    BearerAuth
// @Param       id path string true "ID пользователя"
// @Success     204
// @Router      /admin/users/:id [delete]
func (h *Handler) AdminDelete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.AdminDelete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     QR-код читательского билета
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} map[string]string
// @Router      /users/me/qr [get]
func (h *Handler) GetQR(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)
	user, err := h.uc.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"qr_code": user.QRCode})
}

// @Summary     Загрузить аватар пользователя
// @Tags        users
// @Security    BearerAuth
// @Accept      multipart/form-data
// @Produce     json
// @Param       file formData file true "Изображение аватара (jpg, png)"
// @Success     200 {object} userResponse
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /users/me/avatar [post]
func (h *Handler) UploadAvatar(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	f, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	user, err := h.uc.UploadAvatar(c.Request.Context(), userID, &fileupload.File{
		Filename:    fh.Filename,
		Reader:      f,
		Size:        fh.Size,
		ContentType: contentType,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		QRCode:    u.QRCode,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
