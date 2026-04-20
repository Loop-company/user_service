package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/entity"
)

// UserRepository - интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByDiscriminator(ctx context.Context, disc string) (*entity.User, error)
	UpdateName(ctx context.Context, id, name string) error

	GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error)
	UpdateSettings(ctx context.Context, userID string, settings map[string]interface{}) error
	UpdateSettingKey(ctx context.Context, userID, key string, value interface{}) error
	GetSettingKey(ctx context.Context, userID, key string) (interface{}, error)
}

// UserRepo - реализация репозитория
type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (repo *UserRepo) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, discriminator, name, avatar, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := repo.DB.ExecContext(ctx, query,
		user.ID,
		user.Discriminator,
		user.Name,
		user.Avatar,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		slog.Error("Failed to create user",
			slog.String("user_id", user.ID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	slog.Info("User created successfully",
		slog.String("user_id", user.ID),
		slog.String("discriminator", user.Discriminator),
	)
	return nil
}

func (repo *UserRepo) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, discriminator, name, avatar, created_at, updated_at 
	          FROM users WHERE id = $1`

	err := repo.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Discriminator,
		&user.Name,
		&user.Avatar,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("Failed to get user",
			slog.String("user_id", id),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (repo *UserRepo) GetByDiscriminator(ctx context.Context, disc string) (*entity.User, error) {
	var user entity.User
	query := `SELECT id, discriminator, name, avatar, created_at, updated_at 
	          FROM users WHERE discriminator = $1`

	err := repo.DB.QueryRowContext(ctx, query, disc).Scan(
		&user.ID,
		&user.Discriminator,
		&user.Name,
		&user.Avatar,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by discriminator: %w", err)
	}

	return &user, nil
}

func (repo *UserRepo) UpdateName(ctx context.Context, id, name string) error {
	query := `UPDATE users SET name = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

	result, err := repo.DB.ExecContext(ctx, query, name, id)
	if err != nil {
		return fmt.Errorf("failed to update name: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	slog.Info("User name updated",
		slog.String("user_id", id),
		slog.String("new_name", name),
	)
	return nil
}

// GetSettings - получает настройки пользователя
func (repo *UserRepo) GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error) {
	var settings entity.UserSettings

	query := `SELECT user_id, settings, created_at, updated_at 
	          FROM user_settings 
	          WHERE user_id = $1`

	err := repo.DB.QueryRowContext(ctx, query, userID).Scan(
		&settings.UserID,
		&settings.Settings,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Создаём дефолтные настройки
			slog.Info("Settings not found, creating defaults",
				slog.String("user_id", userID),
			)

			defaultSettings := entity.NewDefaultUserSettings(userID)
			if err := repo.UpdateSettings(ctx, userID, defaultSettings.Settings); err != nil {
				return nil, fmt.Errorf("failed to create default settings: %w", err)
			}

			return defaultSettings, nil
		}

		slog.Error("Failed to get user settings",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return &settings, nil
}

// UpdateSettings - полностью заменяет настройки пользователя (upsert)
func (repo *UserRepo) UpdateSettings(ctx context.Context, userID string, settings map[string]interface{}) error {
	user, err := repo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return errors.New("user not found")
	}

	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	query := `
		INSERT INTO user_settings (user_id, settings, created_at, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			settings = EXCLUDED.settings,
			updated_at = CURRENT_TIMESTAMP
	`

	result, err := repo.DB.ExecContext(ctx, query, userID, settingsJSON)
	if err != nil {
		slog.Error("Failed to update user settings",
			slog.String("user_id", userID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to update settings: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	slog.Info("User settings updated",
		slog.String("user_id", userID),
		slog.Int64("rows_affected", rowsAffected),
	)

	return nil
}

// UpdateSettingKey - обновляет конкретный ключ в настройках (частичное обновление)
func (repo *UserRepo) UpdateSettingKey(ctx context.Context, userID, key string, value interface{}) error {
	currentSettings, err := repo.GetSettings(ctx, userID)
	if err != nil {
		return err
	}

	updatedSettings := repo.updateNestedKey(currentSettings.Settings, key, value)

	return repo.UpdateSettings(ctx, userID, updatedSettings)
}

// updateNestedKey - рекурсивно обновляет вложенный ключ
func (repo *UserRepo) updateNestedKey(settings map[string]interface{}, key string, value interface{}) map[string]interface{} {
	keys := splitKey(key)

	if len(keys) == 1 {
		settings[keys[0]] = value
		return settings
	}

	currentKey := keys[0]
	remainingKeys := keys[1:]

	nested, ok := settings[currentKey].(map[string]interface{})
	if !ok {
		nested = make(map[string]interface{})
		settings[currentKey] = nested
	}

	settings[currentKey] = repo.updateNestedKey(nested, joinKeys(remainingKeys), value)

	return settings
}

// splitKey - разбивает "appearance.theme" на ["appearance", "theme"]
func splitKey(key string) []string {
	result := make([]string, 0)
	current := ""

	for _, ch := range key {
		if ch == '.' {
			result = append(result, current)
			current = ""
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// joinKeys - соединяет ["theme", "dark"] в "theme.dark"
func joinKeys(keys []string) string {
	result := ""
	for i, key := range keys {
		if i > 0 {
			result += "."
		}
		result += key
	}
	return result
}

// GetSettingKey - получает значение конкретного ключа
func (repo *UserRepo) GetSettingKey(ctx context.Context, userID, key string) (interface{}, error) {
	settings, err := repo.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}

	return repo.getNestedKey(settings.Settings, key)
}

// getNestedKey - получает значение вложенного ключа
func (repo *UserRepo) getNestedKey(settings map[string]interface{}, key string) (interface{}, error) {
	keys := splitKey(key)
	current := interface{}(settings)

	for _, k := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("key '%s' is not a map", k)
		}

		current, ok = m[k]
		if !ok {
			return nil, fmt.Errorf("key '%s' not found", k)
		}
	}

	return current, nil
}
