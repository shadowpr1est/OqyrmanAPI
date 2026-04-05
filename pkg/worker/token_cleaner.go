package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

// TokenCleaner — фоновый воркер, который периодически удаляет истёкшие refresh-токены.
// Без этого таблица tokens растёт бесконечно при ротации токенов тысячами пользователей.
type TokenCleaner struct {
	repo     repository.TokenRepository
	interval time.Duration
}

func NewTokenCleaner(repo repository.TokenRepository, interval time.Duration) *TokenCleaner {
	return &TokenCleaner{repo: repo, interval: interval}
}

// Run запускает воркер. Блокируется до отмены контекста.
// Предназначен для запуска в отдельной горутине: go cleaner.Run(ctx)
func (w *TokenCleaner) Run(ctx context.Context) {
	slog.InfoContext(ctx, "TokenCleaner: started")

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			iterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			n, err := w.repo.DeleteExpired(iterCtx)
			cancel()
			if err != nil {
				slog.ErrorContext(ctx, "TokenCleaner: error deleting expired tokens", "err", err)
			} else if n > 0 {
				slog.InfoContext(ctx, "TokenCleaner: deleted expired tokens", "count", n)
			}
		case <-ctx.Done():
			slog.InfoContext(ctx, "TokenCleaner: stopped")
			return
		}
	}
}
