package ai

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type recommendResponse struct {
	Recommendations string `json:"recommendations"`
}

// ── Conversations ─────────────────────────────────────────────────────────────

type createConversationResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type conversationListItem struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type messageDTO struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

type conversationDetailResponse struct {
	ID        uuid.UUID    `json:"id"`
	Title     string       `json:"title"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	Messages  []messageDTO `json:"messages"`
}

type sendMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

type sendMessageResponse struct {
	UserMessage messageDTO `json:"user_message"`
	AIMessage   messageDTO `json:"ai_message"`
}

// ── Streaming SSE events ────���────────────────────────────────────────────────

type streamChunkEvent struct {
	Type    string `json:"type"`    // "chunk" | "error"
	Content string `json:"content"`
}

type streamDoneEvent struct {
	Type        string     `json:"type"` // "done"
	UserMessage messageDTO `json:"user_message"`
	AIMessage   messageDTO `json:"ai_message"`
}

type suggestedPromptsResponse struct {
	Prompts []string `json:"prompts"`
}

// ── Reader selection actions ────────────────────────────────────────────────

type explainSelectionRequest struct {
	Action    string `json:"action" binding:"required,oneof=explain translate identify"`
	Selection string `json:"selection" binding:"required"`
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
