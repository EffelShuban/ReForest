package main

import (
	"context"
	"log"
	"net"
	"time"

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

	db := database.NewConnection(cfg.DBDSN)

	log.Println("Running finance migrations...")
	if err := db.AutoMigrate(&models.Transaction{}, &models.Payment{}); err != nil {
		log.Fatalf("failed to migrate finance database: %v", err)
	}

	jwtProvider := utils.NewJWTProvider(cfg.JWTSecret)

	financeRepo := repository.NewFinanceRepository(db)
	financeSvc := service.NewFinanceService(financeRepo, cfg.XenditAPIKey)
	financeHandler := grpc.NewFinanceHandler(financeSvc)

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			log.Println("Running scheduled payment expiry check...")
			if err := financeSvc.CheckPaymentExpiry(context.Background()); err != nil {
				log.Printf("Error running payment expiry check: %v", err)
			}
		}
	}()

	lis, err := net.Listen("tcp", cfg.FinanceGRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	publicMethods := map[string]bool{
		"/finance.FinanceService/HandleWalletWebhook": true,
		"/finance.FinanceService/CheckPaymentExpiry":    true,
	}

	adminMethods := map[string]bool{
		"/finance.FinanceService/CreateTransaction": true,
	}

	s := googleGrpc.NewServer(
		googleGrpc.UnaryInterceptor(grpc.AuthInterceptor(jwtProvider, publicMethods, adminMethods, nil)),
	)
	pb.RegisterFinanceServiceServer(s, financeHandler)

	log.Printf("Finance Service listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}