package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserCacheInterface interface {
	GetStatus(ctx context.Context, userID string) (string, error)
	SetStatus(ctx context.Context, userID, status string) error
	DeleteStatus(ctx context.Context, userID string) error
}

type UserCache struct {
	client *redis.Client
}

func NewUserCache(client *redis.Client) *UserCache {
	return &UserCache{client: client}
}

// SetStatus - устанавливает статус пользователя (online/offline/active)
func (u *UserCache) SetStatus(ctx context.Context, userID, status string) error {
	key := fmt.Sprintf("user:status:%s", userID)
	return u.client.Set(ctx, key, status, 5*time.Minute).Err()
}

// GetStatus - получает статус пользователя
func (u *UserCache) GetStatus(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("user:status:%s", userID)
	status, err := u.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return "offline", nil
	}
	if err != nil {
		return "", err
	}

	return status, nil
}

// DeleteStatus - удаляет статус (при logout)
func (u *UserCache) DeleteStatus(ctx context.Context, userID string) error {
	key := fmt.Sprintf("user:status:%s", userID)
	return u.client.Del(ctx, key).Err()
}
