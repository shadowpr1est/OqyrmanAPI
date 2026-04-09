package user

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
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
		common.NotFound(c, "user not found")
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
		common.ValidationErr(c, err)
		return
	}

	// Build a partial entity — only set fields present in the request.
	// The repo SQL uses CASE WHEN '' to preserve existing DB values for omitted fields,
	// so no preliminary GetByID round-trip is needed.
	patch := &entity.User{ID: userID}
	if req.Name != nil {
		patch.Name = *req.Name
	}
	if req.Surname != nil {
		patch.Surname = *req.Surname
	}
	if req.Email != nil {
		patch.Email = *req.Email
	}
	if req.Phone != nil {
		patch.Phone = *req.Phone
	}

	result, err := h.uc.Update(c.Request.Context(), patch)
	if err != nil {
		if errors.Is(err, entity.ErrValidation) {
			common.BadRequest(c, common.CodeValidationError, "invalid input")
			return
		}
		if errors.Is(err, entity.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
// @Success     200 {object} map[string]userViewResponse
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
	if offset > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "offset must not exceed 10000"})
		return
	}
	users, total, err := h.uc.ListAllView(c.Request.Context(), limit, offset)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	items := make([]userViewResponse, len(users))
	for i, u := range users {
		items[i] = toUserViewResponse(u)
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "limit": limit, "offset": offset})
}

// @Summary     Обновить пользователя (admin)
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string               true "ID пользователя"
// @Param       input body adminUpdateUserRequest true "Данные для обновления"
// @Success     200 {object} userViewResponse
// @Failure     400,404,409 {object} map[string]string
// @Router      /admin/users/{id} [patch]
func (h *Handler) AdminUpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req adminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	// Парсим library_id если передан
	var libraryID *uuid.UUID
	if req.LibraryID != nil {
		parsed, err := uuid.Parse(*req.LibraryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
			return
		}
		libraryID = &parsed
	}

	// Парсим role если передан
	var role *entity.Role
	if req.Role != nil {
		r := entity.Role(*req.Role)
		role = &r
	}

	result, err := h.uc.AdminUpdateUser(c.Request.Context(), id, role, libraryID, req.Name, req.Surname, req.Email, req.Phone)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if errors.Is(err, entity.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "admin update user error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, toUserViewResponse(result))
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
		common.InternalError(c)
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
		common.NotFound(c, "user not found")
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

	// Проверяем расширение файла до чтения байт — отсекаем явно неподходящие имена.
	switch strings.ToLower(filepath.Ext(fh.Filename)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		// ok
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "only jpeg, png, webp images are allowed"})
		return
	}

	// Detect content type from actual file bytes, not the client-supplied header.
	// Защита от полиглот-файлов: оба условия должны совпадать.
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot process file"})
		return
	}
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
		common.InternalError(c)
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
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Создать сотрудника (admin)
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body createStaffRequest true "Данные сотрудника"
// @Success     201 {object} userViewResponse
// @Failure     400 {object} map[string]string
// @Failure     409 {object} map[string]string
// @Router      /admin/users/staff [post]
func (h *Handler) CreateStaff(c *gin.Context) {
	var req createStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	libraryID, err := uuid.Parse(req.LibraryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid library_id"})
		return
	}

	user, err := h.uc.CreateStaff(c.Request.Context(), req.Email, req.Password, req.Name, req.Surname, req.Phone, libraryID)
	if err != nil {
		if errors.Is(err, entity.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email already taken"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, toUserViewResponse(user))
}
func toUserResponse(u *entity.User) userResponse {
	return userResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		Phone:     u.Phone,
		Name:      u.Name,
		Surname:   u.Surname,
		AvatarURL: u.AvatarURL,
		Role:      string(u.Role),
		QRCode:    u.QRCode,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toUserViewResponse(v *entity.UserView) userViewResponse {
	resp := userViewResponse{
		ID:          v.ID.String(),
		Email:       v.Email,
		Phone:       v.Phone,
		Name:        v.Name,
		Surname:     v.Surname,
		AvatarURL:   v.AvatarURL,
		Role:        string(v.Role),
		LibraryName: v.LibraryName,
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if v.LibraryID != nil {
		s := v.LibraryID.String()
		resp.LibraryID = &s
	}
	return resp
}

// @Summary     Список активных сессий
// @Tags        users
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array}  sessionResponse
// @Router      /users/me/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessions, err := h.uc.ListSessions(c.Request.Context(), userID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	resp := make([]sessionResponse, len(sessions))
	for i, s := range sessions {
		resp[i] = toSessionResponse(s)
	}
	c.JSON(http.StatusOK, resp)
}

// @Summary     Отозвать сессию
// @Tags        users
// @Security    BearerAuth
// @Param       id path string true "ID сессии"
// @Success     204
// @Failure     404 {object} map[string]string
// @Router      /users/me/sessions/{id} [delete]
func (h *Handler) RevokeSession(c *gin.Context) {
	userID := middleware.GetUserID(c)
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}
	if err := h.uc.RevokeSession(c.Request.Context(), sessionID, userID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Отозвать все сессии
// @Tags        users
// @Security    BearerAuth
// @Success     204
// @Router      /users/me/sessions [delete]
func (h *Handler) RevokeAllSessions(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.uc.RevokeAllSessions(c.Request.Context(), userID); err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary     Сменить пароль
// @Tags        users
// @Security    BearerAuth
// @Accept      json
// @Param       input body changePasswordRequest true "Текущий и новый пароль"
// @Success     204
// @Failure     400 {object} map[string]string
// @Router      /users/me/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}
	if err := h.uc.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, entity.ErrValidation) {
			common.BadRequest(c, common.CodeValidationError, "invalid input")
			return
		}
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}
	c.Status(http.StatusNoContent)
}

func toSessionResponse(t *entity.Token) sessionResponse {
	return sessionResponse{
		ID:        t.ID.String(),
		UserAgent: t.UserAgent,
		IP:        t.IP,
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		ExpiresAt: t.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	}
}
