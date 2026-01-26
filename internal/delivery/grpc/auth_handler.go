package grpc

import (
	"context"
	"reforest/pkg/pb"
)

type AuthHandler struct{
	pb.UnimplementedAuthServiceServer
}

func NewAuthHandler() *AuthHandler{
	return &AuthHandler{}
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.AuthResponse, error){
	return &pb.AuthResponse{
		Token: "test",
		Role: "role",
	}, nil
}

func (h *AuthHandler) Register(ctx context.Context, req *pb.LoginRequest) (*pb.RegisterResponse, error){
	return &pb.RegisterResponse{
		Message: "succesfully registered",
		UserId: "001",
	}, nil
}