package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/entity"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/repo"
	"github.com/segmentio/kafka-go"
)

type userRepoMock struct {
	repo.UserRepository
	createFunc             func(ctx context.Context, user *entity.User) error
	getByIDFunc            func(ctx context.Context, id string) (*entity.User, error)
	getByDiscriminatorFunc func(ctx context.Context, disc string) (*entity.User, error)
}

func (m *userRepoMock) Create(ctx context.Context, user *entity.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *userRepoMock) GetByID(ctx context.Context, id string) (*entity.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *userRepoMock) GetByDiscriminator(ctx context.Context, disc string) (*entity.User, error) {
	if m.getByDiscriminatorFunc != nil {
		return m.getByDiscriminatorFunc(ctx, disc)
	}
	return nil, nil
}

func TestHandleEventCreatesUser(t *testing.T) {
	var createdUser *entity.User
	repository := &userRepoMock{
		getByIDFunc: func(ctx context.Context, id string) (*entity.User, error) {
			return nil, nil
		},
		getByDiscriminatorFunc: func(ctx context.Context, disc string) (*entity.User, error) {
			return nil, nil
		},
		createFunc: func(ctx context.Context, user *entity.User) error {
			createdUser = user
			return nil
		},
	}
	consumer := &RegistrationConsumer{repo: repository}

	event := map[string]interface{}{
		"type": "user.registered",
		"data": map[string]string{
			"user_id": "user-1",
			"email":   "user@example.com",
		},
	}
	value, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	if err := consumer.handleEvent(context.Background(), kafka.Message{Value: value}); err != nil {
		t.Fatalf("handleEvent returned error: %v", err)
	}
	if createdUser == nil {
		t.Fatal("expected user to be created")
	}
	if createdUser.ID != "user-1" {
		t.Fatalf("expected user ID %q, got %q", "user-1", createdUser.ID)
	}
	if createdUser.Name != "user@example.com" {
		t.Fatalf("expected user name %q, got %q", "user@example.com", createdUser.Name)
	}
	if len(createdUser.Discriminator) != 6 {
		t.Fatalf("expected discriminator length 6, got %q", createdUser.Discriminator)
	}
}

func TestHandleEventIgnoresUnknownType(t *testing.T) {
	repository := &userRepoMock{
		createFunc: func(ctx context.Context, user *entity.User) error {
			t.Fatal("Create should not be called for unknown events")
			return nil
		},
	}
	consumer := &RegistrationConsumer{repo: repository}

	event := map[string]interface{}{
		"type": "user.updated",
		"data": map[string]string{
			"user_id": "user-1",
		},
	}
	value, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	if err := consumer.handleEvent(context.Background(), kafka.Message{Value: value}); err != nil {
		t.Fatalf("handleEvent returned error: %v", err)
	}
}

func TestGenerateDiscriminatorRetriesDuplicates(t *testing.T) {
	attempts := 0
	repository := &userRepoMock{
		getByDiscriminatorFunc: func(ctx context.Context, disc string) (*entity.User, error) {
			attempts++
			if attempts == 1 {
				return &entity.User{Discriminator: disc}, nil
			}
			return nil, nil
		},
	}
	consumer := &RegistrationConsumer{repo: repository}

	discriminator, err := consumer.generateDiscriminator(context.Background())
	if err != nil {
		t.Fatalf("generateDiscriminator returned error: %v", err)
	}
	if len(discriminator) != 6 {
		t.Fatalf("expected discriminator length 6, got %q", discriminator)
	}
	if attempts < 2 {
		t.Fatal("expected generator to retry after duplicate discriminator")
	}
}

func TestGenerateDiscriminatorReturnsErrorAfterExhaustion(t *testing.T) {
	repository := &userRepoMock{
		getByDiscriminatorFunc: func(ctx context.Context, disc string) (*entity.User, error) {
			return nil, errors.New("temporary error")
		},
	}
	consumer := &RegistrationConsumer{repo: repository}

	if _, err := consumer.generateDiscriminator(context.Background()); err == nil {
		t.Fatal("expected generateDiscriminator to return error after retries")
	}
}
