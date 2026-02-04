package main

import (
	"log"
	"net"

	"reforest/config"
	"reforest/internal/delivery/grpc"
	"reforest/internal/repository"
	"reforest/internal/service"
	"reforest/pkg/database"
	"reforest/pkg/pb"
	"reforest/pkg/utils"

	googleGrpc "google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	db := database.NewConnection(cfg.DBDSN)

	jwtProvider := utils.NewJWTProvider(cfg.JWTSecret)

	authRepo := repository.NewAuthRepository(db)
	emailSender := service.NewMailtrapSender(cfg)
	authSvc := service.NewAuthService(authRepo, jwtProvider, emailSender)
	authHandler := grpc.NewAuthHandler(authSvc)

	lis, err := net.Listen("tcp", cfg.AuthGRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := googleGrpc.NewServer()
	pb.RegisterAuthServiceServer(s, authHandler)

	log.Printf("Auth Service listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
