package cache

import (
	"context"
	"testing"
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
