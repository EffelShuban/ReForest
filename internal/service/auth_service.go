package service

import (
	"context"
	"errors"
	"reforest/internal/models"
	"reforest/internal/repository"
	"reforest/pkg/pb"
	"reforest/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface{
	Register(ctx context.Context, req *pb.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req *pb.LoginRequest) (string, *models.User, error)
}

type authService struct{
	repo repository.AuthRepository
	jwtProvider *utils.JWTProvider
}

func NewAuthService(repo repository.AuthRepository, jwt *utils.JWTProvider) AuthService {
	return &authService{
		repo:        repo,
		jwtProvider: jwt,
	}
}

func (s *authService) Register(ctx context.Context, req *pb.RegisterRequest) (*models.User, error) {
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

    user := &models.User{
        Email:        req.Email,
        PasswordHash: string(hashedPassword),
        RoleType:     req.Role.String(), // "ADMIN" or "SPONSOR"
    }

    return s.repo.CreateUserWithRole(ctx, user)
}

func (s *authService) Login(ctx context.Context, req *pb.LoginRequest) (string, *models.User, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	// Use the utility!
	token, err := s.jwtProvider.GenerateToken(user.ID, user.RoleType)
	return token, user, err
}