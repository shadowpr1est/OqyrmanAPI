package ai

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/delivery/http/handler/common"
)

type recommendResponse struct {
	Recommendations string `json:"recommendations"`
}

type recommendBooksResponse struct {
	Items []bookRecommendItem `json:"items"`
}

type bookRecommendItem struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	ISBN        string              `json:"isbn"`
	CoverURL    string              `json:"cover_url,omitempty"`
	Description string              `json:"description"`
	Language    string              `json:"language"`
	Year        int                 `json:"year,omitempty"`
	AvgRating   float64             `json:"avg_rating"`
	TotalPages  *int                `json:"total_pages,omitempty"`
	File        *common.BookFileRef `json:"file,omitempty"`
	Author      common.AuthorRef    `json:"author"`
	Genre       common.GenreRef     `json:"genre"`
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
	Action    string `json:"action" binding:"required,oneof=ask translate"`
	Selection string `json:"selection" binding:"required"`
	Context   string `json:"context"`
	// TargetLang используется только для action=translate. Допустимо ru|en|kk.
	// Пусто = ru (по умолчанию).
	TargetLang string `json:"target_lang" binding:"omitempty,oneof=ru en kk"`
}

type seedConversationRequest struct {
	Action    string `json:"action" binding:"required,oneof=ask translate"`
	Selection string `json:"selection" binding:"required"`
	Answer    string `json:"answer" binding:"required"`
}

type seedConversationResponse struct {
	ID    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
