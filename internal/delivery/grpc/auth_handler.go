package grpc

import (
	"context"
	"errors"
	"reforest/internal/models"
	"reforest/internal/service"
	"reforest/pkg/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		if errors.Is(err, models.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		return nil, status.Error(codes.Internal, "login failed")
	}

	return &pb.AuthResponse{
		Token: token,
		Role:  user.RoleType,
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	user, err := h.authService.Register(ctx, req)
	if err != nil {
		if errors.Is(err, models.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, "registration failed")
	}

	return &pb.RegisterResponse{
		Message: "successfully registered",
		UserId:  user.ID.String(),
	}, nil
}