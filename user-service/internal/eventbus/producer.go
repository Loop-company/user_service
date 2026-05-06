package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type EventType string

const (
	UserProfileUpdated EventType = "user.profile_updated"
	UserSettingsUpdated EventType = "user.settings_updated"
)

type Event struct {
	EventID   string      `json:"event_id"`
	Type      EventType   `json:"type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"data"`
}

type KafkaProducer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewKafkaProducer(brokers []string, log *slog.Logger) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: false,
		},
		log: log,
	}
}

func (p *KafkaProducer) Close() error {
	return p.writer.Close()
}

func (p *KafkaProducer) publish(ctx context.Context, topic string, eventType EventType, data interface{}) error {
	event := Event{
		EventID:   uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: payload,
	})

	if err != nil {
		p.log.Error("failed to publish message", "topic", topic, "type", eventType, "error", err)
		return err
	}

	p.log.Info("published message", "topic", topic, "type", eventType)
	return nil
}

func (p *KafkaProducer) SendUserProfileUpdated(ctx context.Context, userID string, changes map[string]interface{}) {
	data := map[string]interface{}{
		"user_id": userID,
		"changes": changes,
	}

	// Send to analytics service
	_ = p.publish(ctx, "analytics.events", UserProfileUpdated, data)
}

func (p *KafkaProducer) SendUserSettingsUpdated(ctx context.Context, userID string, settings map[string]interface{}) {
	data := map[string]interface{}{
		"user_id":  userID,
		"settings": settings,
	}

	// Send to analytics service
	_ = p.publish(ctx, "analytics.events", UserSettingsUpdated, data)
}
