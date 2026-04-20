package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/eventbus"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/producer"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/service"
	"github.com/segmentio/kafka-go"
)

// RequestConsumer - обработчик запросов от Gateway
type RequestConsumer struct {
	reader   *kafka.Reader
	service  service.UserService
	producer *producer.ResponseProducer
}

// NewRequestConsumer - создает consumer
func NewRequestConsumer(
	brokers []string,
	groupID string,
	service service.UserService,
	producer *producer.ResponseProducer,
) *RequestConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    "user-service.requests",
		GroupID:  groupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &RequestConsumer{
		reader:   reader,
		service:  service,
		producer: producer,
	}
}

// Start - запускает обработку запросов
func (c *RequestConsumer) Start(ctx context.Context) error {
	slog.Info("Starting request consumer...")

	go func() {
		defer func() {
			if err := c.reader.Close(); err != nil {
				slog.Error("Failed to close request reader", "error", err)
			}
		}()

		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.Error("Failed to fetch message", "error", err)
				time.Sleep(500 * time.Millisecond)
				continue
			}

			if err := c.handleRequest(ctx, msg); err != nil {
				slog.Error("Failed to handle request", "error", err)
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				slog.Error("Failed to commit message", "error", err)
			}
		}
	}()

	return nil
}

// handleRequest - обрабатывает один запрос
func (c *RequestConsumer) handleRequest(ctx context.Context, msg kafka.Message) error {
	var req eventbus.Request
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		slog.Error("Failed to unmarshal request", "error", err)
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	slog.Info("Received request",
		"correlation_id", req.CorrelationID,
		"message_type", req.MessageType,
	)

	var response *eventbus.Response

	switch req.MessageType {
	case "get_profile":
		response = c.handleGetProfile(ctx, &req)
	case "update_name":
		response = c.handleUpdateName(ctx, &req)
	case "update_status":
		response = c.handleUpdateStatus(ctx, &req)
	case "get_settings":
		response = c.handleGetSettings(ctx, &req)
	case "update_settings":
		response = c.handleUpdateSettings(ctx, &req)
	case "update_setting_key":
		response = c.handleUpdateSettingKey(ctx, &req)
	default:
		response = eventbus.NewResponse(
			req.CorrelationID,
			req.MessageType,
			false,
			nil,
			fmt.Sprintf("Unknown message type: %s", req.MessageType),
		)
	}

	if err := c.producer.SendResponse(ctx, response); err != nil {
		slog.Error("Failed to send response to producer", "error", err)
		return err
	}

	return nil
}

func parsePayload[T any](raw json.RawMessage) (T, error) {
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		return v, fmt.Errorf("invalid payload: %w", err)
	}
	return v, nil
}

func (c *RequestConsumer) handleGetProfile(ctx context.Context, req *eventbus.Request) *eventbus.Response {
	payload, err := parsePayload[struct {
		UserID string `json:"user_id"`
	}](req.Payload)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	profile, err := c.service.GetProfile(ctx, payload.UserID)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	return eventbus.NewResponse(req.CorrelationID, req.MessageType, true, profile, "")
}

func (c *RequestConsumer) handleUpdateName(ctx context.Context, req *eventbus.Request) *eventbus.Response {
	payload, err := parsePayload[struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
	}](req.Payload)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	if err := c.service.UpdateName(ctx, payload.UserID, payload.Name); err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	return eventbus.NewResponse(req.CorrelationID, req.MessageType, true, nil, "")
}

func (c *RequestConsumer) handleUpdateStatus(ctx context.Context, req *eventbus.Request) *eventbus.Response {
	payload, err := parsePayload[struct {
		UserID string `json:"user_id"`
		Status string `json:"status"`
	}](req.Payload)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	if err := c.service.UpdateStatus(ctx, payload.UserID, payload.Status); err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	return eventbus.NewResponse(req.CorrelationID, req.MessageType, true, nil, "")
}

func (c *RequestConsumer) handleGetSettings(ctx context.Context, req *eventbus.Request) *eventbus.Response {
	payload, err := parsePayload[struct {
		UserID string `json:"user_id"`
	}](req.Payload)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	settings, err := c.service.GetSettings(ctx, payload.UserID)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	return eventbus.NewResponse(req.CorrelationID, req.MessageType, true, settings, "")
}

func (c *RequestConsumer) handleUpdateSettings(ctx context.Context, req *eventbus.Request) *eventbus.Response {
	payload, err := parsePayload[struct {
		UserID   string                 `json:"user_id"`
		Settings map[string]interface{} `json:"settings"`
	}](req.Payload)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	if err := c.service.UpdateSettings(ctx, payload.UserID, payload.Settings); err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	return eventbus.NewResponse(req.CorrelationID, req.MessageType, true, nil, "")
}

func (c *RequestConsumer) handleUpdateSettingKey(ctx context.Context, req *eventbus.Request) *eventbus.Response {
	payload, err := parsePayload[struct {
		UserID string      `json:"user_id"`
		Key    string      `json:"key"`
		Value  interface{} `json:"value"`
	}](req.Payload)
	if err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	if err := c.service.UpdateSettingKey(ctx, payload.UserID, payload.Key, payload.Value); err != nil {
		return eventbus.NewResponse(req.CorrelationID, req.MessageType, false, nil, err.Error())
	}

	return eventbus.NewResponse(req.CorrelationID, req.MessageType, true, nil, "")
}
