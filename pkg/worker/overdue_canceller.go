package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

// OverdueCanceller — фоновый воркер, который периодически отменяет просроченные брони.
// Для каждой брони где due_date < now() AND status = 'active':
//   - меняет статус на cancelled
//   - восстанавливает available_copies
//
// Всё выполняется в одной DB-транзакции.
type OverdueCanceller struct {
	repo     repository.ReservationRepository
	interval time.Duration
}

func NewOverdueCanceller(repo repository.ReservationRepository, interval time.Duration) *OverdueCanceller {
	return &OverdueCanceller{repo: repo, interval: interval}
}

// Run запускает воркер. Блокируется до отмены контекста.
// Предназначен для запуска в отдельной горутине: go canceller.Run(ctx)
func (w *OverdueCanceller) Run(ctx context.Context) {
	slog.InfoContext(ctx, "OverdueCanceller: started")

	// Запуск сразу при старте — не ждём первого тика
	iterCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	w.runOnce(iterCtx)
	cancel()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			iterCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			w.runOnce(iterCtx)
			cancel()
		case <-ctx.Done():
			slog.InfoContext(ctx, "OverdueCanceller: stopped")
			return
		}
	}
}

func (w *OverdueCanceller) runOnce(ctx context.Context) {
	count, err := w.repo.CancelOverdue(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "OverdueCanceller: error cancelling overdue reservations", "err", err)
		return
	}
	if count > 0 {
		slog.InfoContext(ctx, "OverdueCanceller: cancelled overdue reservations", "count", count)
	}
}
