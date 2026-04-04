package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type AIUseCase interface {
	Recommend(ctx context.Context, userID string) (string, error)

	// Conversation-based chat
	CreateConversation(ctx context.Context, userID uuid.UUID) (*entity.Conversation, error)
	SendMessage(ctx context.Context, convID, userID uuid.UUID, message string) (*entity.ChatMessage, *entity.ChatMessage, error)
	ListConversations(ctx context.Context, userID uuid.UUID) ([]*entity.Conversation, error)
	GetConversation(ctx context.Context, id, userID uuid.UUID) (*entity.Conversation, []*entity.ChatMessage, error)
	DeleteConversation(ctx context.Context, id, userID uuid.UUID) error
}
