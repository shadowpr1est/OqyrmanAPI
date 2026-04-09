package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

// DeadlineReminder — фоновый воркер, который уведомляет пользователей
// о приближающемся сроке возврата/брони за 2 дня до due_date.
type DeadlineReminder struct {
	repo      repository.ReservationRepository
	notifRepo repository.NotificationRepository
	hub       Broadcaster
	interval  time.Duration
	lookahead time.Duration // e.g. 48h
}

func NewDeadlineReminder(
	repo repository.ReservationRepository,
	notifRepo repository.NotificationRepository,
	hub Broadcaster,
	interval time.Duration,
	lookahead time.Duration,
) *DeadlineReminder {
	return &DeadlineReminder{
		repo:      repo,
		notifRepo: notifRepo,
		hub:       hub,
		interval:  interval,
		lookahead: lookahead,
	}
}

func (w *DeadlineReminder) Run(ctx context.Context) {
	slog.InfoContext(ctx, "DeadlineReminder: started", "lookahead", w.lookahead)

	iterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	w.runOnce(iterCtx)
	cancel()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			iterCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			w.runOnce(iterCtx)
			cancel()
		case <-ctx.Done():
			slog.InfoContext(ctx, "DeadlineReminder: stopped")
			return
		}
	}
}

func (w *DeadlineReminder) runOnce(ctx context.Context) {
	approaching, err := w.repo.FindApproachingDeadline(ctx, w.lookahead)
	if err != nil {
		slog.ErrorContext(ctx, "DeadlineReminder: error finding approaching deadlines", "err", err)
		return
	}

	for _, res := range approaching {
		nType := entity.NotifReservationDeadline
		title := "Срок брони истекает"
		body := "Срок вашего бронирования истекает " + res.DueDate.Format("02.01.2006") + ". Не забудьте забрать книгу."

		if res.Status == entity.ReservationActive {
			nType = entity.NotifReturnDeadline
			title = "Срок возврата приближается"
			body = "Верните книгу до " + res.DueDate.Format("02.01.2006") + ", чтобы избежать штрафных санкций."
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
			slog.ErrorContext(ctx, "DeadlineReminder: failed to create notification", "user_id", res.UserID, "err", err)
			continue
		}
		if w.hub != nil {
			w.hub.Send(res.UserID, saved)
		}
	}

	if len(approaching) > 0 {
		slog.InfoContext(ctx, "DeadlineReminder: sent deadline reminders", "count", len(approaching))
	}
}
