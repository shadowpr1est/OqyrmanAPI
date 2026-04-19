package ai

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/middleware"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
)

const maxMessageLen = 2000

type Handler struct {
	uc domainUseCase.AIUseCase
}

func NewHandler(uc domainUseCase.AIUseCase) *Handler {
	return &Handler{uc: uc}
}

// @Summary     Подсказки для чата
// @Description Возвращает персонализированные подсказки-промпты для чат-виджета
// @Tags        ai
// @Security    BearerAuth
// @Produce     json
// @Success     200 {object} suggestedPromptsResponse
// @Failure     401 {object} map[string]string
// @Router      /ai/prompts [get]
func (h *Handler) SuggestedPrompts(c *gin.Context) {
	userID := middleware.GetUserID(c)
	prompts := h.uc.SuggestedPrompts(c.Request.Context(), userID)
	c.JSON(http.StatusOK, suggestedPromptsResponse{Prompts: prompts})
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
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, recommendResponse{Recommendations: result})
}

// @Summary     Создать беседу
// @Description Создаёт новую беседу с AI ассистентом
// @Tags        ai
// @Security    BearerAuth
// @Produce     json
// @Success     201 {object} createConversationResponse
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/conversations [post]
func (h *Handler) CreateConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	conv, err := h.uc.CreateConversation(c.Request.Context(), userID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "create conversation error", "err", err)
		common.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, createConversationResponse{
		ID:        conv.ID,
		Title:     conv.Title,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	})
}

// @Summary     Список бесед
// @Description Возвращает все беседы текущего пользователя
// @Tags        ai
// @Security    BearerAuth
// @Produce     json
// @Success     200 {array} conversationListItem
// @Failure     401 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/conversations [get]
func (h *Handler) ListConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	convs, err := h.uc.ListConversations(c.Request.Context(), userID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "list conversations error", "err", err)
		common.InternalError(c)
		return
	}

	items := make([]conversationListItem, len(convs))
	for i, cv := range convs {
		items[i] = conversationListItem{
			ID:        cv.ID,
			Title:     cv.Title,
			CreatedAt: cv.CreatedAt,
			UpdatedAt: cv.UpdatedAt,
		}
	}
	c.JSON(http.StatusOK, items)
}

// @Summary     Получить беседу
// @Description Возвращает беседу вместе со всеми сообщениями
// @Tags        ai
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID беседы"
// @Success     200 {object} conversationDetailResponse
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/conversations/{id} [get]
func (h *Handler) GetConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	conv, msgs, err := h.uc.GetConversation(c.Request.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrConversationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		case errors.Is(err, entity.ErrForbidden):
			common.Forbidden(c)
		default:
			slog.ErrorContext(c.Request.Context(), "get conversation error", "err", err)
			common.InternalError(c)
		}
		return
	}

	msgDTOs := make([]messageDTO, len(msgs))
	for i, m := range msgs {
		msgDTOs[i] = messageDTO{
			ID:             m.ID,
			ConversationID: m.ConversationID,
			Role:           m.Role,
			Content:        m.Content,
			CreatedAt:      m.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, conversationDetailResponse{
		ID:        conv.ID,
		Title:     conv.Title,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
		Messages:  msgDTOs,
	})
}

// @Summary     Отправить сообщение
// @Description Отправляет сообщение в беседу и получает ответ AI
// @Tags        ai
// @Security    BearerAuth
// @Accept      json
// @Produce     json
// @Param       id    path string          true "ID беседы"
// @Param       input body sendMessageRequest true "Сообщение"
// @Success     200 {object} sendMessageResponse
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/conversations/{id}/messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	if len([]rune(req.Message)) > maxMessageLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message must not exceed 2000 characters"})
		return
	}

	userMsg, aiMsg, err := h.uc.SendMessage(c.Request.Context(), id, userID, req.Message)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrValidation):
			common.BadRequest(c, common.CodeValidationError, "invalid input")
			return
		case errors.Is(err, entity.ErrConversationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		case errors.Is(err, entity.ErrForbidden):
			common.Forbidden(c)
		default:
			slog.ErrorContext(c.Request.Context(), "send message error", "err", err)
			common.InternalError(c)
		}
		return
	}

	c.JSON(http.StatusOK, sendMessageResponse{
		UserMessage: messageDTO{
			ID:             userMsg.ID,
			ConversationID: userMsg.ConversationID,
			Role:           userMsg.Role,
			Content:        userMsg.Content,
			CreatedAt:      userMsg.CreatedAt,
		},
		AIMessage: messageDTO{
			ID:             aiMsg.ID,
			ConversationID: aiMsg.ConversationID,
			Role:           aiMsg.Role,
			Content:        aiMsg.Content,
			CreatedAt:      aiMsg.CreatedAt,
		},
	})
}

