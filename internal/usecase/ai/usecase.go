package ai

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm"
)

const (
	systemPrompt     = "Ты — книжный ассистент платформы Oqyrman. Отвечай на русском языке."
	maxBooksInPrompt = 10 // лимит книг для промпта
)

type aiUseCase struct {
	sessionRepo  repository.ReadingSessionRepository
	wishlistRepo repository.WishlistRepository
	bookRepo     repository.BookRepository // ДОБАВИТЬ
	llm          llm.LLMClient
}

func NewAIUseCase(
	sessionRepo repository.ReadingSessionRepository,
	wishlistRepo repository.WishlistRepository,
	bookRepo repository.BookRepository, // ДОБАВИТЬ
	llm llm.LLMClient,
) domainUseCase.AIUseCase {
	return &aiUseCase{
		sessionRepo:  sessionRepo,
		wishlistRepo: wishlistRepo,
		bookRepo:     bookRepo,
		llm:          llm,
	}
}

func (u *aiUseCase) Recommend(ctx context.Context, userIDStr string) (string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend invalid userID: %w", err)
	}

	sessions, err := u.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend get sessions: %w", err)
	}

	wishlist, err := u.wishlistRepo.ListByUser(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend get wishlist: %w", err)
	}

	// Собираем названия прочитанных книг
	readTitles := make([]string, 0, min(len(sessions), maxBooksInPrompt))
	for i, s := range sessions {
		if i >= maxBooksInPrompt {
			break
		}
		book, err := u.bookRepo.GetByID(ctx, s.BookID)
		if err != nil {
			continue
		}
		readTitles = append(readTitles, fmt.Sprintf("«%s»", book.Title))
	}

	// Собираем названия книг из вишлиста
	wishTitles := make([]string, 0, min(len(wishlist), maxBooksInPrompt))
	for i, w := range wishlist {
		if i >= maxBooksInPrompt {
			break
		}
		book, err := u.bookRepo.GetByID(ctx, w.BookID)
		if err != nil {
			continue
		}
		wishTitles = append(wishTitles, fmt.Sprintf("«%s»", book.Title))
	}

	var prompt strings.Builder
	prompt.WriteString("Дай 5 персональных рекомендаций книг для чтения.\n\n")

	if len(readTitles) > 0 {
		prompt.WriteString(fmt.Sprintf("История чтения пользователя: %s.\n", strings.Join(readTitles, ", ")))
	} else {
		prompt.WriteString("История чтения пользователя пуста.\n")
	}

	if len(wishTitles) > 0 {
		prompt.WriteString(fmt.Sprintf("Вишлист пользователя: %s.\n", strings.Join(wishTitles, ", ")))
	} else {
		prompt.WriteString("Вишлист пользователя пуст.\n")
	}

	prompt.WriteString("\nПорекомендуй 5 книг которые могут понравиться, объясни кратко почему каждая.")

	return u.llm.Complete(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: prompt.String()},
	})
}

func (u *aiUseCase) Chat(ctx context.Context, message string) (string, error) {
	return u.llm.Complete(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: message},
	})
}
