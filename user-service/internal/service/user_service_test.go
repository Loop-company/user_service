package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/cache"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/entity"
	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/repo"
	"github.com/stretchr/testify/assert"
)

// MockUserRepository - мок для UserRepository
type MockUserRepository struct {
	repo.UserRepository

	GetByIDFunc          func(ctx context.Context, id string) (*entity.User, error)
	UpdateNameFunc       func(ctx context.Context, id, name string) error
	GetSettingsFunc      func(ctx context.Context, userID string) (*entity.UserSettings, error)
	UpdateSettingsFunc   func(ctx context.Context, userID string, settings map[string]interface{}) error
	UpdateSettingKeyFunc func(ctx context.Context, userID, key string, value interface{}) error
	GetSettingKeyFunc    func(ctx context.Context, userID, key string) (interface{}, error)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, errors.New("GetByID not implemented in mock")
}

func (m *MockUserRepository) UpdateName(ctx context.Context, id, name string) error {
	if m.UpdateNameFunc != nil {
		return m.UpdateNameFunc(ctx, id, name)
	}
	return errors.New("UpdateName not implemented in mock")
}

func (m *MockUserRepository) GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error) {
	if m.GetSettingsFunc != nil {
		return m.GetSettingsFunc(ctx, userID)
	}
	return nil, errors.New("GetSettings not implemented in mock")
}

func (m *MockUserRepository) UpdateSettings(ctx context.Context, userID string, settings map[string]interface{}) error {
	if m.UpdateSettingsFunc != nil {
		return m.UpdateSettingsFunc(ctx, userID, settings)
	}
	return errors.New("UpdateSettings not implemented in mock")
}

func (m *MockUserRepository) UpdateSettingKey(ctx context.Context, userID, key string, value interface{}) error {
	if m.UpdateSettingKeyFunc != nil {
		return m.UpdateSettingKeyFunc(ctx, userID, key, value)
	}
	return errors.New("UpdateSettingKey not implemented in mock")
}

func (m *MockUserRepository) GetSettingKey(ctx context.Context, userID, key string) (interface{}, error) {
	if m.GetSettingKeyFunc != nil {
		return m.GetSettingKeyFunc(ctx, userID, key)
	}
	return nil, errors.New("GetSettingKey not implemented in mock")
}

// MockUserCache - мок для UserCache
type MockUserCache struct {
	cache.UserCacheInterface

	GetStatusFunc func(ctx context.Context, userID string) (string, error)
	SetStatusFunc func(ctx context.Context, userID, status string) error
}

func (m *MockUserCache) GetStatus(ctx context.Context, userID string) (string, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc(ctx, userID)
	}
	return "offline", nil
}

func (m *MockUserCache) SetStatus(ctx context.Context, userID, status string) error {
	if m.SetStatusFunc != nil {
		return m.SetStatusFunc(ctx, userID, status)
	}
	return nil
}

func TestUserService_GetProfile(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupMocks  func(*MockUserRepository, *MockUserCache)
		wantProfile *entity.UserProfile
		wantErr     bool
		errContains string
	}{
		{
			name:   "success - profile found",
			userID: "test-user-uuid",
			setupMocks: func(repo *MockUserRepository, cache *MockUserCache) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*entity.User, error) {
					return &entity.User{
						ID:            "test-user-uuid",
						Discriminator: "a1b2c3d4",
						Name:          "TestUser",
						Avatar:        "https://example.com/avatar.jpg",
					}, nil
				}
				cache.GetStatusFunc = func(ctx context.Context, userID string) (string, error) {
					return "online", nil
				}
			},
			wantProfile: &entity.UserProfile{
				ID:            "test-user-uuid",
				Discriminator: "a1b2c3d4",
				Name:          "TestUser",
				Avatar:        "https://example.com/avatar.jpg",
				Status:        "online",
			},
			wantErr: false,
		},
		{
			name:   "error - user not found in repo",
			userID: "non-existent",
			setupMocks: func(repo *MockUserRepository, cache *MockUserCache) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*entity.User, error) {
					return nil, nil
				}
			},
			wantProfile: nil,
			wantErr:     true,
			errContains: "user not found",
		},
		{
			name:   "error - repo error",
			userID: "error-user",
			setupMocks: func(repo *MockUserRepository, cache *MockUserCache) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*entity.User, error) {
					return nil, errors.New("database connection failed")
				}
			},
			wantProfile: nil,
			wantErr:     true,
			errContains: "failed to get user",
		},
		{
			name:   "success - status fallback to offline when cache error",
			userID: "user-no-status",
			setupMocks: func(repo *MockUserRepository, cache *MockUserCache) {
				repo.GetByIDFunc = func(ctx context.Context, id string) (*entity.User, error) {
					return &entity.User{
						ID:            "user-no-status",
						Discriminator: "x9y8z7w6",
						Name:          "NoStatusUser",
						Avatar:        "",
					}, nil
				}
				cache.GetStatusFunc = func(ctx context.Context, userID string) (string, error) {
					return "", errors.New("redis unavailable")
				}
			},
			wantProfile: &entity.UserProfile{
				ID:            "user-no-status",
				Discriminator: "x9y8z7w6",
				Name:          "NoStatusUser",
				Avatar:        "",
				Status:        "offline",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo, mockCache)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			profile, err := service.GetProfile(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantProfile, profile)
		})
	}
}

