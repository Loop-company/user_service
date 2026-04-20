package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/entity"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/repo"
	"github.com/segmentio/kafka-go"
)

type RegistrationConsumer struct {
	reader *kafka.Reader
	repo   repo.UserRepository
}

func NewRegistrationConsumer(brokers []string, groupID string, repo repo.UserRepository) *RegistrationConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   "user.events",
		GroupID: groupID,
	})

	return &RegistrationConsumer{
		reader: reader,
		repo:   repo,
	}
}

// Start запускает потребителя
func (c *RegistrationConsumer) Start(ctx context.Context) error {
	slog.Info("Starting registration consumer...")

	go func() {
		defer func() {
			if err := c.reader.Close(); err != nil {
				slog.Error("Failed to close registration reader", "error", err)
			}
		}()

		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				slog.Error("Failed to fetch message", "error", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if err := c.handleEvent(ctx, msg); err != nil {
				slog.Error("Failed to handle registration event", "error", err)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				slog.Error("Failed to commit message", "error", err)
			}
		}
	}()

	return nil
}

func (c *RegistrationConsumer) handleEvent(ctx context.Context, msg kafka.Message) error {
	var event struct {
		Type string `json:"type"`
		Data struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		} `json:"data"`
	}

	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	if event.Type != "user.registered" {
		return nil
	}

	slog.Info("Processing user registration", "user_id", event.Data.UserID)

	if existing, _ := c.repo.GetByID(ctx, event.Data.UserID); existing != nil {
		slog.Warn("User profile already exists, skipping", "user_id", event.Data.UserID)
		return nil
	}

	discriminator, err := c.generateDiscriminator(ctx)
	if err != nil {
		return fmt.Errorf("failed to generate discriminator: %w", err)
	}

	user := &entity.User{
		ID:            event.Data.UserID,
		Discriminator: discriminator,
		Name:          event.Data.Email,
		Avatar:        "",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return c.repo.Create(ctx, user)
}

func (c *RegistrationConsumer) generateDiscriminator(ctx context.Context) (string, error) {
	charset := "0123456789abcdefghijklmnopqrstuvwxyz"
	const length = 6

	for i := 0; i < 200; i++ {
		disc := make([]byte, length)
		for j := range disc {
			disc[j] = charset[rand.Intn(len(charset))]
		}
		code := string(disc)

		existing, err := c.repo.GetByDiscriminator(ctx, code)
		if err != nil {
			slog.Warn("Error checking discriminator", "disc", code, "error", err)
			continue
		}
		if existing == nil {
			return code, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique discriminator")
}
