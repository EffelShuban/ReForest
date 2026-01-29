package main

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"reforest/config"
	"reforest/internal/delivery/grpc"
	"reforest/internal/repository"
	"reforest/internal/service"
	"reforest/pkg/pb"
	"reforest/pkg/utils"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const (
	userIDKey   contextKey = "userID"
	userRoleKey contextKey = "userRole"
)

// AuthInterceptor is a server interceptor for authentication and authorization.
func AuthInterceptor(jwtProvider *utils.JWTProvider) googleGrpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *googleGrpc.UnaryServerInfo, handler googleGrpc.UnaryHandler) (interface{}, error) {
		// List of public methods that don't require authentication.
		publicMethods := map[string]bool{
			"/tree.TreeService/ListSpecies": true,
			"/tree.TreeService/GetSpecies":  true,
			"/tree.TreeService/ListPlots":   true,
			"/tree.TreeService/GetPlot":     true,
			"/tree.TreeService/ListTrees":   true,
			"/tree.TreeService/GetTree":     true,
			"/tree.TreeService/GetTreeLogs": true,
		}

		// If the method is public, skip auth checks.
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		authHeader := values[0]
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || !strings.EqualFold(tokenParts[0], "bearer") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format, expected 'Bearer <token>'")
		}
		token := tokenParts[1]

		claims, err := jwtProvider.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		// Methods that require admin role.
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

		// If it's an admin method, check the role.
		if adminMethods[info.FullMethod] {
			if claims.Role != "ADMIN" {
				return nil, status.Error(codes.PermissionDenied, "this action requires admin privileges")
			}
		}

		// Add claims to context for use in handlers.
		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, userRoleKey, claims.Role)

		return handler(ctx, req)
	}
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

	jwtProvider := utils.NewJWTProvider(cfg.JWTSecret)
	treeRepo := repository.NewTreeManagementRepository(db)
	treeSvc := service.NewTreeManagementService(treeRepo)
	treeHandler := grpc.NewTreeManagementHandler(treeSvc)

	lis, err := net.Listen("tcp", cfg.TreeGRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := googleGrpc.NewServer(
		googleGrpc.UnaryInterceptor(AuthInterceptor(jwtProvider)),
	)
	pb.RegisterTreeServiceServer(s, treeHandler)

	log.Printf("Tree Management Service listening on %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
