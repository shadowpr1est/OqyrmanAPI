package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm"
)

const (
	maxBooksInPrompt  = 10
	maxHistoryMessages = 20   // последние 20 сообщений идут в контекст LLM
	maxTitleLen       = 60    // обрезаем заголовок беседы по первому сообщению
)

var baseSystemPrompt = `Ты — книжный ассистент платформы Oqyrman. Отвечай на русском языке.
Помогаешь пользователю с выбором книг, обсуждаешь литературу и отвечаешь на вопросы о чтении.
Будь дружелюбным, лаконичным и полезным.`

type aiUseCase struct {
	sessionRepo  repository.ReadingSessionRepository
	wishlistRepo repository.WishlistRepository
	bookRepo     repository.BookRepository
	convRepo     repository.ConversationRepository
	llm          llm.LLMClient
}

func NewAIUseCase(
	sessionRepo repository.ReadingSessionRepository,
	wishlistRepo repository.WishlistRepository,
	bookRepo repository.BookRepository,
	convRepo repository.ConversationRepository,
	llm llm.LLMClient,
) domainUseCase.AIUseCase {
	return &aiUseCase{
		sessionRepo:  sessionRepo,
		wishlistRepo: wishlistRepo,
		bookRepo:     bookRepo,
		convRepo:     convRepo,
		llm:          llm,
	}
}

// ── Recommend ─────────────────────────────────────────────────────────────────

func (u *aiUseCase) Recommend(ctx context.Context, userIDStr string) (string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend invalid userID: %w", err)
	}

	sessions, _ := u.sessionRepo.ListByUser(ctx, userID)
	wishlist, _ := u.wishlistRepo.ListByUser(ctx, userID)

	readTitles := make([]string, 0, min(len(sessions), maxBooksInPrompt))
	for i, s := range sessions {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, s.BookID); err == nil {
			readTitles = append(readTitles, fmt.Sprintf("«%s»", book.Title))
		}
	}

	wishTitles := make([]string, 0, min(len(wishlist), maxBooksInPrompt))
	for i, w := range wishlist {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, w.BookID); err == nil {
			wishTitles = append(wishTitles, fmt.Sprintf("«%s»", book.Title))
		}
	}

	var prompt strings.Builder
	prompt.WriteString("Дай 5 персональных рекомендаций книг для чтения.\n\n")
	if len(readTitles) > 0 {
		prompt.WriteString(fmt.Sprintf("История чтения: %s.\n", strings.Join(readTitles, ", ")))
	} else {
		prompt.WriteString("История чтения пуста.\n")
	}
	if len(wishTitles) > 0 {
		prompt.WriteString(fmt.Sprintf("Вишлист: %s.\n", strings.Join(wishTitles, ", ")))
	} else {
		prompt.WriteString("Вишлист пуст.\n")
	}
	prompt.WriteString("\nПорекомендуй 5 книг, объясни кратко почему каждая.")

	return u.llm.Complete(ctx, baseSystemPrompt, []llm.Message{
		{Role: "user", Content: prompt.String()},
	})
}

// ── Conversations ─────────────────────────────────────────────────────────────

func (u *aiUseCase) CreateConversation(ctx context.Context, userID uuid.UUID) (*entity.Conversation, error) {
	now := time.Now()
	conv := &entity.Conversation{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     "Новый чат",
		CreatedAt: now,
		UpdatedAt: now,
	}
	return u.convRepo.Create(ctx, conv)
}

