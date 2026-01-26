package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"reforest/pkg/pb"
)

func main() {
	authServiceUrl := os.Getenv("AUTH_SERVICE_URL")
	if authServiceUrl == "" {
		authServiceUrl = "localhost:50051"
	}
	conn, err := grpc.NewClient(authServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to auth service: %v", err)
	}
	defer conn.Close()

	authClient := pb.NewAuthServiceClient(conn)

	r := gin.Default()

	r.POST("/auth/register", func(c *gin.Context) {
		var body struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
			Role     string `json:"role" binding:"required"` // Expecting "ADMIN" or "SPONSOR"
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var role pb.Role
		switch strings.ToUpper(body.Role) {
		case "ADMIN":
			role = pb.Role_ADMIN
		case "SPONSOR":
			role = pb.Role_SPONSOR
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role, must be ADMIN or SPONSOR"})
			return
		}

		res, err := authClient.Register(context.Background(), &pb.RegisterRequest{
			Email:    body.Email,
			Password: body.Password,
			Role:     role,
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	// Login Endpoint
	r.POST("/auth/login", func(c *gin.Context) {
		var req pb.LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		res, err := authClient.Login(context.Background(), &req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		c.JSON(http.StatusOK, res)
	})

	log.Println("API Gateway running on :8080")
	r.Run(":8080")
}
