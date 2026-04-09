package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

// Broadcaster is satisfied by *hub.NotificationHub.
type Broadcaster interface {
	Send(userID uuid.UUID, n *entity.Notification)
}

// OverdueCanceller — фоновый воркер, который периодически отменяет просроченные брони.
// Для каждой брони где due_date < now() AND status IN ('active', 'pending'):
//   - меняет статус на cancelled
//   - восстанавливает available_copies
//   - уведомляет пользователей
//
// Всё выполняется в одной DB-транзакции.
type OverdueCanceller struct {
	repo      repository.ReservationRepository
	notifRepo repository.NotificationRepository
	hub       Broadcaster
	interval  time.Duration
}

func NewOverdueCanceller(
	repo repository.ReservationRepository,
	notifRepo repository.NotificationRepository,
	hub Broadcaster,
	interval time.Duration,
) *OverdueCanceller {
	return &OverdueCanceller{repo: repo, notifRepo: notifRepo, hub: hub, interval: interval}
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
	cancelled, err := w.repo.CancelOverdue(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "OverdueCanceller: error cancelling overdue reservations", "err", err)
		return
	}
	if len(cancelled) > 0 {
		slog.InfoContext(ctx, "OverdueCanceller: cancelled overdue reservations", "count", len(cancelled))
	}
	for _, res := range cancelled {
		nType := entity.NotifReservationExpired
		title := "Бронирование просрочено"
		body := "Ваше бронирование было автоматически отменено из-за истечения срока."
		if res.Status == entity.ReservationActive {
			nType = entity.NotifReturnOverdue
			title = "Просрочен возврат книги"
			body = "Срок возврата книги истёк. Возможны штрафные санкции. Пожалуйста, верните книгу как можно скорее."
		}
		n := &entity.Notification{
			ID:        uuid.New(),
			UserID:    res.UserID,
			Type:      nType,
			Title:     title,
			Body:      body,
			CreatedAt: time.Now(),
		}
		saved, err := w.notifRepo.Create(ctx, n)
		if err != nil {
			slog.ErrorContext(ctx, "OverdueCanceller: failed to create notification", "user_id", res.UserID, "err", err)
			continue
		}
		if w.hub != nil {
			w.hub.Send(res.UserID, saved)
		}
	}
}
