package hub

import (
	"sync"

	"github.com/google/uuid"
	"github.com/shadowpr1est/OqyrmanAPI/internal/domain/entity"
)

// NotificationHub distributes notifications to SSE subscribers in-process.
type NotificationHub struct {
	mu   sync.RWMutex
	subs map[uuid.UUID]map[string]chan *entity.Notification
}

func New() *NotificationHub {
	return &NotificationHub{
		subs: make(map[uuid.UUID]map[string]chan *entity.Notification),
	}
}

// Subscribe registers a channel for userID. Returns connID and unsubscribe func.
func (h *NotificationHub) Subscribe(userID uuid.UUID) (string, <-chan *entity.Notification, func()) {
	connID := uuid.New().String()
	ch := make(chan *entity.Notification, 8)

	h.mu.Lock()
	if h.subs[userID] == nil {
		h.subs[userID] = make(map[string]chan *entity.Notification)
	}
	h.subs[userID][connID] = ch
	h.mu.Unlock()

	unsub := func() {
		h.mu.Lock()
		delete(h.subs[userID], connID)
		if len(h.subs[userID]) == 0 {
			delete(h.subs, userID)
		}
		h.mu.Unlock()
		close(ch)
	}
	return connID, ch, unsub
}

// Send broadcasts a notification to all subscribers of userID (non-blocking).
func (h *NotificationHub) Send(userID uuid.UUID, n *entity.Notification) {
	h.mu.RLock()
	conns := h.subs[userID]
	h.mu.RUnlock()

	for _, ch := range conns {
		select {
		case ch <- n:
		default: // skip if subscriber is slow
		}
	}
}
