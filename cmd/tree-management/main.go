package main

import (
	"context"
	"log"
	"net"
	"time"

	"reforest/config"
	"reforest/internal/delivery/grpc"
	"reforest/internal/repository"
	"reforest/internal/service"
	"reforest/pkg/mq"
	"reforest/pkg/pb"
	"reforest/pkg/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func metadataForwardingInterceptor(ctx context.Context, method string, req, reply interface{}, cc *googleGrpc.ClientConn, invoker googleGrpc.UnaryInvoker, opts ...googleGrpc.CallOption) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// Forward all metadata, including the "authorization" header
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoDSN))
	if err != nil {
		log.Fatalf("failed to connect to mongo: %v", err)
	}
	db := mongoClient.Database("reforest_db")

	financeConn, err := googleGrpc.NewClient("finance-service:50053",
		googleGrpc.WithTransportCredentials(insecure.NewCredentials()),
		googleGrpc.WithUnaryInterceptor(metadataForwardingInterceptor),
	)
	if err != nil {
		log.Fatalf("failed to connect to finance service: %v", err)
	}
	defer financeConn.Close()
	financeClient := pb.NewFinanceServiceClient(financeConn)

	mqClient, err := mq.NewClient(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
	defer mqClient.Close()

	jwtProvider := utils.NewJWTProvider(cfg.JWTSecret)
	treeRepo := repository.NewTreeManagementRepository(db)
	treeSvc := service.NewTreeManagementService(treeRepo, financeClient, mqClient)
	treeSvc.StartConsumers()
	treeHandler := grpc.NewTreeManagementHandler(treeSvc)

	lis, err := net.Listen("tcp", cfg.TreeGRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	publicMethods := map[string]bool{
		"/tree.TreeService/ListSpecies": true,
		"/tree.TreeService/GetSpecies":  true,
		"/tree.TreeService/ListPlots":   true,
		"/tree.TreeService/GetPlot":     true,
		"/tree.TreeService/ListTrees":   true,
		"/tree.TreeService/GetTree":     true,
		"/tree.TreeService/GetTreeLogs": true,
	}

	adminMethods := map[string]bool{
		"/tree.TreeService/CreateSpecies":              true,
		"/tree.TreeService/UpdateSpecies":              true,
		"/tree.TreeService/DeleteSpecies":              true,
		"/tree.TreeService/CreatePlot":                 true,
		"/tree.TreeService/UpdatePlot":                 true,
		"/tree.TreeService/DeletePlot":                 true,
		"/tree.TreeService/UpdateTree":                 true,
		"/tree.TreeService/DeleteTree":                 true,
		"/tree.TreeService/CreateLog":                  true,
		"/tree.TreeService/UpdateLog":                  true,
		"/tree.TreeService/DeleteLog":                  true,
		"/tree.TreeService/TriggerBiweeklyMaintenance": true,
	}

	sponsorMethods := map[string]bool{
		"/tree.TreeService/AdoptTree": true,
	}

	s := googleGrpc.NewServer(
		googleGrpc.UnaryInterceptor(grpc.AuthInterceptor(jwtProvider, publicMethods, adminMethods, sponsorMethods)),
	)
	pb.RegisterTreeServiceServer(s, treeHandler)

	log.Printf("Tree Management Service listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
