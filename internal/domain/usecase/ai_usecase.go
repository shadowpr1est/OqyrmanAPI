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
	SuggestedPrompts(ctx context.Context, userID uuid.UUID) []string

	// Conversation-based chat
	CreateConversation(ctx context.Context, userID uuid.UUID) (*entity.Conversation, error)
	SendMessage(ctx context.Context, convID, userID uuid.UUID, message string) (*entity.ChatMessage, *entity.ChatMessage, error)
	SendMessageStream(ctx context.Context, convID, userID uuid.UUID, message string, cb StreamCallback) (*entity.ChatMessage, *entity.ChatMessage, error)
	ListConversations(ctx context.Context, userID uuid.UUID) ([]*entity.Conversation, error)
	GetConversation(ctx context.Context, id, userID uuid.UUID) (*entity.Conversation, []*entity.ChatMessage, error)
	DeleteConversation(ctx context.Context, id, userID uuid.UUID) error
}
