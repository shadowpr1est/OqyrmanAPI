package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

// Subscriber is satisfied by *hub.NotificationHub.
type Subscriber interface {
	Subscribe(userID uuid.UUID) (connID string, ch <-chan *entity.Notification, cancel func())
}

type Handler struct {
	uc  domainUseCase.NotificationUseCase
	hub Subscriber // nil → SSE endpoint returns 503
}

func NewHandler(uc domainUseCase.NotificationUseCase, hub Subscriber) *Handler {
	return &Handler{uc: uc, hub: hub}
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
	userID := middleware.GetUserID(c)

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
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
	userID := middleware.GetUserID(c)

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
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
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
	userID := middleware.GetUserID(c)

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
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		common.InternalError(c)
		return
	}

	c.Status(http.StatusNoContent)
}

func toNotificationResponse(n *entity.Notification) notificationResponse {
	resp := notificationResponse{
		ID:        n.ID.String(),
		Type:      string(n.Type),
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

// @Summary     SSE — поток уведомлений
// @Tags        notifications
// @Security    BearerAuth
// @Produce     text/event-stream
// @Success     200
// @Router      /notifications/stream [get]
func (h *Handler) Stream(c *gin.Context) {
	if h.hub == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "SSE not available"})
		return
	}

	userID := middleware.GetUserID(c)
	_, ch, cancel := h.hub.Subscribe(userID)
	defer cancel()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ctx := c.Request.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case n, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(toNotificationResponse(n))
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			c.Writer.Flush()
		}
	}
}
