package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/entity"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/eventbus"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/producer"
	servicepkg "github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/service"
	"github.com/segmentio/kafka-go"
)

type userServiceMock struct {
	servicepkg.UserService
	getProfileFunc     func(ctx context.Context, userID string) (*entity.UserProfile, error)
	updateNameFunc     func(ctx context.Context, userID, name string) error
	updateStatusFunc   func(ctx context.Context, userID, status string) error
	getSettingsFunc    func(ctx context.Context, userID string) (*entity.UserSettings, error)
	updateSettingsFunc func(ctx context.Context, userID string, settings map[string]interface{}) error
	updateSettingKeyFn func(ctx context.Context, userID, key string, value interface{}) error
}

func (m *userServiceMock) GetProfile(ctx context.Context, userID string) (*entity.UserProfile, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return nil, nil
}

func (m *userServiceMock) UpdateName(ctx context.Context, userID, name string) error {
	if m.updateNameFunc != nil {
		return m.updateNameFunc(ctx, userID, name)
	}
	return nil
}

func (m *userServiceMock) UpdateStatus(ctx context.Context, userID, status string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, userID, status)
	}
	return nil
}

func (m *userServiceMock) GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error) {
	if m.getSettingsFunc != nil {
		return m.getSettingsFunc(ctx, userID)
	}
	return nil, nil
}

func (m *userServiceMock) UpdateSettings(ctx context.Context, userID string, settings map[string]interface{}) error {
	if m.updateSettingsFunc != nil {
		return m.updateSettingsFunc(ctx, userID, settings)
	}
	return nil
}

func (m *userServiceMock) UpdateSettingKey(ctx context.Context, userID, key string, value interface{}) error {
	if m.updateSettingKeyFn != nil {
		return m.updateSettingKeyFn(ctx, userID, key, value)
	}
	return nil
}

func TestParsePayloadSuccess(t *testing.T) {
	payload := []byte(`{"user_id":"123","name":"Alice"}`)

	parsed, err := parsePayload[struct {
		UserID string `json:"user_id"`
		Name   string `json:"name"`
	}](payload)
	if err != nil {
		t.Fatalf("parsePayload returned error: %v", err)
	}

	if parsed.UserID != "123" || parsed.Name != "Alice" {
		t.Fatalf("unexpected parsed payload: %#v", parsed)
	}
}

func TestParsePayloadInvalidJSON(t *testing.T) {
	if _, err := parsePayload[struct {
		UserID string `json:"user_id"`
	}]([]byte(`{"user_id":`)); err == nil {
		t.Fatal("expected parsePayload to fail for invalid JSON")
	}
}

func TestHandleGetProfile(t *testing.T) {
	service := &userServiceMock{
		getProfileFunc: func(ctx context.Context, userID string) (*entity.UserProfile, error) {
			return &entity.UserProfile{ID: userID, Name: "Alice"}, nil
		},
	}
	consumer := &RequestConsumer{service: service}

	req := &eventbus.Request{
		CorrelationID: "corr-1",
		MessageType:   "get_profile",
		Payload:       json.RawMessage(`{"user_id":"123"}`),
	}

	resp := consumer.handleGetProfile(context.Background(), req)
	if !resp.Success {
		t.Fatalf("expected success response, got %#v", resp)
	}
}

func TestHandleUpdateNameReturnsServiceError(t *testing.T) {
	service := &userServiceMock{
		updateNameFunc: func(ctx context.Context, userID, name string) error {
			return errors.New("validation failed")
		},
	}
	consumer := &RequestConsumer{service: service}

	req := &eventbus.Request{
		CorrelationID: "corr-2",
		MessageType:   "update_name",
		Payload:       json.RawMessage(`{"user_id":"123","name":"A"}`),
	}

	resp := consumer.handleUpdateName(context.Background(), req)
	if resp.Success {
		t.Fatalf("expected error response, got %#v", resp)
	}
	if resp.Error == "" {
		t.Fatal("expected error message to be set")
	}
}