// @Summary     Отправить сообщение (streaming)
// @Description Отправляет сообщение и стримит ответ AI через Server-Sent Events
// @Tags        ai
// @Security    BearerAuth
// @Accept      json
// @Produce     text/event-stream
// @Param       id    path string            true "ID беседы"
// @Param       input body sendMessageRequest true "Сообщение"
// @Success     200 {string} string "SSE stream"
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/conversations/{id}/messages/stream [post]
func (h *Handler) SendMessageStream(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	if len([]rune(req.Message)) > maxMessageLen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message must not exceed 2000 characters"})
		return
	}

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	userMsg, aiMsg, err := h.uc.SendMessageStream(c.Request.Context(), id, userID, req.Message, func(chunk string) error {
		// SSE формат: data: {"chunk":"..."}\n\n
		c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "chunk", Content: chunk}) + "\n\n")
		flusher.Flush()
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrValidation):
			c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "error", Content: "invalid input"}) + "\n\n")
		case errors.Is(err, entity.ErrConversationNotFound):
			c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "error", Content: "conversation not found"}) + "\n\n")
		case errors.Is(err, entity.ErrForbidden):
			c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "error", Content: "forbidden"}) + "\n\n")
		default:
			slog.ErrorContext(c.Request.Context(), "send message stream error", "err", err)
			c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "error", Content: "internal error"}) + "\n\n")
		}
		flusher.Flush()
		return
	}

	// Финальное событие с полными данными сообщений
	c.Writer.WriteString("data: " + toJSON(streamDoneEvent{
		Type: "done",
		UserMessage: messageDTO{
			ID:             userMsg.ID,
			ConversationID: userMsg.ConversationID,
			Role:           userMsg.Role,
			Content:        userMsg.Content,
			CreatedAt:      userMsg.CreatedAt,
		},
		AIMessage: messageDTO{
			ID:             aiMsg.ID,
			ConversationID: aiMsg.ConversationID,
			Role:           aiMsg.Role,
			Content:        aiMsg.Content,
			CreatedAt:      aiMsg.CreatedAt,
		},
	}) + "\n\n")
	flusher.Flush()
}

// @Summary     AI по выделенному фрагменту (streaming)
// @Description Объясняет / переводит / идентифицирует выделенный фрагмент книги. Стрим SSE, беседа не создаётся.
// @Tags        ai
// @Security    BearerAuth
// @Accept      json
// @Produce     text/event-stream
// @Param       bookId path string                  true "ID книги"
// @Param       input  body explainSelectionRequest true "Действие и выделенный текст"
// @Success     200 {string} string "SSE stream"
// @Failure     400 {object} map[string]string
// @Failure     401 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/books/{bookId}/explain [post]
func (h *Handler) ExplainSelection(c *gin.Context) {
	userID := middleware.GetUserID(c)

	bookID, err := uuid.Parse(c.Param("bookId"))
	if err != nil {
		common.BadRequest(c, common.CodeValidationError, "invalid book id")
		return
	}

	var req explainSelectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationErr(c, err)
		return
	}

	if len([]rune(req.Selection)) > maxMessageLen {
		common.BadRequest(c, common.CodeValidationError, "selection must not exceed 2000 characters")
		return
	}

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	err = h.uc.ExplainSelection(c.Request.Context(), userID, bookID, req.Action, req.Selection, func(chunk string) error {
		c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "chunk", Content: chunk}) + "\n\n")
		flusher.Flush()
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrValidation):
			c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "error", Content: "invalid input"}) + "\n\n")
		case errors.Is(err, entity.ErrNotFound):
			c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "error", Content: "book not found"}) + "\n\n")
		default:
			slog.ErrorContext(c.Request.Context(), "explain selection error", "err", err)
			c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "error", Content: "internal error"}) + "\n\n")
		}
		flusher.Flush()
		return
	}

	// Done event — фронт слушает тот же тип, что и у обычного чата,
	// но без user/ai message payload'ов (ничего не сохраняем).
	c.Writer.WriteString("data: " + toJSON(streamChunkEvent{Type: "done"}) + "\n\n")
	flusher.Flush()
}

// @Summary     Удалить беседу
// @Description Удаляет беседу и все её сообщения
// @Tags        ai
// @Security    BearerAuth
// @Produce     json
// @Param       id path string true "ID беседы"
// @Success     204
// @Failure     401 {object} map[string]string
// @Failure     403 {object} map[string]string
// @Failure     404 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /ai/conversations/{id} [delete]
func (h *Handler) DeleteConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid conversation id"})
		return
	}

	err = h.uc.DeleteConversation(c.Request.Context(), id, userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrConversationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		case errors.Is(err, entity.ErrForbidden):
			common.Forbidden(c)
		default:
			slog.ErrorContext(c.Request.Context(), "delete conversation error", "err", err)
			common.InternalError(c)
		}
		return
	}

	c.Status(http.StatusNoContent)
}
