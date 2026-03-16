package usecase

import "context"

type AIUseCase interface {
	Recommend(ctx context.Context, userID string) (string, error)
	Chat(ctx context.Context, message string) (string, error)
}
