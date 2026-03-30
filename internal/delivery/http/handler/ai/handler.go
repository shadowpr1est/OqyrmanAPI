package ai

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

type Handler struct {
	uc domainUseCase.AIUseCase
}

func NewHandler(uc domainUseCase.AIUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Персональные рекомендации книг
// @Description Генерирует рекомендации на основе истории чтения и вишлиста пользователя
// @Tags        ai
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} recommendResponse
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/recommend [post]
func (h *Handler) Recommend(c *gin.Context) {
	userID := middleware.GetUserID(c)

	result, err := h.uc.Recommend(c.Request.Context(), userID.String())
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, recommendResponse{Recommendations: result})
}

// @Summary     Чат с книжным ассистентом
// @Description Отвечает на вопросы пользователя по книгам и чтению
// @Tags        ai
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       input body chatRequest  true "Сообщение пользователя"
// @Success     200 {object} chatResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/chat [post]
const maxChatMessageLen = 2000

func (h *Handler) Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len([]rune(req.Message)) > maxChatMessageLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message must not exceed 2000 characters"})
		return
	}

	result, err := h.uc.Chat(c.Request.Context(), req.Message)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "internal error", "err", err, "path", c.FullPath())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, chatResponse{Reply: result})
}
