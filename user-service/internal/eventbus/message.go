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

type Request struct {
	CorrelationID string      `json:"correlation_id"`
	MessageType   string      `json:"message_type"`
	Timestamp     string      `json:"timestamp"`
	Data          interface{} `json:"data"`
	ReplyTo       string      `json:"reply_to"`
}

type Response struct {
	CorrelationID string      `json:"correlation_id"`
	MessageType   string      `json:"message_type"`
	Timestamp     string      `json:"timestamp"`
	Success       bool        `json:"success"`
	Data          interface{} `json:"data,omitempty"`
	Error         string      `json:"error,omitempty"`
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

func NewRequest(messageType string, data interface{}, replyTo string) *Request {
	return &Request{
		CorrelationID: uuid.New().String(),
		MessageType:   messageType,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Data:          data,
		ReplyTo:       replyTo,
	}
}

func NewResponse(correlationID, messageType string, success bool, data interface{}, errorMessage string) *Response {
	return &Response{
		CorrelationID: correlationID,
		MessageType:   messageType,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Success:       success,
		Data:          data,
		Error:         errorMessage,
	}
}