func TestHandleGetSettings(t *testing.T) {
	service := &userServiceMock{
		getSettingsFunc: func(ctx context.Context, userID string) (*entity.UserSettings, error) {
			return &entity.UserSettings{
				UserID: userID,
				Settings: map[string]interface{}{
					"appearance": map[string]interface{}{"theme": "dark"},
				},
			}, nil
		},
	}
	consumer := &RequestConsumer{service: service}

	req := &eventbus.Request{
		CorrelationID: "corr-3",
		MessageType:   "get_settings",
		Payload:       json.RawMessage(`{"user_id":"123"}`),
	}

	resp := consumer.handleGetSettings(context.Background(), req)
	if !resp.Success {
		t.Fatalf("expected success response, got %#v", resp)
	}
}

func TestHandleUpdateSettings(t *testing.T) {
	called := false
	service := &userServiceMock{
		updateSettingsFunc: func(ctx context.Context, userID string, settings map[string]interface{}) error {
			called = true
			return nil
		},
	}
	consumer := &RequestConsumer{service: service}

	req := &eventbus.Request{
		CorrelationID: "corr-4",
		MessageType:   "update_settings",
		Payload:       json.RawMessage(`{"user_id":"123","settings":{"appearance":{"theme":"light"}}}`),
	}

	resp := consumer.handleUpdateSettings(context.Background(), req)
	if !resp.Success {
		t.Fatalf("expected success response, got %#v", resp)
	}
	if !called {
		t.Fatal("expected UpdateSettings to be called")
	}
}

func TestHandleUpdateSettingKey(t *testing.T) {
	called := false
	service := &userServiceMock{
		updateSettingKeyFn: func(ctx context.Context, userID, key string, value interface{}) error {
			called = true
			return nil
		},
	}
	consumer := &RequestConsumer{service: service}

	req := &eventbus.Request{
		CorrelationID: "corr-5",
		MessageType:   "update_setting_key",
		Payload:       json.RawMessage(`{"user_id":"123","key":"appearance.theme","value":"dark"}`),
	}

	resp := consumer.handleUpdateSettingKey(context.Background(), req)
	if !resp.Success {
		t.Fatalf("expected success response, got %#v", resp)
	}
	if !called {
		t.Fatal("expected UpdateSettingKey to be called")
	}
}

func TestHandleUpdateStatus(t *testing.T) {
	called := false
	service := &userServiceMock{
		updateStatusFunc: func(ctx context.Context, userID, status string) error {
			called = true
			return nil
		},
	}
	consumer := &RequestConsumer{service: service}

	req := &eventbus.Request{
		CorrelationID: "corr-6",
		MessageType:   "update_status",
		Payload:       json.RawMessage(`{"user_id":"123","status":"online"}`),
	}

	resp := consumer.handleUpdateStatus(context.Background(), req)
	if !resp.Success {
		t.Fatalf("expected success response, got %#v", resp)
	}
	if !called {
		t.Fatal("expected UpdateStatus to be called")
	}
}

func TestNewRequestConsumer(t *testing.T) {
	service := &userServiceMock{}
	responseProducer := producer.NewResponseProducer([]string{"localhost:9092"})
	defer responseProducer.Close()

	consumer := NewRequestConsumer([]string{"localhost:9092"}, "group-1", service, responseProducer)
	if consumer == nil {
		t.Fatal("expected consumer to be created")
	}
	if consumer.reader == nil {
		t.Fatal("expected kafka reader to be initialized")
	}
}

func TestHandleRequestUnknownTypeReturnsProducerError(t *testing.T) {
	service := &userServiceMock{}
	responseProducer := producer.NewResponseProducer([]string{"127.0.0.1:1"})
	defer responseProducer.Close()

	consumer := &RequestConsumer{
		service:  service,
		producer: responseProducer,
	}

	req := eventbus.NewRequest("unknown_type", map[string]string{"user_id": "123"}, "user-service.responses")
	msg := kafka.Message{Value: mustMarshalRequest(t, req)}

	if err := consumer.handleRequest(context.Background(), msg); err == nil {
		t.Fatal("expected producer send to fail on invalid kafka address")
	}
}

func mustMarshalRequest(t *testing.T, req *eventbus.Request) []byte {
	t.Helper()

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	return data
}
