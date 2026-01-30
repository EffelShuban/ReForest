package service

import (
	"context"
	"reforest/internal/models"
	"reforest/internal/repository"
	"reforest/pkg/pb"
	"reforest/pkg/utils"
	"time"

	"github.com/google/uuid"
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, models.ErrInternal
	}

	dob, _ := time.Parse("2006-01-02", req.DateOfBirth)
	var age int
	if !dob.IsZero() {
		age = int(time.Since(dob).Hours() / 24 / 365)
	}

	newID := uuid.New()
	user := &models.User{
		ID:           newID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		RoleType:     req.RoleType,
		Profile: models.Profile{
			ID:          newID,
			FullName:    req.FullName,
			DateOfBirth: dob,
			Age:         age,
			Balance:     0,
		},
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *authService) Login(ctx context.Context, req *pb.LoginRequest) (string, *models.User, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", nil, models.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", nil, models.ErrInvalidCredentials
	}

	token, err := s.jwtProvider.GenerateToken(user.ID, user.RoleType)
	if err != nil {
		return "", nil, models.ErrInternal
	}
	return token, user, nil
}