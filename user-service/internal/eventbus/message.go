package eventbus

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Request - запрос от API Gateway к User Service
type Request struct {
	CorrelationID string          `json:"correlation_id"` // ID для связи запроса и ответа
	ReplyTo       string          `json:"reply_to"`       // Топик для ответа (user-service.responses)
	MessageType   string          `json:"message_type"`   // Тип: get_profile, update_name, etc.
	Payload       json.RawMessage `json:"payload"`        // Данные запроса
	Timestamp     string          `json:"timestamp"`
}

// Response - ответ от User Service к API Gateway
type Response struct {
	CorrelationID string      `json:"correlation_id"` // Тот же ID что в запросе
	MessageType   string      `json:"message_type"`
	Success       bool        `json:"success"`         // Успешно или нет
	Data          interface{} `json:"data,omitempty"`  // Данные ответа
	Error         string      `json:"error,omitempty"` // Сообщение об ошибке
	Timestamp     string      `json:"timestamp"`
}

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
	UserRegistered    EventType = "user.registered"
	UserUpdated       EventType = "user.updated"
	UserStatusChanged EventType = "user.status_changed"
)

// NewRequest - создает новый запрос
func NewRequest(messageType string, payload interface{}, replyTo string) *Request {
	payloadBytes, _ := json.Marshal(payload)

	return &Request{
		CorrelationID: uuid.New().String(),
		ReplyTo:       replyTo,
		MessageType:   messageType,
		Payload:       payloadBytes,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// NewResponse - создает ответ
func NewResponse(correlationID, messageType string, success bool, data interface{}, errMsg string) *Response {
	return &Response{
		CorrelationID: correlationID,
		MessageType:   messageType,
		Success:       success,
		Data:          data,
		Error:         errMsg,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}
}

// NewEvent - создает событие
func NewEvent(eventType EventType, data interface{}) *Event {
	return &Event{
		EventID:   uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}
}
