package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

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

	treeMgmtServiceURL := os.Getenv("TREE_MANAGEMENT_SERVICE_URL")
	if treeMgmtServiceURL == "" {
		treeMgmtServiceURL = "localhost:50052"
	}
	treeConn, err := grpc.NewClient(treeMgmtServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to tree management service: %v", err)
	}
	defer treeConn.Close()

	treeClient := pb.NewTreeServiceClient(treeConn)

	r := gin.Default()

	// Middleware to pass JWT from HTTP header to gRPC metadata
	authForwarder := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// Pass on the header to the downstream gRPC service.
		ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.New(map[string]string{"authorization": authHeader}))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}

	r.POST("/auth/register", func(c *gin.Context) {
		var body struct {
			Email       string `json:"email" binding:"required,email"`
			Password    string `json:"password" binding:"required,min=6"`
			Role        string `json:"role" binding:"required"` // Expecting "ADMIN" or "SPONSOR"
			FullName    string `json:"full_name"`
			DateOfBirth string `json:"date_of_birth"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate role is one of the allowed types from ERD
		switch strings.ToUpper(body.Role) {
		case "ADMIN", "SPONSOR":
			// valid
		default: 
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role, must be ADMIN or SPONSOR"})
			return
		}

		res, err := authClient.Register(context.Background(), &pb.RegisterRequest{
			Email:       body.Email,
			Password:    body.Password,
			RoleType:    strings.ToUpper(body.Role),
			FullName:    body.FullName,
			DateOfBirth: body.DateOfBirth,
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

	// --- Tree Management Routes ---
	// Public routes
	r.GET("/species", func(c *gin.Context) {
		res, err := treeClient.ListSpecies(c.Request.Context(), &emptypb.Empty{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	r.GET("/species/:id", func(c *gin.Context) {
		id := c.Param("id")
		res, err := treeClient.GetSpecies(c.Request.Context(), &pb.IdRequest{Id: id})
		if err != nil {
			st, _ := status.FromError(err)
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
			}
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// ... other public routes for plots, trees etc. can be added here

	// Admin routes
	adminRoutes := r.Group("/admin", authForwarder)
	{
		// Species
		adminRoutes.POST("/species", func(c *gin.Context) {
			var req pb.Species
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := treeClient.CreateSpecies(c.Request.Context(), &req)
			if err != nil {
				st, _ := status.FromError(err)
				if st.Code() == codes.PermissionDenied {
					c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusCreated, res)
		})

		adminRoutes.PUT("/species/:id", func(c *gin.Context) {
			var req pb.Species
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.Id = c.Param("id")
			res, err := treeClient.UpdateSpecies(c.Request.Context(), &req)
			if err != nil {
				st, _ := status.FromError(err)
				// Handle different errors like not found, invalid argument etc.
				c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
				return
			}
			c.JSON(http.StatusOK, res)
		})

		adminRoutes.DELETE("/species/:id", func(c *gin.Context) {
			id := c.Param("id")
			_, err := treeClient.DeleteSpecies(c.Request.Context(), &pb.IdRequest{Id: id})
			if err != nil {
				st, _ := status.FromError(err)
				// Handle different errors
				c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
				return
			}
			c.Status(http.StatusNoContent)
		})

		// ... other admin routes for plots, trees etc.
	}

	log.Println("API Gateway running on :8080")
	r.Run(":8080")
}
