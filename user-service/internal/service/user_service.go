package service

import (
	"context"
	"fmt"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/cache"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/entity"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/repo"
)

// UserService - интерфейс сервиса
type UserService interface {
	GetProfile(ctx context.Context, userID string) (*entity.UserProfile, error)
	UpdateName(ctx context.Context, userID, name string) error

	UpdateStatus(ctx context.Context, userID, status string) error
	GetStatus(ctx context.Context, userID string) (string, error)

	GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error)
	UpdateSettings(ctx context.Context, userID string, settings map[string]interface{}) error
	UpdateSettingKey(ctx context.Context, userID, key string, value interface{}) error
	GetSettingKey(ctx context.Context, userID, key string) (interface{}, error)
}

type EventBus interface {
	SendUserProfileUpdated(ctx context.Context, userID string, changes map[string]interface{})
	SendUserSettingsUpdated(ctx context.Context, userID string, settings map[string]interface{})
}

type UserServiceImpl struct {
	repo     repo.UserRepository
	cache    cache.UserCacheInterface
	eventBus EventBus
}

func NewUserService(repo repo.UserRepository, cache cache.UserCacheInterface, eventBus EventBus) *UserServiceImpl {
	return &UserServiceImpl{
		repo:     repo,
		cache:    cache,
		eventBus: eventBus,
	}
}

func (s *UserServiceImpl) GetProfile(ctx context.Context, userID string) (*entity.UserProfile, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	status, err := s.cache.GetStatus(ctx, userID)
	if err != nil {
		status = "offline"
	}

	return &entity.UserProfile{
		ID:            user.ID,
		Discriminator: user.Discriminator,
		Name:          user.Name,
		Avatar:        user.Avatar,
		Status:        status,
	}, nil
}

// UpdateName - обновляет имя пользователя
func (s *UserServiceImpl) UpdateName(ctx context.Context, userID, name string) error {
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if len(name) > 32 {
		return fmt.Errorf("name must be at most 32 characters")
	}

	err := s.repo.UpdateName(ctx, userID, name)
	if err == nil {
		s.eventBus.SendUserProfileUpdated(ctx, userID, map[string]interface{}{"name": name})
	}
	return err
}

// UpdateStatus - обновляет статус пользователя (online/offline/active)
func (s *UserServiceImpl) UpdateStatus(ctx context.Context, userID, status string) error {
	validStatuses := map[string]bool{
		"online":    true,
		"offline":   true,
		"idle":      true,
		"dnd":       true,
		"invisible": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	err := s.cache.SetStatus(ctx, userID, status)
	if err == nil {
		s.eventBus.SendUserProfileUpdated(ctx, userID, map[string]interface{}{"status": status})
	}
	return err
}

// GetStatus - получает статус пользователя
func (s *UserServiceImpl) GetStatus(ctx context.Context, userID string) (string, error) {
	return s.cache.GetStatus(ctx, userID)
}

// GetSettings - получает все настройки пользователя
func (s *UserServiceImpl) GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error) {
	return s.repo.GetSettings(ctx, userID)
}

// UpdateSettings - обновляет все настройки
func (s *UserServiceImpl) UpdateSettings(ctx context.Context, userID string, settings map[string]interface{}) error {
	err := s.repo.UpdateSettings(ctx, userID, settings)
	if err == nil {
		s.eventBus.SendUserSettingsUpdated(ctx, userID, settings)
	}
	return err
}

// UpdateSettingKey - обновляет конкретный ключ в настройках
func (s *UserServiceImpl) UpdateSettingKey(ctx context.Context, userID, key string, value interface{}) error {
	err := s.repo.UpdateSettingKey(ctx, userID, key, value)
	if err == nil {
		s.eventBus.SendUserSettingsUpdated(ctx, userID, map[string]interface{}{key: value})
	}
	return err
}

// GetSettingKey - получает значение конкретного ключа
func (s *UserServiceImpl) GetSettingKey(ctx context.Context, userID, key string) (interface{}, error) {
	return s.repo.GetSettingKey(ctx, userID, key)
}
