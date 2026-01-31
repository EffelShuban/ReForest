package main

import (
	"log"
	"net"

	"reforest/config"
	"reforest/internal/delivery/grpc"
	"reforest/internal/models"
	"reforest/internal/repository"
	"reforest/internal/service"
	"reforest/pkg/database"
	"reforest/pkg/pb"
	"reforest/pkg/utils"

	googleGrpc "google.golang.org/grpc"
)

func main() {
	cfg := config.Load()

	// Using the same Postgres DB as Auth service for simplicity in this context,
	// or it could be a separate DB in a real microservices setup.
	db := database.NewConnection(cfg.DBDSN)

	// Auto-migrate finance models
	log.Println("Running finance migrations...")
	if err := db.AutoMigrate(&models.Wallet{}, &models.Transaction{}); err != nil {
		log.Fatalf("failed to migrate finance database: %v", err)
	}

	jwtProvider := utils.NewJWTProvider(cfg.JWTSecret)

	financeRepo := repository.NewFinanceRepository(db)
	financeSvc := service.NewFinanceService(financeRepo, cfg.XenditAPIKey)
	financeHandler := grpc.NewFinanceHandler(financeSvc)

	lis, err := net.Listen("tcp", cfg.FinanceGRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	publicMethods := map[string]bool{
		"/finance.FinanceService/HandleWalletWebhook": true,
		"/finance.FinanceService/GetTransactionHistory": true,
		"/finance.FinanceService/CheckPaymentExpiry":    true,
	}

	adminMethods := map[string]bool{
		"/finance.FinanceService/CreateTransaction": true,
	}

	s := googleGrpc.NewServer(
		googleGrpc.UnaryInterceptor(grpc.AuthInterceptor(jwtProvider, publicMethods, adminMethods)),
	)
	pb.RegisterFinanceServiceServer(s, financeHandler)

	log.Printf("Finance Service listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}