func TestUserService_UpdateName(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		newName     string
		setupMocks  func(*MockUserRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:    "success - valid name update",
			userID:  "user-123",
			newName: "NewDisplayName",
			setupMocks: func(repo *MockUserRepository) {
				repo.UpdateNameFunc = func(ctx context.Context, id, name string) error {
					assert.Equal(t, "user-123", id)
					assert.Equal(t, "NewDisplayName", name)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:    "error - name too short",
			userID:  "user-123",
			newName: "A", // < 2 chars
			setupMocks: func(repo *MockUserRepository) {
			},
			wantErr:     true,
			errContains: "name must be at least 2 characters",
		},
		{
			name:    "error - name too long",
			userID:  "user-123",
			newName: "ThisNameIsWayTooLongAndShouldBeRejectedByValidation", // > 32 chars
			setupMocks: func(repo *MockUserRepository) {
			},
			wantErr:     true,
			errContains: "name must be at most 32 characters",
		},
		{
			name:    "error - repo returns error",
			userID:  "user-123",
			newName: "ValidName",
			setupMocks: func(repo *MockUserRepository) {
				repo.UpdateNameFunc = func(ctx context.Context, id, name string) error {
					return errors.New("user not found in database")
				}
			},
			wantErr:     true,
			errContains: "user not found in database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			err := service.UpdateName(ctx, tt.userID, tt.newName)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_UpdateStatus(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		status      string
		setupMocks  func(*MockUserCache)
		wantErr     bool
		errContains string
	}{
		{
			name:   "success - valid status: online",
			userID: "user-456",
			status: "online",
			setupMocks: func(cache *MockUserCache) {
				cache.SetStatusFunc = func(ctx context.Context, userID, status string) error {
					assert.Equal(t, "user-456", userID)
					assert.Equal(t, "online", status)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:   "success - valid status: dnd",
			userID: "user-456",
			status: "dnd",
			setupMocks: func(cache *MockUserCache) {
				cache.SetStatusFunc = func(ctx context.Context, userID, status string) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:        "error - invalid status value",
			userID:      "user-456",
			status:      "super-online-pro-max",
			setupMocks:  func(cache *MockUserCache) {},
			wantErr:     true,
			errContains: "invalid status",
		},
		{
			name:   "error - cache returns error",
			userID: "user-456",
			status: "offline",
			setupMocks: func(cache *MockUserCache) {
				cache.SetStatusFunc = func(ctx context.Context, userID, status string) error {
					return errors.New("redis connection timeout")
				}
			},
			wantErr:     true,
			errContains: "redis connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockCache)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			err := service.UpdateStatus(ctx, tt.userID, tt.status)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_GetSettings(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		setupMocks func(*MockUserRepository)
		wantErr    bool
	}{
		{
			name:   "success - settings found",
			userID: "user-789",
			setupMocks: func(repo *MockUserRepository) {
				repo.GetSettingsFunc = func(ctx context.Context, userID string) (*entity.UserSettings, error) {
					return &entity.UserSettings{
						UserID: userID,
						Settings: map[string]interface{}{
							"appearance": map[string]interface{}{
								"theme": "dark",
							},
						},
					}, nil
				}
			},
			wantErr: false,
		},
		{
			name:   "success - no settings, returns defaults",
			userID: "new-user",
			setupMocks: func(repo *MockUserRepository) {
				repo.GetSettingsFunc = func(ctx context.Context, userID string) (*entity.UserSettings, error) {
					return &entity.UserSettings{
						UserID:   userID,
						Settings: make(map[string]interface{}),
					}, nil
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			settings, err := service.GetSettings(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, settings)
				assert.Equal(t, tt.userID, settings.UserID)
			}
		})
	}
}

func TestUserService_UpdateSettings(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		settings    map[string]interface{}
		setupMocks  func(*MockUserRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:   "success - update all settings",
			userID: "user-101",
			settings: map[string]interface{}{
				"appearance": map[string]interface{}{
					"theme": "amoled",
				},
			},
			setupMocks: func(repo *MockUserRepository) {
				repo.UpdateSettingsFunc = func(ctx context.Context, userID string, settings map[string]interface{}) error {
					assert.Equal(t, "user-101", userID)
					assert.Equal(t, "amoled", settings["appearance"].(map[string]interface{})["theme"])
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:     "error - repo fails",
			userID:   "user-101",
			settings: map[string]interface{}{"key": "value"},
			setupMocks: func(repo *MockUserRepository) {
				repo.UpdateSettingsFunc = func(ctx context.Context, userID string, settings map[string]interface{}) error {
					return errors.New("database write failed")
				}
			},
			wantErr:     true,
			errContains: "database write failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			err := service.UpdateSettings(ctx, tt.userID, tt.settings)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_UpdateSettingKey(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		key         string
		value       interface{}
		setupMocks  func(*MockUserRepository)
		wantErr     bool
		errContains string
	}{
		{
			name:   "success - update nested setting key",
			userID: "user-202",
			key:    "appearance.theme",
			value:  "light",
			setupMocks: func(repo *MockUserRepository) {
				repo.UpdateSettingKeyFunc = func(ctx context.Context, userID, key string, value interface{}) error {
					assert.Equal(t, "user-202", userID)
					assert.Equal(t, "appearance.theme", key)
					assert.Equal(t, "light", value)
					return nil
				}
			},
			wantErr: false,
		},
		{
			name:   "error - repo update setting key fails",
			userID: "user-202",
			key:    "appearance.theme",
			value:  "light",
			setupMocks: func(repo *MockUserRepository) {
				repo.UpdateSettingKeyFunc = func(ctx context.Context, userID, key string, value interface{}) error {
					return errors.New("failed to update nested key")
				}
			},
			wantErr:     true,
			errContains: "failed to update nested key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			err := service.UpdateSettingKey(ctx, tt.userID, tt.key, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_GetSettingKey(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		key         string
		setupMocks  func(*MockUserRepository)
		wantValue   interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:   "success - get nested setting key",
			userID: "user-303",
			key:    "appearance.theme",
			setupMocks: func(repo *MockUserRepository) {
				repo.GetSettingKeyFunc = func(ctx context.Context, userID, key string) (interface{}, error) {
					assert.Equal(t, "user-303", userID)
					assert.Equal(t, "appearance.theme", key)
					return "dark", nil
				}
			},
			wantValue: "dark",
			wantErr:   false,
		},
		{
			name:   "error - repo get setting key fails",
			userID: "user-303",
			key:    "appearance.theme",
			setupMocks: func(repo *MockUserRepository) {
				repo.GetSettingKeyFunc = func(ctx context.Context, userID, key string) (interface{}, error) {
					return nil, errors.New("setting key not found")
				}
			},
			wantErr:     true,
			errContains: "setting key not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockRepo)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			value, err := service.GetSettingKey(ctx, tt.userID, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantValue, value)
			}
		})
	}
}

func TestUserService_GetStatus(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		setupMocks  func(*MockUserCache)
		wantStatus  string
		wantErr     bool
		errContains string
	}{
		{
			name:   "success - get status",
			userID: "user-404",
			setupMocks: func(cache *MockUserCache) {
				cache.GetStatusFunc = func(ctx context.Context, userID string) (string, error) {
					return "online", nil
				}
			},
			wantStatus: "online",
			wantErr:    false,
		},
		{
			name:   "error - cache returns error",
			userID: "user-404",
			setupMocks: func(cache *MockUserCache) {
				cache.GetStatusFunc = func(ctx context.Context, userID string) (string, error) {
					return "", errors.New("status unavailable")
				}
			},
			wantErr:     true,
			errContains: "status unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockUserRepository{}
			mockCache := &MockUserCache{}
			if tt.setupMocks != nil {
				tt.setupMocks(mockCache)
			}

			service := NewUserService(mockRepo, mockCache)
			ctx := context.Background()

			status, err := service.GetStatus(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStatus, status)
			}
		})
	}
}
