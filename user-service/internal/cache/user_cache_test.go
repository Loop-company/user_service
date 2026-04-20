package cache

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestNewUserCache(t *testing.T) {
	cache := NewUserCache(nil)
	if cache == nil {
		t.Fatal("expected user cache to be created")
	}
}

func TestDeleteStatusPanicsWithNilClient(t *testing.T) {
	cache := NewUserCache(nil)

	defer func() {
		if recover() == nil {
			t.Fatal("expected DeleteStatus to panic with nil redis client")
		}
	}()

	_ = cache.DeleteStatus(context.Background(), "user-1")
}

func TestSetStatusReturnsErrorWhenRedisUnavailable(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  10,
		ReadTimeout:  10,
		WriteTimeout: 10,
	})
	defer client.Close()

	cache := NewUserCache(client)
	if err := cache.SetStatus(context.Background(), "user-1", "online"); err == nil {
		t.Fatal("expected SetStatus to fail when redis is unavailable")
	}
}

func TestGetStatusReturnsErrorWhenRedisUnavailable(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  10,
		ReadTimeout:  10,
		WriteTimeout: 10,
	})
	defer client.Close()

	cache := NewUserCache(client)
	if _, err := cache.GetStatus(context.Background(), "user-1"); err == nil {
		t.Fatal("expected GetStatus to fail when redis is unavailable")
	}
}
