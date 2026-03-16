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
		return "", err
	}

	sessions, _ := u.sessionRepo.ListByUser(ctx, userID)
	wishlist, _ := u.wishlistRepo.ListByUser(ctx, userID)

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
