package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"reforest/pkg/pb"
)

func handleGrpcError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "an unknown error occurred"})
		return
	}
	switch st.Code() {
	case codes.NotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": st.Message()})
	case codes.InvalidArgument:
		c.JSON(http.StatusBadRequest, gin.H{"error": st.Message()})
	case codes.PermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": st.Message()})
	case codes.Unauthenticated:
		c.JSON(http.StatusUnauthorized, gin.H{"error": st.Message()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}

func respondProto(c *gin.Context, code int, msg proto.Message) {
	m := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}
	b, err := m.Marshal(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}
	c.Data(code, "application/json", b)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment variables")
	}

	authServiceUrl := os.Getenv("AUTH_SERVICE_URL")
	conn, err := grpc.NewClient(authServiceUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to auth service: %v", err)
	}
	defer conn.Close()
	authClient := pb.NewAuthServiceClient(conn)

	treeMgmtServiceURL := os.Getenv("TREE_MANAGEMENT_SERVICE_URL")
	treeConn, err := grpc.NewClient(treeMgmtServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to tree management service: %v", err)
	}
	defer treeConn.Close()
	treeClient := pb.NewTreeServiceClient(treeConn)

	financeServiceURL := os.Getenv("FINANCE_SERVICE_URL")
	financeConn, err := grpc.NewClient(financeServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to finance service: %v", err)
	}
	defer financeConn.Close()
	financeClient := pb.NewFinanceServiceClient(financeConn)

	r := gin.Default()

	authForwarder := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		mdMap := map[string]string{"authorization": authHeader}

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenParts := strings.Split(strings.TrimPrefix(authHeader, "Bearer "), ".")
			if len(tokenParts) > 1 {
				payload, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
				if err == nil {
					var claims struct {
						UserID string `json:"user_id"`
					}
					if err := json.Unmarshal(payload, &claims); err == nil && claims.UserID != "" {
						mdMap["x-user-id"] = claims.UserID
					}
				}
			}
		}

		ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.New(mdMap))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}

	r.POST("/auth/register", func(c *gin.Context) {
		var body struct {
			Email       string `json:"email" binding:"required,email"`
			Password    string `json:"password" binding:"required,min=6"`
			Role        string `json:"role" binding:"required"` // "ADMIN" or "SPONSOR"
			FullName    string `json:"full_name"`
			DateOfBirth string `json:"date_of_birth"`
		}

		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

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

		respondProto(c, http.StatusCreated, res)
	})

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

		respondProto(c, http.StatusOK, res)
	})
	r.GET("/species", func(c *gin.Context) {
		res, err := treeClient.ListSpecies(c.Request.Context(), &emptypb.Empty{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		respondProto(c, http.StatusOK, res)
	})

	r.GET("/species/:id", func(c *gin.Context) {
		id := c.Param("id")
		res, err := treeClient.GetSpecies(c.Request.Context(), &pb.IdRequest{Id: id})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		respondProto(c, http.StatusOK, res)
	})

	r.GET("/plots", func(c *gin.Context) {
		res, err := treeClient.ListPlots(c.Request.Context(), &emptypb.Empty{})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		respondProto(c, http.StatusOK, res)
	})

	r.GET("/plots/:id", func(c *gin.Context) {
		id := c.Param("id")
		res, err := treeClient.GetPlot(c.Request.Context(), &pb.IdRequest{Id: id})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		respondProto(c, http.StatusOK, res)
	})

	r.GET("/trees", func(c *gin.Context) {
		res, err := treeClient.ListTrees(c.Request.Context(), &emptypb.Empty{})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		respondProto(c, http.StatusOK, res)
	})

	r.GET("/trees/:id", func(c *gin.Context) {
		id := c.Param("id")
		res, err := treeClient.GetTree(c.Request.Context(), &pb.IdRequest{Id: id})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		respondProto(c, http.StatusOK, res)
	})

	r.GET("/trees/:id/logs", func(c *gin.Context) {
		id := c.Param("id")
		res, err := treeClient.GetTreeLogs(c.Request.Context(), &pb.IdRequest{Id: id})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		respondProto(c, http.StatusOK, res)
	})

	authRoutes := r.Group("/", authForwarder)
	{
		authRoutes.POST("/trees", func(c *gin.Context) {
			var req pb.AdoptTreeRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := treeClient.AdoptTree(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusCreated, res)
		})

		authRoutes.GET("/wallet/balance", func(c *gin.Context) {
			res, err := financeClient.GetBalance(c.Request.Context(), &emptypb.Empty{})
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusOK, res)
		})

		authRoutes.POST("/wallet/topup", func(c *gin.Context) {
			var req pb.TopUpRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := financeClient.TopUpWallet(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusCreated, res)
		})

		authRoutes.GET("/wallet/transactions", func(c *gin.Context) {
			res, err := financeClient.GetTransactionHistory(c.Request.Context(), &emptypb.Empty{})
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusOK, res)
		})
	}

	r.POST("/webhooks/finance", func(c *gin.Context) {
		data, err := c.GetRawData()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
			return
		}

		_, err = financeClient.HandleWalletWebhook(c.Request.Context(), &pb.WebhookRequest{
			Event: "INVOICE_CALLBACK",
			Data:  data,
		})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "received"})
	})

	r.POST("/jobs/payment-expiry-check", func(c *gin.Context) {
		_, err := financeClient.CheckPaymentExpiry(c.Request.Context(), &emptypb.Empty{})
		if err != nil {
			handleGrpcError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "expiry check completed"})
	})

	adminRoutes := r.Group("/admin", authForwarder)
	{
		adminRoutes.POST("/species", func(c *gin.Context) {
			var req pb.Species
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := treeClient.CreateSpecies(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusCreated, res)
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
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusOK, res)
		})

		adminRoutes.DELETE("/species/:id", func(c *gin.Context) {
			id := c.Param("id")
			_, err := treeClient.DeleteSpecies(c.Request.Context(), &pb.IdRequest{Id: id})
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			c.Status(http.StatusNoContent)
		})

		adminRoutes.POST("/plots", func(c *gin.Context) {
			var req pb.Plot
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			res, err := treeClient.CreatePlot(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusCreated, res)
		})

		adminRoutes.PUT("/plots/:id", func(c *gin.Context) {
			var req pb.Plot
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.Id = c.Param("id")
			res, err := treeClient.UpdatePlot(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusOK, res)
		})

		adminRoutes.DELETE("/plots/:id", func(c *gin.Context) {
			id := c.Param("id")
			_, err := treeClient.DeletePlot(c.Request.Context(), &pb.IdRequest{Id: id})
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			c.Status(http.StatusNoContent)
		})

		adminRoutes.PUT("/trees/:id", func(c *gin.Context) {
			var req pb.Tree
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.Id = c.Param("id")
			res, err := treeClient.UpdateTree(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusOK, res)
		})

		adminRoutes.DELETE("/trees/:id", func(c *gin.Context) {
			id := c.Param("id")
			_, err := treeClient.DeleteTree(c.Request.Context(), &pb.IdRequest{Id: id})
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			c.Status(http.StatusNoContent)
		})

		adminRoutes.POST("/trees/:id/logs", func(c *gin.Context) {
			var req pb.CreateLogRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.AdoptedTreeId = c.Param("id")
			res, err := treeClient.CreateLog(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusCreated, res)
		})

		adminRoutes.PUT("/logs/:id", func(c *gin.Context) {
			var req pb.LogEntry
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.Id = c.Param("id")
			res, err := treeClient.UpdateLog(c.Request.Context(), &req)
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			respondProto(c, http.StatusOK, res)
		})

		adminRoutes.DELETE("/logs/:id", func(c *gin.Context) {
			id := c.Param("id")
			_, err := treeClient.DeleteLog(c.Request.Context(), &pb.IdRequest{Id: id})
			if err != nil {
				handleGrpcError(c, err)
				return
			}
			c.Status(http.StatusNoContent)
		})
	}

	log.Println("API Gateway running on :8080")
	r.Run(":8080")
}
