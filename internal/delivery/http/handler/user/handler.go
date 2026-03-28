package user

import (
	"log/slog"
	"errors"
	"net/http"
	"strconv"

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
	userID := middleware.GetUserID(c)

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
	userID := middleware.GetUserID(c)

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.uc.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if req.Email != nil {
		existing.Email = *req.Email
	}
	if req.Phone != nil {
		existing.Phone = *req.Phone
	}
	if req.FullName != nil {
		existing.FullName = *req.FullName
	}
	// FIX: avatar_url убран из updateUserRequest и из этого хендлера.
	// Раньше пользователь мог передать avatar_url в теле запроса,
	// получить 200 OK, но аватар не менялся — user_repo.Update не включает
	// avatar_url в SQL. Теперь поле отсутствует в DTO — нет ложных ожиданий.
	// Для смены аватара: POST /users/me/avatar (multipart/form-data).

	result, err := h.uc.Update(c.Request.Context(), existing)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(result))
}

// @Summary     Список пользователей
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Param       limit  query int false "Лимит"    default(20)
// @Param       offset query int false "Смещение" default(0)
// @Success     200 {object} map[string]interface{}
// @Router      /admin/users [get]
func (h *Handler) ListAll(c *gin.Context) {
	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(c.Query("offset")); err == nil && o >= 0 {
		offset = o
	}
	users, total, err := h.uc.ListAll(c.Request.Context(), limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
// @Router      /admin/users/{id}/role [patch]
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

	var libraryID *uuid.UUID
	if req.LibraryID != nil {
		parsed, err := uuid.Parse(*req.LibraryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
			return
		}
		libraryID = &parsed
	}

	if err := h.uc.UpdateRole(c.Request.Context(), id, entity.Role(req.Role), libraryID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Удалить пользователя (admin)
// @Tags        users
// @Security    BearerAuth
// @Param       id path string true "ID пользователя"
// @Success     204
// @Router      /admin/users/{id} [delete]
func (h *Handler) AdminDelete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.uc.AdminDelete(c.Request.Context(), id); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	userID := middleware.GetUserID(c)
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
	userID := middleware.GetUserID(c)

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
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpeg, png, webp images are allowed"})
		return
	}

	user, err := h.uc.UploadAvatar(c.Request.Context(), userID, &fileupload.File{
		Filename:    fh.Filename,
		Reader:      f,
		Size:        fh.Size,
		ContentType: contentType,
	})
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, toUserResponse(user))
}

// @Summary     Удалить аккаунт
// @Tags        users
// @Security    BearerAuth
// @Success     204
// @Failure     401 {object} map[string]string
// @Router      /users/me [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)

	if err := h.uc.Delete(c.Request.Context(), userID); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
