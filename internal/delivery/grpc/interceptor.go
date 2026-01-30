package grpc

import (
	"context"
	"strings"

	"reforest/pkg/utils"

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

func AuthInterceptor(jwtProvider *utils.JWTProvider, publicMethods map[string]bool, adminMethods map[string]bool) googleGrpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *googleGrpc.UnaryServerInfo, handler googleGrpc.UnaryHandler) (interface{}, error) {
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

		if adminMethods[info.FullMethod] {
			if claims.Role != "ADMIN" {
				return nil, status.Error(codes.PermissionDenied, "this action requires admin privileges")
			}
		}

		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, userRoleKey, claims.Role)

		return handler(ctx, req)
	}
}