package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"reforest/config"
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
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": st.Message()})
	}
}

func respondProto(c *gin.Context, code int, msg proto.Message) {
	m := protojson.MarshalOptions{
		UseProtoNames:   true, // Preserves snake_case from proto definitions
		EmitUnpopulated: true, // Optional: Include fields with default values (like 0 or "")
	}
	b, err := m.Marshal(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal response"})
		return
	}
	c.Data(code, "application/json", b)
}

func main() {
	cfg := config.Load()

	conn, err := grpc.NewClient(cfg.AuthServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to auth service: %v", err)
	}
	defer conn.Close()

	authClient := pb.NewAuthServiceClient(conn)

	treeConn, err := grpc.NewClient(cfg.TreeManagementServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect to tree management service: %v", err)
	}
	defer treeConn.Close()

	treeClient := pb.NewTreeServiceClient(treeConn)

	r := gin.Default()

	authForwarder := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		ctx := metadata.NewOutgoingContext(c.Request.Context(), metadata.New(map[string]string{"authorization": authHeader}))
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

		respondProto(c, http.StatusOK, res)
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
	}

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
