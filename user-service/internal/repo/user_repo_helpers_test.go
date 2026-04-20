package repo

import "testing"

func TestSplitKey(t *testing.T) {
	got := splitKey("appearance.theme")
	if len(got) != 2 || got[0] != "appearance" || got[1] != "theme" {
		t.Fatalf("unexpected split result: %#v", got)
	}
}

func TestJoinKeys(t *testing.T) {
	got := joinKeys([]string{"appearance", "theme"})
	if got != "appearance.theme" {
		t.Fatalf("expected joined key %q, got %q", "appearance.theme", got)
	}
}

func TestUpdateNestedKeyCreatesPath(t *testing.T) {
	repository := &UserRepo{}
	settings := map[string]interface{}{}

	updated := repository.updateNestedKey(settings, "appearance.theme", "dark")

	appearance, ok := updated["appearance"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected nested appearance map, got %#v", updated["appearance"])
	}
	if appearance["theme"] != "dark" {
		t.Fatalf("expected theme to be updated, got %#v", appearance["theme"])
	}
}

func TestGetNestedKey(t *testing.T) {
	repository := &UserRepo{}
	settings := map[string]interface{}{
		"appearance": map[string]interface{}{
			"theme": "dark",
		},
	}

	value, err := repository.getNestedKey(settings, "appearance.theme")
	if err != nil {
		t.Fatalf("getNestedKey returned error: %v", err)
	}
	if value != "dark" {
		t.Fatalf("expected value %q, got %#v", "dark", value)
	}
}

func TestGetNestedKeyReturnsErrorForMissingKey(t *testing.T) {
	repository := &UserRepo{}

	if _, err := repository.getNestedKey(map[string]interface{}{}, "appearance.theme"); err == nil {
		t.Fatal("expected getNestedKey to return error for missing key")
	}
}
