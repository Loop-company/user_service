package producer

import (
	"context"
	"testing"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/eventbus"
)

func TestNewResponseProducer(t *testing.T) {
	producer := NewResponseProducer([]string{"localhost:9092"})

	if producer == nil {
		t.Fatal("expected response producer to be created")
	}
	if producer.writer == nil {
		t.Fatal("expected kafka writer to be initialized")
	}

	if err := producer.Close(); err != nil {
		t.Fatalf("expected Close to succeed, got %v", err)
	}
}

func TestSendResponseReturnsErrorWhenWriterUnavailable(t *testing.T) {
	producer := NewResponseProducer([]string{"127.0.0.1:1"})
	defer producer.Close()

	response := eventbus.NewResponse("corr-1", "get_profile", true, map[string]string{"id": "123"}, "")
	if err := producer.SendResponse(context.Background(), response); err == nil {
		t.Fatal("expected SendResponse to fail when kafka is unavailable")
	}
}

func TestSendResponseReturnsMarshalError(t *testing.T) {
	producer := NewResponseProducer([]string{"localhost:9092"})
	defer producer.Close()

	response := &eventbus.Response{
		CorrelationID: "corr-2",
		MessageType:   "broken",
		Success:       true,
		Data:          make(chan int),
	}

	if err := producer.SendResponse(context.Background(), response); err == nil {
		t.Fatal("expected SendResponse to fail for non-marshalable data")
	}
}
