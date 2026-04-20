package producer

import "testing"

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
