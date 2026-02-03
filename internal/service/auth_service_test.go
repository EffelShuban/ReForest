package service

import (
	"context"
	"testing"

	"reforest/internal/models"
	"reforest/pkg/pb"
	"reforest/pkg/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
	createdUser *models.User
	createErr   error
	getUser     *models.User
	getErr      error
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	m.createdUser = user
	return user, m.createErr
}

func (m *mockAuthRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.getUser, m.getErr
}

func TestAuthService_Register_Success(t *testing.T) {
	repo := &mockAuthRepo{}
	jwtProvider := utils.NewJWTProvider("secret")
	svc := NewAuthService(repo, jwtProvider)

	req := &pb.RegisterRequest{
		Email:       "user@example.com",
		Password:    "Password1",
		RoleType:    "ADMIN",
		FullName:    "User Test",
		DateOfBirth: "2000-01-01",
	}

	user, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if repo.createdUser == nil {
		t.Fatalf("CreateUser was not called")
	}

	if repo.createdUser.Email != req.Email || repo.createdUser.RoleType != req.RoleType {
		t.Fatalf("User fields not set correctly")
	}

	if repo.createdUser.PasswordHash == req.Password {
		t.Fatalf("Password was not hashed")
	}

	if user.Profile.ID != user.ID {
		t.Fatalf("Profile ID should match User ID")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	repo := &mockAuthRepo{}
	jwtProvider := utils.NewJWTProvider("secret")
	svc := NewAuthService(repo, jwtProvider)

	pw := "secret123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	userID := uuid.New()

	repo.getUser = &models.User{
		ID:           userID,
		Email:        "user@example.com",
		PasswordHash: string(hashed),
		RoleType:     "SPONSOR",
	}

	token, user, err := svc.Login(context.Background(), &pb.LoginRequest{
		Email:    "user@example.com",
		Password: pw,
	})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if token == "" {
		t.Fatalf("expected token to be returned")
	}
	if user.ID != userID {
		t.Fatalf("returned user mismatch")
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	repo := &mockAuthRepo{}
	jwtProvider := utils.NewJWTProvider("secret")
	svc := NewAuthService(repo, jwtProvider)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo.getUser = &models.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: string(hashed),
		RoleType:     "SPONSOR",
	}

	_, _, err := svc.Login(context.Background(), &pb.LoginRequest{
		Email:    "user@example.com",
		Password: "wrong",
	})
	if err == nil || err != models.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := &mockAuthRepo{getErr: models.ErrInvalidCredentials}
	jwtProvider := utils.NewJWTProvider("secret")
	svc := NewAuthService(repo, jwtProvider)

	_, _, err := svc.Login(context.Background(), &pb.LoginRequest{
		Email:    "user@example.com",
		Password: "any",
	})
	if err == nil || err != models.ErrInvalidCredentials {
		t.Fatalf("expected invalid credentials error, got %v", err)
	}
}
