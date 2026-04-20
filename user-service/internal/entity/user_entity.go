package entity

import (
	"time"
)

// User - основная сущность пользователя (PostgreSQL)
type User struct {
	ID            string    `json:"id" db:"id"`
	Discriminator string    `json:"discriminator" db:"discriminator"`
	Name          string    `json:"name" db:"name"`
	Avatar        string    `json:"avatar" db:"avatar_url"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UserProfile - публичный профиль для отображения
type UserProfile struct {
	ID            string `json:"id"`
	Discriminator string `json:"discriminator"`
	Name          string `json:"name"`
	Avatar        string `json:"avatar"`
	Status        string `json:"status"`
}

// UserSettings - настройки пользователя (хранятся в отдельной таблице)
type UserSettings struct {
	UserID    string                 `json:"user_id" db:"user_id"`
	Settings  map[string]interface{} `json:"settings" db:"settings"` // JSONB
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// SettingsPreset - предопределённые структуры для валидации
type SettingsPreset struct {
	Appearance AppearanceSettings `json:"appearance,omitempty"`
	Audio      AudioSettings      `json:"audio,omitempty"`
	Privacy    PrivacySettings    `json:"privacy,omitempty"`
	Keybinds   KeybindsSettings   `json:"keybinds,omitempty"`
}

// AppearanceSettings - настройки внешнего вида
type AppearanceSettings struct {
	Theme           string  `json:"theme,omitempty"`             // dark/light/amoled
	Language        string  `json:"language,omitempty"`          // en/ru/etc
	Zoom            float64 `json:"zoom,omitempty"`              // 0.5 - 2.0
	MessageFontSize int     `json:"message_font_size,omitempty"` // 12-24
	CompactMode     bool    `json:"compact_mode,omitempty"`
	ShowAnimations  bool    `json:"show_animations,omitempty"`
}

// AudioSettings - настройки аудио
type AudioSettings struct {
	InputVolume      int    `json:"input_volume,omitempty"`  // 0-100
	OutputVolume     int    `json:"output_volume,omitempty"` // 0-100
	InputDevice      string `json:"input_device,omitempty"`  // ID устройства
	OutputDevice     string `json:"output_device,omitempty"` // ID устройства
	PushToTalk       bool   `json:"push_to_talk,omitempty"`
	PushToTalkKey    string `json:"push_to_talk_key,omitempty"` // например "Ctrl+Space"
	NoiseSuppression bool   `json:"noise_suppression,omitempty"`
	EchoCancellation bool   `json:"echo_cancellation,omitempty"`
}

// PrivacySettings - настройки приватности
type PrivacySettings struct {
	ShowStatus          bool   `json:"show_status,omitempty"` // показывать онлайн статус
	AllowDM             string `json:"allow_dm,omitempty"`    // all/friends/none
	ShowGameActivity    bool   `json:"show_game_activity,omitempty"`
	AllowFriendRequests bool   `json:"allow_friend_requests,omitempty"`
}

// KeybindsSettings - горячие клавиши
type KeybindsSettings struct {
	ToggleMute         string `json:"toggle_mute,omitempty"`   // например "Ctrl+M"
	ToggleDeafen       string `json:"toggle_deafen,omitempty"` // например "Ctrl+D"
	TogglePtt          string `json:"toggle_ptt,omitempty"`    // например "Ctrl+Space"
	ActivatePushToTalk string `json:"activate_push_to_talk,omitempty"`
}

// NewDefaultUserSettings - создаёт настройки по умолчанию
func NewDefaultUserSettings(userID string) *UserSettings {
	defaultSettings := map[string]interface{}{
		"appearance": map[string]interface{}{
			"theme":           "dark",
			"language":        "en",
			"zoom":            1.0,
			"show_animations": true,
		},
		"audio": map[string]interface{}{
			"input_volume":      80,
			"output_volume":     100,
			"push_to_talk":      false,
			"noise_suppression": true,
		},
		"privacy": map[string]interface{}{
			"show_status":           true,
			"allow_dm":              "friends",
			"show_game_activity":    true,
			"allow_friend_requests": true,
		},
		"keybinds": map[string]interface{}{
			"toggle_mute":   "Ctrl+M",
			"toggle_deafen": "Ctrl+D",
		},
	}

	return &UserSettings{
		UserID:    userID,
		Settings:  defaultSettings,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
