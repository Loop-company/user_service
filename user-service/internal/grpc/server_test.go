package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/entity"
	userpb "github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/proto"
)

type mockUserService struct {
	getProfileFunc       func(context.Context, string) (*entity.UserProfile, error)
	updateNameFunc       func(context.Context, string, string) error
	updateStatusFunc     func(context.Context, string, string) error
	getStatusFunc        func(context.Context, string) (string, error)
	getSettingsFunc      func(context.Context, string) (*entity.UserSettings, error)
	updateSettingsFunc   func(context.Context, string, map[string]interface{}) error
	updateSettingKeyFunc func(context.Context, string, string, interface{}) error
	getSettingKeyFunc    func(context.Context, string, string) (interface{}, error)
}

func (m mockUserService) GetProfile(ctx context.Context, userID string) (*entity.UserProfile, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m mockUserService) UpdateName(ctx context.Context, userID, name string) error {
	if m.updateNameFunc != nil {
		return m.updateNameFunc(ctx, userID, name)
	}
	return nil
}

func (m mockUserService) UpdateStatus(ctx context.Context, userID, status string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(ctx, userID, status)
	}
	return nil
}

func (m mockUserService) GetStatus(ctx context.Context, userID string) (string, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx, userID)
	}
	return "", nil
}

func (m mockUserService) GetSettings(ctx context.Context, userID string) (*entity.UserSettings, error) {
	if m.getSettingsFunc != nil {
		return m.getSettingsFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m mockUserService) UpdateSettings(ctx context.Context, userID string, settings map[string]interface{}) error {
	if m.updateSettingsFunc != nil {
		return m.updateSettingsFunc(ctx, userID, settings)
	}
	return nil
}

func (m mockUserService) UpdateSettingKey(ctx context.Context, userID, key string, value interface{}) error {
	if m.updateSettingKeyFunc != nil {
		return m.updateSettingKeyFunc(ctx, userID, key, value)
	}
	return nil
}

func (m mockUserService) GetSettingKey(ctx context.Context, userID, key string) (interface{}, error) {
	if m.getSettingKeyFunc != nil {
		return m.getSettingKeyFunc(ctx, userID, key)
	}
	return nil, nil
}

func TestGetProfile(t *testing.T) {
	server := NewUserServer(mockUserService{
		getProfileFunc: func(_ context.Context, userID string) (*entity.UserProfile, error) {
			if userID != "user-1" {
				t.Fatalf("userID = %q", userID)
			}
			return &entity.UserProfile{
				ID:            "user-1",
				Name:          "Kira",
				Discriminator: "abc123",
				Avatar:        "avatar.png",
				Status:        "online",
			}, nil
		},
	})

	resp, err := server.GetProfile(context.Background(), &userpb.GetProfileRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("GetProfile returned error: %v", err)
	}
	if resp.GetUserId() != "user-1" || resp.GetName() != "Kira" || resp.GetDiscriminator() != "abc123" ||
		resp.GetAvatar() != "avatar.png" || resp.GetStatus() != "online" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetProfileReturnsServiceError(t *testing.T) {
	wantErr := errors.New("profile unavailable")
	server := NewUserServer(mockUserService{
		getProfileFunc: func(context.Context, string) (*entity.UserProfile, error) {
			return nil, wantErr
		},
	})

	_, err := server.GetProfile(context.Background(), &userpb.GetProfileRequest{UserId: "user-1"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestUpdateName(t *testing.T) {
	var gotUserID, gotName string
	server := NewUserServer(mockUserService{
		updateNameFunc: func(_ context.Context, userID, name string) error {
			gotUserID = userID
			gotName = name
			return nil
		},
	})

	resp, err := server.UpdateName(context.Background(), &userpb.UpdateNameRequest{UserId: "user-1", Name: "NewName"})
	if err != nil {
		t.Fatalf("UpdateName returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if gotUserID != "user-1" || gotName != "NewName" {
		t.Fatalf("got userID=%q name=%q", gotUserID, gotName)
	}
}

func TestUpdateStatus(t *testing.T) {
	var gotUserID, gotStatus string
	server := NewUserServer(mockUserService{
		updateStatusFunc: func(_ context.Context, userID, status string) error {
			gotUserID = userID
			gotStatus = status
			return nil
		},
	})

	resp, err := server.UpdateStatus(context.Background(), &userpb.UpdateStatusRequest{UserId: "user-1", Status: "dnd"})
	if err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if gotUserID != "user-1" || gotStatus != "dnd" {
		t.Fatalf("got userID=%q status=%q", gotUserID, gotStatus)
	}
}

func TestGetSettingsConvertsValuesToStrings(t *testing.T) {
	server := NewUserServer(mockUserService{
		getSettingsFunc: func(_ context.Context, userID string) (*entity.UserSettings, error) {
			if userID != "user-1" {
				t.Fatalf("userID = %q", userID)
			}
			return &entity.UserSettings{
				UserID: "user-1",
				Settings: map[string]interface{}{
					"theme":    "dark",
					"zoom":     1.25,
					"compact":  true,
					"optional": nil,
				},
			}, nil
		},
	})

	resp, err := server.GetSettings(context.Background(), &userpb.GetSettingsRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if resp.GetSettings()["theme"] != "dark" || resp.GetSettings()["zoom"] != "1.25" ||
		resp.GetSettings()["compact"] != "true" || resp.GetSettings()["optional"] != "" {
		t.Fatalf("unexpected settings: %#v", resp.GetSettings())
	}
}

func TestGetSettingsReturnsServiceError(t *testing.T) {
	wantErr := errors.New("settings unavailable")
	server := NewUserServer(mockUserService{
		getSettingsFunc: func(context.Context, string) (*entity.UserSettings, error) {
			return nil, wantErr
		},
	})

	_, err := server.GetSettings(context.Background(), &userpb.GetSettingsRequest{UserId: "user-1"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
}

func TestUpdateSettingsConvertsRequestMap(t *testing.T) {
	var gotUserID string
	var gotSettings map[string]interface{}
	server := NewUserServer(mockUserService{
		updateSettingsFunc: func(_ context.Context, userID string, settings map[string]interface{}) error {
			gotUserID = userID
			gotSettings = settings
			return nil
		},
	})

	resp, err := server.UpdateSettings(context.Background(), &userpb.UpdateSettingsRequest{
		UserId: "user-1",
		Settings: map[string]string{
			"theme": "light",
			"lang":  "ru",
		},
	})
	if err != nil {
		t.Fatalf("UpdateSettings returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if gotUserID != "user-1" || gotSettings["theme"] != "light" || gotSettings["lang"] != "ru" {
		t.Fatalf("got userID=%q settings=%#v", gotUserID, gotSettings)
	}
}
