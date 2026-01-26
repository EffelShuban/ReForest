package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthClaims defines the structure of the JWT claims.
type AuthClaims struct {
	UserID uuid.UUID
	Role   string
}

type JWTProvider struct {
	secretKey string
}

func NewJWTProvider(secret string) *JWTProvider {
	return &JWTProvider{secretKey: secret}
}

func (j *JWTProvider) GenerateToken(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTProvider) ValidateToken(tokenString string) (*AuthClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return nil, errors.New("user_id claim is missing or not a string")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user_id format")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, errors.New("role claim is missing or not a string")
		}

		return &AuthClaims{
			UserID: userID,
			Role:   role,
		}, nil
	}

	return nil, errors.New("invalid token")
}