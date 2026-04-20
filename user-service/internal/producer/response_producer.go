package producer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/eventbus"
	"github.com/segmentio/kafka-go"
)

// ResponseProducer - отправляет ответы в API Gateway
type ResponseProducer struct {
	writer *kafka.Writer
}

// NewResponseProducer - создает продюсер ответов
func NewResponseProducer(brokers []string) *ResponseProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    "user-service.responses",
		Balancer: &kafka.LeastBytes{},
	}

	return &ResponseProducer{writer: writer}
}

// SendResponse - отправляет ответ в Gateway
func (p *ResponseProducer) SendResponse(ctx context.Context, response *eventbus.Response) error {
	bytes, err := json.Marshal(response)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(response.CorrelationID),
		Value: bytes,
	}

	err = p.writer.WriteMessages(ctx, msg)
	if err != nil {
		slog.Error("Failed to send response",
			"correlation_id", response.CorrelationID,
			"error", err,
		)
		return err
	}

	slog.Debug("Response sent",
		"correlation_id", response.CorrelationID,
		"success", response.Success,
	)
	return nil
}

// Close - закрывает writer
func (p *ResponseProducer) Close() error {
	return p.writer.Close()
}
