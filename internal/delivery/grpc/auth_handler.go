package grpc

import (
	"context"
	"reforest/internal/service"
	"reforest/pkg/pb"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error) {
	token, user, err := h.authService.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.AuthResponse{
		Token: token,
		Role:  user.RoleType,
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.authService.Register(ctx, req)
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{
		Message: "successfully registered",
		UserId:  user.ID.String(),
	}, nil
}