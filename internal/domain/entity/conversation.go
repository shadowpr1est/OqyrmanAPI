package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrConversationNotFound = errors.New("conversation not found")

type Conversation struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ChatMessage struct {
	ID             uuid.UUID `db:"id"`
	ConversationID uuid.UUID `db:"conversation_id"`
	Role           string    `db:"role"` // "user" | "assistant"
	Content        string    `db:"content"`
	CreatedAt      time.Time `db:"created_at"`
}
