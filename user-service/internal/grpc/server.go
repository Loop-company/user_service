package grpc

import (
	"context"
	"fmt"

	"github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/internal/service"
	userpb "github.com/Egor4iksls4/DiscordEquivalent/backend/user-service/proto"
)

type UserServer struct {
	userpb.UnimplementedUserServiceServer
	service service.UserService
}

func NewUserServer(s service.UserService) *UserServer {
	return &UserServer{service: s}
}

func (s *UserServer) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
	profile, err := s.service.GetProfile(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &userpb.GetProfileResponse{
		UserId:        profile.ID,
		Name:          profile.Name,
		Discriminator: profile.Discriminator,
		Avatar:        profile.Avatar,
		Status:        profile.Status,
	}, nil
}

func (s *UserServer) UpdateName(ctx context.Context, req *userpb.UpdateNameRequest) (*userpb.UserEmpty, error) {
	err := s.service.UpdateName(ctx, req.UserId, req.Name)
	return &userpb.UserEmpty{}, err
}

func (s *UserServer) UpdateStatus(ctx context.Context, req *userpb.UpdateStatusRequest) (*userpb.UserEmpty, error) {
	err := s.service.UpdateStatus(ctx, req.UserId, req.Status)
	return &userpb.UserEmpty{}, err
}

func (s *UserServer) GetSettings(ctx context.Context, req *userpb.GetSettingsRequest) (*userpb.GetSettingsResponse, error) {
	settings, err := s.service.GetSettings(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for k, v := range settings.Settings {
		result[k] = toString(v)
	}

	return &userpb.GetSettingsResponse{
		Settings: result,
	}, nil
}

func (s *UserServer) UpdateSettings(ctx context.Context, req *userpb.UpdateSettingsRequest) (*userpb.UserEmpty, error) {
	data := make(map[string]interface{})
	for k, v := range req.Settings {
		data[k] = v
	}

	err := s.service.UpdateSettings(ctx, req.UserId, data)
	return &userpb.UserEmpty{}, err
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
