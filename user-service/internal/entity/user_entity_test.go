package entity

import "testing"

func TestNewDefaultUserSettings(t *testing.T) {
	settings := NewDefaultUserSettings("user-1")

	if settings.UserID != "user-1" {
		t.Fatalf("expected user ID %q, got %q", "user-1", settings.UserID)
	}

	appearance, ok := settings.Settings["appearance"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected appearance settings map, got %#v", settings.Settings["appearance"])
	}
	if appearance["theme"] != "dark" {
		t.Fatalf("expected default theme %q, got %#v", "dark", appearance["theme"])
	}
}
