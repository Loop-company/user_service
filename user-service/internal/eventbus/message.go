package eventbus

import (
	"time"

	"github.com/google/uuid"
)

// Event - событие для других сервисов (Auth, Chat, Presence, etc.)
type Event struct {
	EventID   string      `json:"event_id"`
	Type      EventType   `json:"type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// EventType - типы событий
type EventType string

const (
	UserRegistered EventType = "user.registered"
	UserUpdated    EventType = "user.updated"
)

// NewEvent - создает событие
func NewEvent(eventType EventType, data interface{}) *Event {
	return &Event{
		EventID:   uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}
}
