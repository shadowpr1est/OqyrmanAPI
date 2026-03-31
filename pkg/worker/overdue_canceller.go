package worker

import (
	"context"
	"log"
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
	log.Println("OverdueCanceller: started")

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
			log.Println("OverdueCanceller: stopped")
			return
		}
	}
}

func (w *OverdueCanceller) runOnce(ctx context.Context) {
	count, err := w.repo.CancelOverdue(ctx)
	if err != nil {
		log.Printf("OverdueCanceller: error cancelling overdue reservations: %v", err)
		return
	}
	if count > 0 {
		log.Printf("OverdueCanceller: cancelled %d overdue reservation(s), copies restored", count)
	}
}
