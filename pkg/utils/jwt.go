package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTProvider struct {
	secretKey string
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secretKey: secret}
}

// GenerateToken creates a new JWT for a specific user
func (j *JWTProvider) GenerateToken(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken is useful for the API Gateway or other services
func (j *JWTProvider) ValidateToken(tokenString string) (jwt.MapClaims, error) {
    // Logic to parse and validate would go here later
    return nil, nil 
}