package ai

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
	domainUseCase "github.com/shadowpr1est/OqyrmanAPI/internal/domain/usecase"
	"github.com/shadowpr1est/OqyrmanAPI/pkg/llm"
)

const systemPrompt = "Ты — книжный ассистент платформы Oqyrman. Отвечай на русском языке."

type aiUseCase struct {
	sessionRepo  repository.ReadingSessionRepository
	wishlistRepo repository.WishlistRepository
	llm          llm.LLMClient
}

func NewAIUseCase(
	sessionRepo repository.ReadingSessionRepository,
	wishlistRepo repository.WishlistRepository,
	llm llm.LLMClient,
) domainUseCase.AIUseCase {
	return &aiUseCase{sessionRepo: sessionRepo, wishlistRepo: wishlistRepo, llm: llm}
}

func (u *aiUseCase) Recommend(ctx context.Context, userIDStr string) (string, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend invalid userID: %w", err)
	}

	// FIX: ошибки игнорировались через `_`.
	// При недоступной БД usecase получал пустые данные и генерировал
	// бессмысленные рекомендации без каких-либо признаков ошибки.
	// Теперь ошибки возвращаются — клиент получает 500 вместо мусора.
	sessions, err := u.sessionRepo.ListByUser(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend get sessions: %w", err)
	}

	wishlist, err := u.wishlistRepo.ListByUser(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("aiUseCase.Recommend get wishlist: %w", err)
	}

	prompt := fmt.Sprintf(
		"Дай 5 персональных рекомендаций книг. У пользователя %d книг в истории чтения и %d в вишлисте.",
		len(sessions), len(wishlist),
	)

	return u.llm.Complete(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: prompt},
	})
}

func (u *aiUseCase) Chat(ctx context.Context, message string) (string, error) {
	return u.llm.Complete(ctx, systemPrompt, []llm.Message{
		{Role: "user", Content: message},
	})
}
