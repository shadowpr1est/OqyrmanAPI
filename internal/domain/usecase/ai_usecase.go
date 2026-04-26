package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

// StreamCallback вызывается для каждого фрагмента текста при стриминге.
type StreamCallback func(chunk string) error

type AIUseCase interface {
	Recommend(ctx context.Context, userID string) (string, error)
	RecommendBooks(ctx context.Context, userID uuid.UUID) ([]*entity.BookView, error)
	SuggestedPrompts(ctx context.Context, userID uuid.UUID) []string

	// Conversation-based chat
	CreateConversation(ctx context.Context, userID uuid.UUID) (*entity.Conversation, error)
	SendMessage(ctx context.Context, convID, userID uuid.UUID, message string) (*entity.ChatMessage, *entity.ChatMessage, error)
	SendMessageStream(ctx context.Context, convID, userID uuid.UUID, message string, cb StreamCallback) (*entity.ChatMessage, *entity.ChatMessage, error)
	ListConversations(ctx context.Context, userID uuid.UUID) ([]*entity.Conversation, error)
	GetConversation(ctx context.Context, id, userID uuid.UUID) (*entity.Conversation, []*entity.ChatMessage, error)
	DeleteConversation(ctx context.Context, id, userID uuid.UUID) error

	// Reader selection actions — one-shot streaming, no conversation persisted.
	// action ∈ {"ask", "translate"}. `surrounding` — соседний текст для контекста (опционально).
	// targetLang — целевой язык для translate (ru|en|kk), пусто = ru.
	ExplainSelection(ctx context.Context, userID, bookID uuid.UUID, action, selection, surrounding, targetLang string, cb StreamCallback) error

	// SeedConversationFromSelection создаёт новую беседу с уже подгруженной парой
	// (запрос пользователя по фрагменту, ответ ИИ). LLM не вызывается — ответ
	// пришёл из ExplainSelection. Возвращает беседу с готовым контекстом, чтобы
	// пользователь мог сразу задать follow-up.
	SeedConversationFromSelection(ctx context.Context, userID, bookID uuid.UUID, action, selection, answer string) (*entity.Conversation, error)
}