func (u *aiUseCase) SendMessage(ctx context.Context, convID, userID uuid.UUID, message string) (*entity.ChatMessage, *entity.ChatMessage, error) {
	if len([]rune(strings.TrimSpace(message))) == 0 {
		return nil, nil, fmt.Errorf("%w: message must not be empty", entity.ErrValidation)
	}

	// Проверяем владение беседой
	conv, err := u.convRepo.GetByID(ctx, convID)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, entity.ErrConversationNotFound
		}
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage get conv: %w", err)
	}
	if conv.UserID != userID {
		return nil, nil, entity.ErrForbidden
	}

	// Сохраняем сообщение пользователя
	userMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "user",
		Content:        message,
		CreatedAt:      time.Now(),
	}
	if _, err := u.convRepo.SaveMessage(ctx, userMsg); err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage save user msg: %w", err)
	}

	// Если первое сообщение — ставим заголовок из текста
	if conv.Title == "Новый чат" {
		title := truncate(message, maxTitleLen)
		_ = u.convRepo.UpdateTitle(ctx, convID, title)
	}

	// Загружаем историю (последние N сообщений, уже включая только что сохранённое)
	history, err := u.convRepo.ListMessages(ctx, convID, maxHistoryMessages)
	if err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage load history: %w", err)
	}

	// Формируем системный промпт с контекстом пользователя
	systemPrompt := u.buildSystemPrompt(ctx, userID)

	// Переводим историю в формат LLM
	msgs := make([]llm.Message, len(history))
	for i, h := range history {
		msgs[i] = llm.Message{Role: h.Role, Content: h.Content}
	}

	// Вызываем LLM
	reply, err := u.llm.Complete(ctx, systemPrompt, msgs)
	if err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage llm: %w", err)
	}

	// Сохраняем ответ ИИ
	aiMsg := &entity.ChatMessage{
		ID:             uuid.New(),
		ConversationID: convID,
		Role:           "assistant",
		Content:        reply,
		CreatedAt:      time.Now(),
	}
	if _, err := u.convRepo.SaveMessage(ctx, aiMsg); err != nil {
		return nil, nil, fmt.Errorf("aiUseCase.SendMessage save ai msg: %w", err)
	}

	_ = u.convRepo.Touch(ctx, convID)

	return userMsg, aiMsg, nil
}

func (u *aiUseCase) ListConversations(ctx context.Context, userID uuid.UUID) ([]*entity.Conversation, error) {
	return u.convRepo.ListByUser(ctx, userID)
}

func (u *aiUseCase) GetConversation(ctx context.Context, id, userID uuid.UUID) (*entity.Conversation, []*entity.ChatMessage, error) {
	conv, err := u.convRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return nil, nil, entity.ErrConversationNotFound
		}
		return nil, nil, fmt.Errorf("aiUseCase.GetConversation: %w", err)
	}
	if conv.UserID != userID {
		return nil, nil, entity.ErrForbidden
	}
	messages, err := u.convRepo.ListMessages(ctx, id, 200)
	if err != nil {
		return nil, nil, err
	}
	return conv, messages, nil
}

func (u *aiUseCase) DeleteConversation(ctx context.Context, id, userID uuid.UUID) error {
	conv, err := u.convRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, entity.ErrNotFound) {
			return entity.ErrConversationNotFound
		}
		return fmt.Errorf("aiUseCase.DeleteConversation: %w", err)
	}
	if conv.UserID != userID {
		return entity.ErrForbidden
	}
	return u.convRepo.Delete(ctx, id)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildSystemPrompt добавляет в базовый промпт контекст пользователя (что читает, вишлист).
func (u *aiUseCase) buildSystemPrompt(ctx context.Context, userID uuid.UUID) string {
	var sb strings.Builder
	sb.WriteString(baseSystemPrompt)

	sessions, _ := u.sessionRepo.ListByUser(ctx, userID)
	wishlist, _ := u.wishlistRepo.ListByUser(ctx, userID)

	var reading, completed, wishTitles []string
	for i, s := range sessions {
		if i >= maxBooksInPrompt {
			break
		}
		book, err := u.bookRepo.GetByID(ctx, s.BookID)
		if err != nil {
			continue
		}
		title := fmt.Sprintf("«%s»", book.Title)
		switch s.Status {
		case entity.StatusReading:
			reading = append(reading, title)
		case entity.StatusFinished:
			completed = append(completed, title)
		}
	}
	for i, w := range wishlist {
		if i >= maxBooksInPrompt {
			break
		}
		if book, err := u.bookRepo.GetByID(ctx, w.BookID); err == nil {
			wishTitles = append(wishTitles, fmt.Sprintf("«%s»", book.Title))
		}
	}

	if len(reading) > 0 || len(completed) > 0 || len(wishTitles) > 0 {
		sb.WriteString("\n\nКонтекст пользователя:")
		if len(reading) > 0 {
			sb.WriteString(fmt.Sprintf("\n- Читает сейчас: %s", strings.Join(reading, ", ")))
		}
		if len(completed) > 0 {
			sb.WriteString(fmt.Sprintf("\n- Прочитал: %s", strings.Join(completed, ", ")))
		}
		if len(wishTitles) > 0 {
			sb.WriteString(fmt.Sprintf("\n- В вишлисте: %s", strings.Join(wishTitles, ", ")))
		}
	}

	return sb.String()
}

func truncate(s string, maxChars int) string {
	if utf8.RuneCountInString(s) <= maxChars {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxChars]) + "…"
}
