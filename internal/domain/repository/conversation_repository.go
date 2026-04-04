package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

type ConversationRepository interface {
	Create(ctx context.Context, conv *entity.Conversation) (*entity.Conversation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Conversation, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*entity.Conversation, error)
	UpdateTitle(ctx context.Context, id uuid.UUID, title string) error
	Touch(ctx context.Context, id uuid.UUID) error // обновляет updated_at
	Delete(ctx context.Context, id uuid.UUID) error

	SaveMessage(ctx context.Context, msg *entity.ChatMessage) (*entity.ChatMessage, error)
	// ListMessages возвращает последние limit сообщений в хронологическом порядке.
	ListMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]*entity.ChatMessage, error)
}
