package eventbus

import (
	"testing"
	"time"
)

func TestNewRequest(t *testing.T) {
	req := NewRequest("get_profile", map[string]string{"user_id": "123"}, "user-service.responses")

	if req.CorrelationID == "" {
		t.Fatal("expected correlation ID to be generated")
	}
	if req.MessageType != "get_profile" {
		t.Fatalf("expected message type %q, got %q", "get_profile", req.MessageType)
	}
	if req.ReplyTo != "user-service.responses" {
		t.Fatalf("expected reply topic %q, got %q", "user-service.responses", req.ReplyTo)
	}
	if _, err := time.Parse(time.RFC3339, req.Timestamp); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %v", err)
	}
}

func TestNewResponse(t *testing.T) {
	resp := NewResponse("corr-1", "get_profile", true, map[string]string{"id": "123"}, "")

	if resp.CorrelationID != "corr-1" {
		t.Fatalf("expected correlation ID %q, got %q", "corr-1", resp.CorrelationID)
	}
	if !resp.Success {
		t.Fatal("expected success to be true")
	}
	if _, err := time.Parse(time.RFC3339, resp.Timestamp); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %v", err)
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent(UserRegistered, map[string]string{"user_id": "123"})

	if event.EventID == "" {
		t.Fatal("expected event ID to be generated")
	}
	if event.Type != UserRegistered {
		t.Fatalf("expected event type %q, got %q", UserRegistered, event.Type)
	}
	if _, err := time.Parse(time.RFC3339, event.Timestamp); err != nil {
		t.Fatalf("expected RFC3339 timestamp, got %v", err)
	}
}
