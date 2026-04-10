package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/repository"
)

// UserIDLister is satisfied by the user repository.
type UserIDLister interface {
	ListAllIDs(ctx context.Context) ([]uuid.UUID, error)
}

// EventReminder — фоновый воркер, который уведомляет всех пользователей
// о предстоящих мероприятиях за 24 часа до начала.
type EventReminder struct {
	eventRepo repository.EventRepository
	userRepo  UserIDLister
	notifRepo repository.NotificationRepository
	hub       Broadcaster
	interval  time.Duration
	lookahead time.Duration // e.g. 24h
}

func NewEventReminder(
	eventRepo repository.EventRepository,
	userRepo UserIDLister,
	notifRepo repository.NotificationRepository,
	hub Broadcaster,
	interval time.Duration,
	lookahead time.Duration,
) *EventReminder {
	return &EventReminder{
		eventRepo: eventRepo,
		userRepo:  userRepo,
		notifRepo: notifRepo,
		hub:       hub,
		interval:  interval,
		lookahead: lookahead,
	}
}

func (w *EventReminder) Run(ctx context.Context) {
	slog.InfoContext(ctx, "EventReminder: started", "lookahead", w.lookahead)

	iterCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	w.runOnce(iterCtx)
	cancel()

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			iterCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			w.runOnce(iterCtx)
			cancel()
		case <-ctx.Done():
			slog.InfoContext(ctx, "EventReminder: stopped")
			return
		}
	}
}

func (w *EventReminder) runOnce(ctx context.Context) {
	events, err := w.eventRepo.FindUpcoming(ctx, w.lookahead)
	if err != nil {
		slog.ErrorContext(ctx, "EventReminder: error finding upcoming events", "err", err)
		return
	}
	if len(events) == 0 {
		return
	}

	userIDs, err := w.userRepo.ListAllIDs(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "EventReminder: error listing users", "err", err)
		return
	}

	var totalSent int
	for _, event := range events {
		title := "Завтра мероприятие!"
		body := fmt.Sprintf(
			"«%s» начнётся %s",
			event.Title,
			event.StartsAt.Format("02.01.2006 в 15:04"),
		)
		if event.Location != nil && *event.Location != "" {
			body += fmt.Sprintf(". Место: %s", *event.Location)
		}

		for _, uid := range userIDs {
			n := &entity.Notification{
				ID:        uuid.New(),
				UserID:    uid,
				Type:      entity.NotifEventReminder,
				Title:     title,
				Body:      body,
				CreatedAt: time.Now(),
			}
			saved, err := w.notifRepo.Create(ctx, n)
			if err != nil {
				slog.ErrorContext(ctx, "EventReminder: failed to create notification",
					"user_id", uid, "event_id", event.ID, "err", err)
				continue
			}
			if w.hub != nil {
				w.hub.Send(uid, saved)
			}
			totalSent++
		}

		if err := w.eventRepo.MarkReminderSent(ctx, event.ID); err != nil {
			slog.ErrorContext(ctx, "EventReminder: failed to mark reminder sent",
				"event_id", event.ID, "err", err)
		}
	}

	slog.InfoContext(ctx, "EventReminder: sent event reminders",
		"events", len(events), "notifications", totalSent)
}
