package notification

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.NotificationUseCase
}

func NewHandler(uc domainUseCase.NotificationUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Мои уведомления
// @Tags        notifications
// @Security    BearerAuth
// @Produce     json
// @Param       limit  query int false "Лимит"    default(20)
// @Param       offset query int false "Смещение" default(0)
// @Success     200 {object} listNotificationsResponse
// @Router      /notifications [get]
func (h *Handler) ListMy(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	items, total, err := h.uc.ListMy(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]*notificationResponse, len(items))
	for i, n := range items {
		r := toNotificationResponse(n)
		resp[i] = &r
	}

	c.JSON(http.StatusOK, listNotificationsResponse{
		Items:  resp,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	})
}

// @Summary     Отметить уведомление как прочитанное
// @Tags        notifications
// @Security    BearerAuth
// @Param       id path string true "ID уведомления"
// @Success     204
// @Failure     404 {object} map[string]string
// @Router      /notifications/{id}/read [patch]
func (h *Handler) MarkRead(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.MarkRead(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary     Удалить уведомление
// @Tags        notifications
// @Security    BearerAuth
// @Param       id path string true "ID уведомления"
// @Success     204
// @Failure     404 {object} map[string]string
// @Router      /notifications/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet(middleware.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.uc.Delete(c.Request.Context(), id, userID); err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func toNotificationResponse(n *entity.Notification) notificationResponse {
	resp := notificationResponse{
		ID:        n.ID.String(),
		Title:     n.Title,
		Body:      n.Body,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if n.ReadAt != nil {
		t := n.ReadAt.Format("2006-01-02T15:04:05Z")
		resp.ReadAt = &t
	}
	return resp
}
