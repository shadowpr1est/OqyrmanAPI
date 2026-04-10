package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotifReservationSuccess  NotificationType = "reservation_success"
	NotifPickupSuccess       NotificationType = "pickup_success"
	NotifReservationDeadline NotificationType = "reservation_deadline"
	NotifReturnDeadline      NotificationType = "return_deadline"
	NotifReservationExpired  NotificationType = "reservation_expired"
	NotifReturnOverdue       NotificationType = "return_overdue"
	NotifEventReminder       NotificationType = "event_reminder"
	NotifGeneral             NotificationType = "general"
)

type Notification struct {
	ID        uuid.UUID        `db:"id"`
	UserID    uuid.UUID        `db:"user_id"`
	Type      NotificationType `db:"type"`
	Title     string           `db:"title"`
	Body      string           `db:"body"`
	IsRead    bool             `db:"is_read"`
	CreatedAt time.Time        `db:"created_at"`
	ReadAt    *time.Time       `db:"read_at"`
}
